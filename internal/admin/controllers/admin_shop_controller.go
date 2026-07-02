package controllers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/tenant"
)

// AdminShopStore is the persistence surface for the store-admin catalog and orders views.
type AdminShopStore interface {
	ListProducts(ctx context.Context, tenantID uuid.UUID) ([]domain.Product, error)
	GetProduct(ctx context.Context, tenantID uuid.UUID, id string) (*domain.Product, error)
	CreateProduct(ctx context.Context, tenantID uuid.UUID, in domain.ProductInput) (string, error)
	UpdateProduct(ctx context.Context, tenantID uuid.UUID, id string, in domain.ProductInput) error
	SetPublished(ctx context.Context, tenantID uuid.UUID, id string, published bool) error
	ListOrders(ctx context.Context, tenantID uuid.UUID, status string, limit int) ([]domain.OrderSummary, error)
	GetOrder(ctx context.Context, tenantID uuid.UUID, id string) (*domain.Order, error)
}

// OrderConfirmer manually fulfills an order after the admin verifies an offline (manual) payment.
type OrderConfirmer interface {
	ConfirmOrder(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, error)
}

type AdminShopController struct {
	store     AdminShopStore
	confirmer OrderConfirmer
}

func NewAdminShopController(store AdminShopStore, confirmer OrderConfirmer) *AdminShopController {
	return &AdminShopController{store: store, confirmer: confirmer}
}

// ListProducts returns every product (published or draft).
func (ctrl *AdminShopController) ListProducts(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	products, err := ctrl.store.ListProducts(c.Request.Context(), t.ID)
	if err != nil {
		adminFail(c, http.StatusInternalServerError, "failed to list products")
		return
	}
	adminOK(c, gin.H{"items": products})
}

// GetProduct returns one product with all prices.
func (ctrl *AdminShopController) GetProduct(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	product, err := ctrl.store.GetProduct(c.Request.Context(), t.ID, c.Param("id"))
	if err != nil {
		adminFail(c, http.StatusNotFound, "product not found")
		return
	}
	adminOK(c, gin.H{"product": product})
}

// CreateProduct inserts a new product with its prices.
func (ctrl *AdminShopController) CreateProduct(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	var in domain.ProductInput
	if err := c.ShouldBindJSON(&in); err != nil {
		adminFail(c, http.StatusBadRequest, err.Error())
		return
	}
	if in.Slug == "" || in.Name == "" {
		adminFail(c, http.StatusBadRequest, "slug and name are required")
		return
	}
	id, err := ctrl.store.CreateProduct(c.Request.Context(), t.ID, in)
	if err != nil {
		adminFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	adminOK(c, gin.H{"id": id})
}

// UpdateProduct updates a product's core fields and replaces its price set.
func (ctrl *AdminShopController) UpdateProduct(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	var in domain.ProductInput
	if err := c.ShouldBindJSON(&in); err != nil {
		adminFail(c, http.StatusBadRequest, err.Error())
		return
	}
	if in.Slug == "" || in.Name == "" {
		adminFail(c, http.StatusBadRequest, "slug and name are required")
		return
	}
	if err := ctrl.store.UpdateProduct(c.Request.Context(), t.ID, c.Param("id"), in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			adminFail(c, http.StatusNotFound, "product not found")
			return
		}
		adminFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	adminOK(c, gin.H{"id": c.Param("id")})
}

type setPublishedRequest struct {
	Published bool `json:"published"`
}

// SetPublished toggles a product's storefront visibility.
func (ctrl *AdminShopController) SetPublished(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	var req setPublishedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		adminFail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := ctrl.store.SetPublished(c.Request.Context(), t.ID, c.Param("id"), req.Published); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			adminFail(c, http.StatusNotFound, "product not found")
			return
		}
		adminFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	adminOK(c, gin.H{"id": c.Param("id"), "published": req.Published})
}

// ListOrders returns order summaries, optionally filtered by ?status=.
func (ctrl *AdminShopController) ListOrders(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	orders, err := ctrl.store.ListOrders(c.Request.Context(), t.ID, c.Query("status"), 0)
	if err != nil {
		adminFail(c, http.StatusInternalServerError, "failed to list orders")
		return
	}
	adminOK(c, gin.H{"items": orders})
}

// ConfirmOrder manually marks an order as paid and fulfills it (mints codes + queues the email). This
// is the manual-payment path: the customer pays offline, the admin confirms here.
func (ctrl *AdminShopController) ConfirmOrder(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	if ctrl.confirmer == nil {
		adminFail(c, http.StatusServiceUnavailable, "confirmation not configured")
		return
	}
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		adminFail(c, http.StatusBadRequest, "invalid order id")
		return
	}
	order, err := ctrl.confirmer.ConfirmOrder(c.Request.Context(), t.ID, orderID)
	if err != nil {
		adminFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	adminOK(c, gin.H{"order": order})
}

// GetOrder returns an order with items and minted codes.
func (ctrl *AdminShopController) GetOrder(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	order, err := ctrl.store.GetOrder(c.Request.Context(), t.ID, c.Param("id"))
	if err != nil {
		adminFail(c, http.StatusNotFound, "order not found")
		return
	}
	adminOK(c, gin.H{"order": order})
}
