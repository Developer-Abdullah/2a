package admin

import (
	"context"
	"crypto/ecdsa"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"platform/internal/admin/controllers"
	"platform/internal/auth"
	"platform/internal/billing"
	"platform/internal/notify"
	"platform/internal/ratelimit"
	"platform/internal/repository"
	"platform/internal/shop"
	"platform/internal/store/postgres"
	"platform/internal/tenant"
	"platform/pkg/storage"
)

func SetupAdminRouter(db *postgres.DB, s3 *storage.S3Client, queue *asynq.Client, privKey *ecdsa.PrivateKey, pubKey *ecdsa.PublicKey) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(securityMiddleware())
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/healthz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	adminRepo := repository.NewAdminRepository(db)
	tenantRepo := repository.NewTenantRepository(db)

	authController := controllers.NewAdminAuthController(adminRepo, privKey)
	appController := controllers.NewAdminAppController(repository.NewAdminAppRepository(db), s3, queue)
	activationController := controllers.NewAdminActivationController(repository.NewActivationRepository(db))
	ratingController := controllers.NewAdminRatingController(repository.NewRatingRepository(db))
	notificationController := controllers.NewAdminNotificationController(repository.NewNotificationRepository(db), queue)
	dashboardController := controllers.NewAdminDashboardController(repository.NewAdminReadRepository(db))
	platformController := controllers.NewPlatformController(repository.NewPlatformRepository(db))
	reviewController := controllers.NewAdminReviewController(repository.NewReviewRepository(db), queue)

	billingRegistry, graceDays := billing.FromEnv()
	billingSvc := billing.NewService(repository.NewBillingRepository(db), billingRegistry, graceDays)
	billingController := controllers.NewAdminBillingController(billingSvc)

	// Store admin: catalog CRUD + orders. The shop service backs manual order confirmation (mint codes
	// + queue the delivery email) when an admin verifies an offline payment.
	shopSvc := shop.NewService(repository.NewShopRepository(db), billingRegistry)
	shopSvc.SetNotifier(notify.NewOrderEmailer(queue))
	shopController := controllers.NewAdminShopController(repository.NewAdminShopRepository(db), shopSvc, s3)

	v1 := router.Group("/v1/admin")
	// Tight brute-force cap on login, on top of the coarse per-IP limit in securityMiddleware.
	v1.POST("/auth/login", ratelimit.Middleware(10, time.Minute), authController.Login)

	// All data routes require a resolved tenant (X-Tenant-Slug / subdomain), a valid admin JWT, and
	// — for a merchant admin — that the resolved tenant matches the one baked into their token.
	protected := v1.Group("")
	protected.Use(tenant.ResolverMiddleware(tenantRepo))
	protected.Use(auth.RequireAuth(pubKey))
	protected.Use(enforceTenantScope())

	protected.GET("/apps", appController.ListApps)
	protected.POST("/apps", appController.CreateApp)
	protected.GET("/apps/:id/versions", appController.ListVersions)
	// Uploading and submitting a new build require an open subscription (existing builds are unaffected).
	protected.POST("/apps/:id/upload-url", billingGuard(billingSvc), appController.GetUploadURL)
	protected.POST("/apps/:id/versions", billingGuard(billingSvc), appController.CreateVersion)

	protected.POST("/billing/checkout", billingController.Checkout)
	protected.GET("/billing/subscription", billingController.Subscription)

	protected.POST("/activation/codes", activationController.Generate)
	protected.GET("/activation/codes", dashboardController.Codes)

	// Storefront catalog + orders (single store, tenant-scoped like the rest of the protected group).
	protected.GET("/products", shopController.ListProducts)
	protected.POST("/products", shopController.CreateProduct)
	protected.GET("/products/:id", shopController.GetProduct)
	protected.PUT("/products/:id", shopController.UpdateProduct)
	protected.POST("/products/:id/publish", shopController.SetPublished)
	protected.POST("/products/:id/image", shopController.UploadImage)
	protected.GET("/orders", shopController.ListOrders)
	protected.GET("/orders/:id", shopController.GetOrder)
	protected.POST("/orders/:id/confirm", shopController.ConfirmOrder)
	protected.GET("/ratings", ratingController.List)
	protected.POST("/notifications", notificationController.Send)
	protected.GET("/notifications", dashboardController.Notifications)

	// Platform-owner routes: operate ACROSS all tenants (public schema), so NO tenant resolver —
	// just a valid super-admin JWT.
	platformGroup := v1.Group("/platform")
	platformGroup.Use(auth.RequireAuth(pubKey))
	platformGroup.Use(auth.RequirePlatformOwner())
	platformGroup.GET("/tenants", platformController.ListTenants)
	platformGroup.POST("/tenants", platformController.CreateTenant)
	platformGroup.POST("/tenants/:slug/status", platformController.SetStatus)

	// Manual review gate (owner-only, cross-tenant): approve/reject uploaded versions before signing.
	platformGroup.GET("/reviews/pending", reviewController.ListPending)
	platformGroup.POST("/reviews/:slug/:versionId/approve", reviewController.Approve)
	platformGroup.POST("/reviews/:slug/:versionId/reject", reviewController.Reject)

	// Dashboard read-only views
	protected.GET("/stats/overview", dashboardController.Stats)
	protected.GET("/users", dashboardController.Users)
	protected.GET("/users/:id", dashboardController.UserDetail)
	protected.GET("/devices", dashboardController.Devices)
	protected.GET("/signing/jobs", dashboardController.SigningJobs)
	protected.GET("/certificates", dashboardController.Certificates)
	protected.GET("/audit", dashboardController.Audit)
	protected.GET("/admins", dashboardController.Admins)

	return router
}

// billingGuard blocks new-build submissions when the tenant's subscription has lapsed past its grace
// window. It gates only new submissions — already-signed builds keep distributing (see billing docs).
func billingGuard(svc *billing.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := tenant.GetFromContext(c)
		if t != nil {
			allowed, err := svc.DistributionAllowed(c.Request.Context(), t.ID.String(), time.Now())
			if err == nil && !allowed {
				c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{"success": false, "error": gin.H{"message": "subscription expired; renew to submit new builds"}})
				return
			}
		}
		c.Next()
	}
}

// enforceTenantScope blocks a merchant admin (whose token carries a tenant slug) from operating on
// any tenant other than their own. The platform owner's token has an empty tenant and passes freely.
func enforceTenantScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := auth.GetClaims(c)
		t := tenant.GetFromContext(c)
		if claims != nil && claims.TenantID != "" && t != nil && claims.TenantID != t.Slug {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"message": "tenant access denied"}})
			return
		}
		c.Next()
	}
}
