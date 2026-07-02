package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"platform/internal/billing"
	"platform/internal/tenant"
)

// AdminBillingController exposes the merchant-facing billing surface: start a checkout for a plan
// and read the current subscription status. Both are tenant-scoped via the resolved tenant context.
type AdminBillingController struct {
	svc *billing.Service
}

func NewAdminBillingController(svc *billing.Service) *AdminBillingController {
	return &AdminBillingController{svc: svc}
}

type checkoutRequest struct {
	Provider    string `json:"provider" binding:"required"`
	PlanID      string `json:"plan_id" binding:"required"`
	CallbackURL string `json:"callback_url"`
}

// Checkout starts a hosted checkout for the authenticated tenant and returns the redirect URL.
func (ctrl *AdminBillingController) Checkout(c *gin.Context) {
	if ctrl == nil || ctrl.svc == nil {
		adminFail(c, http.StatusServiceUnavailable, "billing not configured")
		return
	}
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	var req checkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		adminFail(c, http.StatusBadRequest, err.Error())
		return
	}
	callback := req.CallbackURL
	if callback == "" {
		callback = "https://" + c.Request.Host + "/billing/return"
	}

	url, err := ctrl.svc.StartCheckout(c.Request.Context(), req.Provider, t.ID.String(), req.PlanID, callback)
	if err != nil {
		adminFail(c, http.StatusBadGateway, err.Error())
		return
	}
	adminOK(c, gin.H{"redirect_url": url})
}

// Subscription returns the tenant's current subscription state for the billing page.
func (ctrl *AdminBillingController) Subscription(c *gin.Context) {
	if ctrl == nil || ctrl.svc == nil {
		adminFail(c, http.StatusServiceUnavailable, "billing not configured")
		return
	}
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	snap, err := ctrl.svc.SubscriptionView(c.Request.Context(), t.ID.String())
	if err != nil {
		adminFail(c, http.StatusInternalServerError, "failed to load subscription")
		return
	}
	adminOK(c, gin.H{
		"status":            snap.Status,
		"expires_at":        snap.ExpiresAt,
		"grace_until":       snap.GraceUntil,
		"distribution_open": snap.Allowed,
	})
}
