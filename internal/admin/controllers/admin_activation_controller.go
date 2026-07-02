package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"platform/internal/tenant"
)

type AdminActivationStore interface {
	GenerateCodes(ctx context.Context, tenantID uuid.UUID, count int, codeType, deviceType string, maxDevices, maxUses int, adminID string) ([]string, error)
}

type AdminActivationController struct {
	store AdminActivationStore
}

func NewAdminActivationController(store AdminActivationStore) *AdminActivationController {
	return &AdminActivationController{store: store}
}

type adminGenerateRequest struct {
	Count      int    `json:"count" binding:"required"`
	CodeType   string `json:"code_type"`
	DeviceType string `json:"device_type"`
	MaxDevices int    `json:"max_devices"`
	MaxUses    int    `json:"max_uses"`
}

// Generate issues a batch of activation codes for the tenant.
func (ctrl *AdminActivationController) Generate(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}

	var req adminGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid request"}})
		return
	}

	// created_by_admin_id is left NULL: the JWT subject is a platform_admins row, not a
	// per-tenant admin_users row, so attributing it here would violate the cross-table FK.
	codes, err := ctrl.store.GenerateCodes(c.Request.Context(), t.ID, req.Count, req.CodeType, req.DeviceType, req.MaxDevices, req.MaxUses, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"codes": codes}})
}
