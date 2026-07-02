package controllers

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"platform/internal/billing"
)

// BillingWebhookController receives provider payment webhooks. These endpoints are public (the
// provider posts server-to-server with no tenant header) and are trusted ONLY after the provider's
// signature verifies inside the billing service.
type BillingWebhookController struct {
	svc *billing.Service
}

func NewBillingWebhookController(svc *billing.Service) *BillingWebhookController {
	return &BillingWebhookController{svc: svc}
}

// Handle verifies and applies a webhook for /billing/webhook/:provider. It always returns 200 for a
// verified-but-unactionable event; only a signature failure or persistence error is surfaced as an
// error status so the provider retries.
func (ctrl *BillingWebhookController) Handle(c *gin.Context) {
	if ctrl == nil || ctrl.svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not configured"})
		return
	}
	provider := c.Param("provider")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	if err := ctrl.svc.HandleWebhook(c.Request.Context(), provider, c.Request.Header, body); err != nil {
		if errors.Is(err, billing.ErrInvalidSignature) {
			// Do not retry an unsigned/forged webhook.
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
		log.Ctx(c.Request.Context()).Error().Err(err).Str("provider", provider).Msg("webhook processing failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "processing failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
