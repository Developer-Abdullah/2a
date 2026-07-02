package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/tenant"
	"platform/pkg/response"
)

type UpdateRepo interface {
	ListUpdates(ctx context.Context, tenantID uuid.UUID) ([]domain.UpdateDTO, error)
}

type UpdateController struct {
	repo UpdateRepo
}

func NewUpdateController(repo UpdateRepo) *UpdateController {
	return &UpdateController{repo: repo}
}

// CheckUpdates returns the available update transitions, driving the iOS update badge/banner.
func (ctrl *UpdateController) CheckUpdates(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	if currentTenant == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "tenant context missing")
		return
	}

	updates, err := ctrl.repo.ListUpdates(c.Request.Context(), currentTenant.ID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "list_updates_failed", "failed to load updates")
		return
	}

	response.Success(c, gin.H{"items": updates})
}
