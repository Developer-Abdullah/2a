package controllers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"platform/internal/auth"
	"platform/internal/domain"
	"platform/internal/tenant"
	"platform/pkg/response"
)

type UserRepo interface {
	GetUserProfile(ctx context.Context, tenantID uuid.UUID, userID string) (*domain.User, error)
	GetBoundDevices(ctx context.Context, tenantID uuid.UUID, userID string) ([]domain.Device, error)
	ReleaseDeviceSlot(ctx context.Context, tenantID uuid.UUID, userID string, deviceID string) error
}

type UserController struct {
	repo UserRepo
}

func NewUserController(repo UserRepo) *UserController {
	return &UserController{repo: repo}
}

func (ctrl *UserController) GetProfile(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	claims := auth.GetClaims(c)
	if currentTenant == nil || claims == nil {
		response.Fail(c, http.StatusUnauthorized, "auth_context_missing", "authentication context missing")
		return
	}

	user, err := ctrl.repo.GetUserProfile(c.Request.Context(), currentTenant.ID, claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Fail(c, http.StatusNotFound, "user_not_found", "user not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "get_profile_failed", "failed to load profile")
		return
	}

	devices, err := ctrl.repo.GetBoundDevices(c.Request.Context(), currentTenant.ID, claims.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "get_devices_failed", "failed to load devices")
		return
	}

	response.Success(c, gin.H{
		"user":    user,
		"devices": devices,
	})
}

func (ctrl *UserController) ReleaseDevice(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	claims := auth.GetClaims(c)
	if currentTenant == nil || claims == nil {
		response.Fail(c, http.StatusUnauthorized, "auth_context_missing", "authentication context missing")
		return
	}

	err := ctrl.repo.ReleaseDeviceSlot(c.Request.Context(), currentTenant.ID, claims.UserID, c.Param("device_id"))
	if err != nil {
		if err == sql.ErrNoRows {
			response.Fail(c, http.StatusNotFound, "device_not_found", "device not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, "release_device_failed", "failed to release device")
		return
	}

	response.Success(c, gin.H{"released": true})
}
