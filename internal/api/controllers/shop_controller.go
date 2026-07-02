package controllers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"platform/internal/billing"
	"platform/internal/domain"
	"platform/internal/shop"
	"platform/internal/tenant"
	"platform/pkg/response"
)

// ShopController serves the public storefront: catalog reads, order creation, checkout start, and
// order status. All handlers resolve the current store from the tenant context.
type ShopController struct {
	svc *shop.Service
	// publicBaseURL is where the payment provider returns the buyer after checkout; the order id is
	// appended so the storefront can show the order status page.
	publicBaseURL string
}

func NewShopController(svc *shop.Service, publicBaseURL string) *ShopController {
	return &ShopController{svc: svc, publicBaseURL: strings.TrimRight(publicBaseURL, "/")}
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
