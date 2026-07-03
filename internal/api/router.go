package api

import (
	"context"
	"crypto/ecdsa"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"platform/internal/api/controllers"
	"platform/internal/auth"
	"platform/internal/billing"
	"platform/internal/enrollment"
	"platform/internal/notify"
	"platform/internal/ratelimit"
	"platform/internal/repository"
	"platform/internal/service"
	"platform/internal/shop"
	"platform/internal/store/postgres"
	"platform/internal/tenant"
	"platform/pkg/storage"
)

func SetupRouter(db *postgres.DB, s3 *storage.S3Client, queue *asynq.Client, privKey *ecdsa.PrivateKey, pubKey *ecdsa.PublicKey) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(securityMiddleware())
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/healthz", healthzHandler(db))

	// Payment webhooks are public (provider posts server-to-server) and verified by signature inside
	// the billing service. Registered outside /v1 so the tenant resolver does not reject them.
	billingRegistry, graceDays := billing.FromEnv()
	billingSvc := billing.NewService(repository.NewBillingRepository(db), billingRegistry, graceDays)
	router.POST("/billing/webhook/:provider", controllers.NewBillingWebhookController(billingSvc).Handle)

	// تمرير db مباشرة بدلاً من db.DB لكي نستفيد من ميزات الـ Multi-Tenant
	tenantRepo := repository.NewTenantRepository(db)

	// Storefront: catalog + orders + checkout, all on the single store tenant. The webhook is public
	// (provider posts server-to-server, no tenant header) so it resolves the store by a fixed slug.
	storeSlug := getenv("STORE_TENANT_SLUG", "store")
	storefrontURL := getenv("STOREFRONT_URL", "http://localhost:3001")
	shopSvc := shop.NewService(repository.NewShopRepository(db), billingRegistry)
	// Deliver activation codes by email on fulfillment (best-effort; the code is also shown on the
	// order page). The worker sends via SMTP when configured, otherwise logs.
	shopSvc.SetNotifier(notify.NewOrderEmailer(queue))
	shopController := controllers.NewShopController(shopSvc, s3, storefrontURL)
	router.POST("/shop/webhook/:provider", controllers.NewShopWebhookController(shopSvc, tenantRepo, storeSlug).Handle)
	enrollmentRepo := enrollment.NewSessionRepository(db)
	certProvider := service.NewEnrollmentCertProvider(db.DB, s3, make([]byte, 32))

	enrollmentHandler := enrollment.NewHandler(enrollmentRepo, certProvider)
	appController := controllers.NewAppController(repository.NewAppRepository(db))
	activationController := controllers.NewActivationController(repository.NewActivationRepository(db), privKey)
	userController := controllers.NewUserController(repository.NewUserRepository(db))
	signingController := controllers.NewSigningController(repository.NewSigningRepository(db))
	ratingController := controllers.NewRatingController(repository.NewRatingRepository(db))
	notificationController := controllers.NewNotificationController(repository.NewNotificationRepository(db))
	updateController := controllers.NewUpdateController(repository.NewUpdateRepository(db))
	installController := controllers.NewInstallController(repository.NewInstallRepository(db), s3)

	// Storefront routes are public (customers browse and buy without an account) but tenant-scoped,
	// so they hang off the same resolver group as the rest of /shop's siblings below.
	shopGroup := router.Group("/shop")
	shopGroup.Use(tenant.ResolverMiddleware(tenantRepo))
	shopGroup.GET("/products", shopController.ListProducts)
	shopGroup.GET("/products/:slug", shopController.GetProduct)
	shopGroup.GET("/products/:slug/image", shopController.ProductImage)
	shopGroup.GET("/products/:slug/reviews", shopController.Reviews)
	// Ratings are guessable-free public writes — cap per IP to blunt spam.
	shopGroup.POST("/products/:slug/ratings", ratelimit.Middleware(5, time.Minute), shopController.SubmitRating)
	shopGroup.POST("/orders", shopController.CreateOrder)
	shopGroup.POST("/orders/:id/checkout", shopController.StartCheckout)
	shopGroup.GET("/orders", shopController.MyOrders)
	shopGroup.GET("/orders/:id", shopController.GetOrder)
	// Manual-payment proof: customers attach a transfer screenshot; rate-limited to blunt abuse.
	shopGroup.POST("/orders/:id/proof", ratelimit.Middleware(5, time.Minute), shopController.UploadOrderProof)
	shopGroup.GET("/orders/:id/proof", shopController.OrderProof)
	// Read-only activation-code status lookup (does not consume a device); rate-limited vs enumeration.
	shopGroup.GET("/activation", ratelimit.Middleware(15, time.Minute), shopController.ActivationStatus)

	v1 := router.Group("/v1")
	v1.Use(tenant.ResolverMiddleware(tenantRepo))
	v1.GET("/enroll", enrollmentHandler.HandleGetProfile)
	v1.POST("/enroll/callback", enrollmentHandler.HandleAppleCallback)
	v1.GET("/enroll/status", enrollmentHandler.HandlePollStatus)
	// Activation codes are guessable secrets — cap validation attempts per IP to blunt brute force.
	v1.POST("/activation/validate", ratelimit.Middleware(10, time.Minute), activationController.Validate)

	// Ratings are intentionally public: the schema allows unactivated users to rate. When a
	// bearer token is present the rating is attributed to that user, otherwise it is anonymous.
	v1.POST("/ratings", ratingController.Submit)
	v1.GET("/ratings/summary", ratingController.Summary)

	// OTA install endpoints are unauthenticated because iOS fetches them directly via
	// itms-services:// with no app-controlled headers. Tenant is resolved from host/subdomain.
	v1.GET("/install/:version_id/manifest.plist", installController.Manifest)
	v1.GET("/install/:version_id/download", installController.Download)

	userProtected := v1.Group("")
	userProtected.Use(auth.RequireAuth(pubKey))
	userProtected.GET("/apps", appController.ListApps)
	userProtected.GET("/apps/:id", appController.GetApp)
	userProtected.GET("/apps/:id/install-info", installController.InstallInfo)
	userProtected.GET("/profile", userController.GetProfile)
	userProtected.GET("/notifications", notificationController.List)
	userProtected.GET("/updates", updateController.CheckUpdates)
	userProtected.PUT("/activation/devices/:device_id/release", userController.ReleaseDevice)
	userProtected.POST("/activation/codes", activationController.Generate)
	userProtected.GET("/signing/stream/:job_id", signingController.StreamProgress)
	return router
}

// getenv returns the environment variable named key, or fallback when it is unset/empty.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// healthzHandler is a readiness probe: it pings the database so a DB outage marks the service
// unhealthy (distinct from /health, which is a bare liveness check).
func healthzHandler(db *postgres.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
