package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"platform/internal/auth"
	"platform/internal/domain"
	"platform/internal/tenant"
	"platform/pkg/response"
)

type NotificationRepo interface {
	ListForUser(ctx context.Context, tenantID uuid.UUID, userID string) ([]domain.NotificationDTO, error)
}

type NotificationController struct {
	repo NotificationRepo
}

func NewNotificationController(repo NotificationRepo) *NotificationController {
	return &NotificationController{repo: repo}
}

// List returns the authenticated user's notification feed.
func (ctrl *NotificationController) List(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	claims := auth.GetClaims(c)
	if currentTenant == nil || claims == nil {
		response.Fail(c, http.StatusUnauthorized, "auth_context_missing", "authentication context missing")
		return
	}

	items, err := ctrl.repo.ListForUser(c.Request.Context(), currentTenant.ID, claims.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "list_notifications_failed", "failed to load notifications")
		return
	}

	response.Success(c, gin.H{"items": items})
}
