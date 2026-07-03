package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"platform/internal/auth"
	"platform/internal/billing"
	"platform/internal/domain"
	"platform/internal/shop"
	"platform/internal/tenant"
	"platform/pkg/response"
	"platform/pkg/storage"
)

// ShopController serves the public storefront: catalog reads, order creation, checkout start, and
// order status. All handlers resolve the current store from the tenant context.
type ShopController struct {
	svc *shop.Service
	s3  *storage.S3Client
	// publicBaseURL is where the payment provider returns the buyer after checkout; the order id is
	// appended so the storefront can show the order status page.
	publicBaseURL string
}

func NewShopController(svc *shop.Service, s3 *storage.S3Client, publicBaseURL string) *ShopController {
	return &ShopController{svc: svc, s3: s3, publicBaseURL: strings.TrimRight(publicBaseURL, "/")}
}

func (ctrl *ShopController) currency(c *gin.Context) string {
	if cur := c.Query("currency"); cur != "" {
		return cur
	}
	return "EGP"
}

// ListProducts: GET /shop/products?currency=EGP
func (ctrl *ShopController) ListProducts(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "store context missing")
		return
	}
	products, err := ctrl.svc.ListProducts(c.Request.Context(), t.ID, ctrl.currency(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "list_failed", "could not load products")
		return
	}
	response.Success(c, gin.H{"products": products})
}

// GetProduct: GET /shop/products/:slug?currency=EGP
func (ctrl *ShopController) GetProduct(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "store context missing")
		return
	}
	product, err := ctrl.svc.GetProduct(c.Request.Context(), t.ID, c.Param("slug"), ctrl.currency(c))
	if err != nil {
		response.Fail(c, http.StatusNotFound, "not_found", "product not found")
		return
	}
	response.Success(c, gin.H{"product": product})
}

// ProductImage streams a product's image from object storage. Public and header-free so an <img> tag
// (proxied by the storefront) can load it. Returns 404 when the product has no image.
func (ctrl *ShopController) ProductImage(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "store context missing")
		return
	}
	if ctrl.s3 == nil {
		response.Fail(c, http.StatusServiceUnavailable, "storage_unavailable", "image storage not configured")
		return
	}
	key, err := ctrl.svc.ProductImageKey(c.Request.Context(), t.ID, c.Param("slug"))
	if err != nil || key == "" {
		response.Fail(c, http.StatusNotFound, "no_image", "no image for this product")
		return
	}
	body, contentType, err := ctrl.s3.GetObjectStream(c.Request.Context(), key)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "no_image", "image not found")
		return
	}
	defer body.Close()
	if contentType == "" {
		contentType = "image/jpeg"
	}
	c.Header("Cache-Control", "public, max-age=300")
	c.DataFromReader(http.StatusOK, -1, contentType, body, nil)
}

// Reviews returns recent public reviews (comments) for a product.
func (ctrl *ShopController) Reviews(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "store context missing")
		return
	}
	reviews, err := ctrl.svc.ListReviews(c.Request.Context(), t.ID, c.Param("slug"), 10)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "reviews_failed", "could not load reviews")
		return
	}
	response.Success(c, gin.H{"reviews": reviews})
}

type submitProductRatingRequest struct {
	Rating  int     `json:"rating" binding:"required,min=1,max=5"`
	Comment *string `json:"comment"`
}

// SubmitRating records a 1-5 rating for a product. Auth is optional (attributed to the user when a
// token is present, else anonymous). Rate-limited at the route to blunt spam.
func (ctrl *ShopController) SubmitRating(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "store context missing")
		return
	}
	var req submitProductRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_request", "rating must be an integer between 1 and 5")
		return
	}
	var userID *string
	if claims := auth.GetClaims(c); claims != nil && claims.UserID != "" {
		uid := claims.UserID
		userID = &uid
	}
	ipHash := hashIP(c.ClientIP())
	if err := ctrl.svc.SubmitRating(c.Request.Context(), t.ID, c.Param("slug"), userID, req.Rating, req.Comment, ipHash); err != nil {
		response.Fail(c, http.StatusNotFound, "product_not_found", "product not found")
		return
	}
	response.Success(c, gin.H{"submitted": true})
}

// UploadOrderProof stores the customer's transfer-screenshot (multipart "file") for a manual payment
// and links it to the order so the admin can review it before confirming.
func (ctrl *ShopController) UploadOrderProof(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "store context missing")
		return
	}
	if ctrl.s3 == nil {
		response.Fail(c, http.StatusServiceUnavailable, "storage_unavailable", "upload storage not configured")
		return
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_id", "invalid order id")
		return
	}
	// Only orders that exist accept a proof; fetch also confirms tenant scoping.
	order, err := ctrl.svc.GetOrder(c.Request.Context(), t.ID, orderID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "not_found", "order not found")
		return
	}
	if order.Status == "failed" {
		response.Fail(c, http.StatusBadRequest, "order_closed", "order is closed")
		return
	}

	header, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "no_file", "no file uploaded")
		return
	}
	if header.Size > 8<<20 { // 8 MiB cap
		response.Fail(c, http.StatusRequestEntityTooLarge, "too_large", "image too large (max 8MB)")
		return
	}
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		response.Fail(c, http.StatusBadRequest, "not_image", "only images are accepted")
		return
	}
	file, err := header.Open()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "read_failed", "cannot read upload")
		return
	}
	defer file.Close()

	key := fmt.Sprintf("%s/payment-proofs/%s/%s", t.ID.String(), orderID, uuid.NewString())
	if err := ctrl.s3.PutObjectStream(c.Request.Context(), key, file, contentType); err != nil {
		response.Fail(c, http.StatusInternalServerError, "store_failed", "failed to store image")
		return
	}
	if err := ctrl.svc.SetOrderProof(c.Request.Context(), t.ID, orderID, key); err != nil {
		response.Fail(c, http.StatusInternalServerError, "save_failed", "failed to save proof")
		return
	}
	response.Success(c, gin.H{"uploaded": true})
}

// OrderProof streams the uploaded transfer screenshot for an order. The order id (an unguessable
// UUID) is the access token, same as the codes shown on the order page.
func (ctrl *ShopController) OrderProof(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "store context missing")
		return
	}
	if ctrl.s3 == nil {
		response.Fail(c, http.StatusServiceUnavailable, "storage_unavailable", "storage not configured")
		return
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_id", "invalid order id")
		return
	}
	key, err := ctrl.svc.OrderProofKey(c.Request.Context(), t.ID, orderID)
	if err != nil || key == "" {
		response.Fail(c, http.StatusNotFound, "no_proof", "no proof for this order")
		return
	}
	body, contentType, err := ctrl.s3.GetObjectStream(c.Request.Context(), key)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "no_proof", "proof not found")
		return
	}
	defer body.Close()
	if contentType == "" {
		contentType = "image/jpeg"
	}
	c.Header("Cache-Control", "private, max-age=60")
	c.DataFromReader(http.StatusOK, -1, contentType, body, nil)
}

// MyOrders lists a customer's orders by email (order tracking). Returns summaries only; codes remain
// on the per-order page.
func (ctrl *ShopController) MyOrders(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "store context missing")
		return
	}
	email := strings.TrimSpace(c.Query("email"))
	if email == "" {
		response.Fail(c, http.StatusBadRequest, "email_required", "email is required")
		return
	}
	orders, err := ctrl.svc.OrdersByEmail(c.Request.Context(), t.ID, email)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "lookup_failed", "could not load orders")
		return
	}
	response.Success(c, gin.H{"orders": orders})
}

// hashIP keeps raw IPs out of the ratings table while preserving the daily-uniqueness key.
func hashIP(ip string) string {
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:])
}

type createOrderRequest struct {
	Email    string            `json:"email" binding:"required,email"`
	Phone    string            `json:"phone"`
	Currency string            `json:"currency" binding:"required"`
	Items    []domain.CartLine `json:"items" binding:"required,min=1,dive"`
}

// CreateOrder: POST /shop/orders — creates a pending order priced from the catalog.
func (ctrl *ShopController) CreateOrder(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "store context missing")
		return
	}
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_request", "invalid order request")
		return
	}
	order, err := ctrl.svc.CreateOrder(c.Request.Context(), t.ID, req.Email, req.Phone, req.Currency, req.Items)
	if err != nil {
		switch {
		case errors.Is(err, shop.ErrEmptyCart):
			response.Fail(c, http.StatusBadRequest, "empty_cart", "cart is empty")
		case errors.Is(err, shop.ErrUnsupportedCurrency):
			response.Fail(c, http.StatusBadRequest, "unsupported_currency", "currency is not supported")
		default:
			response.Fail(c, http.StatusBadRequest, "order_failed", "could not create order")
		}
		return
	}
	response.Success(c, gin.H{"order": order})
}

// StartCheckout: POST /shop/orders/:id/checkout — opens a hosted payment and returns the redirect URL.
func (ctrl *ShopController) StartCheckout(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "store context missing")
		return
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_id", "invalid order id")
		return
	}
	callbackURL := ctrl.publicBaseURL + "/order/" + orderID.String()
	redirectURL, err := ctrl.svc.StartCheckout(c.Request.Context(), t.ID, orderID, callbackURL)
	if err != nil {
		log.Ctx(c.Request.Context()).Error().Err(err).Str("order_id", orderID.String()).Msg("checkout failed")
		response.Fail(c, http.StatusBadGateway, "checkout_failed", "could not start payment")
		return
	}
	response.Success(c, gin.H{"redirect_url": redirectURL})
}

// GetOrder: GET /shop/orders/:id — order status plus codes once fulfilled.
func (ctrl *ShopController) GetOrder(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "store context missing")
		return
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_id", "invalid order id")
		return
	}
	order, err := ctrl.svc.GetOrder(c.Request.Context(), t.ID, orderID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "not_found", "order not found")
		return
	}
	response.Success(c, gin.H{"order": order})
}

// ShopWebhookController receives payment webhooks for storefront orders. It is public (the provider
// posts server-to-server) and trusted only after the provider signature verifies inside the service.
// The store tenant is resolved from a fixed slug because the provider sends no tenant header.
type ShopWebhookController struct {
	svc       *shop.Service
	tenants   TenantLookup
	storeSlug string
}

// TenantLookup resolves the single store tenant by slug for the webhook path.
type TenantLookup interface {
	GetBySlugOrID(ctx context.Context, identifier string) (*domain.Tenant, error)
}

func NewShopWebhookController(svc *shop.Service, tenants TenantLookup, storeSlug string) *ShopWebhookController {
	if storeSlug == "" {
		storeSlug = "store"
	}
	return &ShopWebhookController{svc: svc, tenants: tenants, storeSlug: storeSlug}
}

func (ctrl *ShopWebhookController) Handle(c *gin.Context) {
	if ctrl == nil || ctrl.svc == nil {
		response.Fail(c, http.StatusServiceUnavailable, "not_configured", "shop billing not configured")
		return
	}
	provider := c.Param("provider")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "bad_body", "cannot read body")
		return
	}
	t, err := ctrl.tenants.GetBySlugOrID(c.Request.Context(), ctrl.storeSlug)
	if err != nil {
		response.Fail(c, http.StatusServiceUnavailable, "store_missing", "store not found")
		return
	}
	if err := ctrl.svc.HandleWebhook(c.Request.Context(), t.ID, provider, c.Request.Header, body); err != nil {
		if errors.Is(err, billing.ErrInvalidSignature) {
			response.Fail(c, http.StatusUnauthorized, "invalid_signature", "invalid signature")
			return
		}
		log.Ctx(c.Request.Context()).Error().Err(err).Str("provider", provider).Msg("shop webhook failed")
		response.Fail(c, http.StatusInternalServerError, "processing_failed", "processing failed")
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}
