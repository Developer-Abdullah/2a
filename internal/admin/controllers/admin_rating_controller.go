package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/tenant"
)

type AdminRatingStore interface {
	GetSummary(ctx context.Context, tenantID uuid.UUID) (*domain.RatingSummary, error)
	ListRecent(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.PlatformRating, error)
}

type AdminRatingController struct {
	store AdminRatingStore
}

func NewAdminRatingController(store AdminRatingStore) *AdminRatingController {
	return &AdminRatingController{store: store}
}

// List returns the rating summary plus the most recent individual ratings.
func (ctrl *AdminRatingController) List(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}

	summary, err := ctrl.store.GetSummary(c.Request.Context(), t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load ratings"}})
		return
	}

	items, err := ctrl.store.ListRecent(c.Request.Context(), t.ID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load ratings"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"summary": gin.H{
				"average":   summary.Average,
				"count":     summary.Count,
				"breakdown": gin.H{"1": summary.R1, "2": summary.R2, "3": summary.R3, "4": summary.R4, "5": summary.R5},
			},
			"items": items,
		},
	})
}
