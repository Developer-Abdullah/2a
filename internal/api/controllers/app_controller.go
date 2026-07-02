package controllers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/tenant"
	"platform/pkg/response"
)

type AppRepository interface {
	ListApps(ctx context.Context, tenantID uuid.UUID, cursor string, limit int, category string) ([]domain.AppItemDTO, string, error)
	GetAppByID(ctx context.Context, tenantID uuid.UUID, appID string) (*domain.AppItemDTO, error)
}

type AppController struct {
	repo AppRepository
}

func NewAppController(repo AppRepository) *AppController {
	return &AppController{repo: repo}
}

func (ctrl *AppController) ListApps(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	if currentTenant == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "tenant context missing")
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_limit", "limit must be a number")
		return
	}

	items, nextCursor, err := ctrl.repo.ListApps(
		c.Request.Context(),
		currentTenant.ID,
		c.Query("cursor"),
		limit,
		c.Query("category"),
	)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "list_apps_failed", "failed to list applications")
		return
	}

	response.Success(c, gin.H{
		"items": items,
		"pagination": gin.H{
			"has_more":    nextCursor != "",
			"next_cursor": nextCursor,
		},
	})
}

func (ctrl *AppController) GetApp(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	if currentTenant == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "tenant context missing")
		return
	}

	app, err := ctrl.repo.GetAppByID(c.Request.Context(), currentTenant.ID, c.Param("id"))
	if err != nil {
		if err == sql.ErrNoRows {
			response.Fail(c, http.StatusNotFound, "app_not_found", "application not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "get_app_failed", "failed to get application")
		return
	}

	response.Success(c, app)
}
