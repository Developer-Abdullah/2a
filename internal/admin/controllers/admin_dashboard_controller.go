package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/tenant"
)

// AdminReadStore is the read surface backing the dashboard list/stat pages.
type AdminReadStore interface {
	Stats(ctx context.Context, tenantID uuid.UUID) (*domain.AdminStatsDTO, error)
	SalesStats(ctx context.Context, tenantID uuid.UUID) (*domain.SalesStats, error)
	ListNotifications(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminNotificationDTO, error)
	ListCodes(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminCodeDTO, error)
	ListUsers(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminUserDTO, error)
	ListDevices(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminDeviceDTO, error)
	ListSigningJobs(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminSigningJobDTO, error)
	ListCertificates(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminCertDTO, error)
	ListAudit(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminAuditDTO, error)
	ListAdmins(ctx context.Context) ([]domain.AdminAccountDTO, error)
	UserDetail(ctx context.Context, tenantID uuid.UUID, userID string) (*domain.AdminUserDetail, error)
}

type AdminDashboardController struct {
	store AdminReadStore
}

func NewAdminDashboardController(store AdminReadStore) *AdminDashboardController {
	return &AdminDashboardController{store: store}
}

func (ctrl *AdminDashboardController) Stats(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}
	stats, err := ctrl.store.Stats(c.Request.Context(), t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load stats"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func (ctrl *AdminDashboardController) SalesStats(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}
	stats, err := ctrl.store.SalesStats(c.Request.Context(), t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load sales stats"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func (ctrl *AdminDashboardController) Notifications(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}
	items, err := ctrl.store.ListNotifications(c.Request.Context(), t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load notifications"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": items}})
}

func (ctrl *AdminDashboardController) Codes(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}
	items, err := ctrl.store.ListCodes(c.Request.Context(), t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load codes"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": items}})
}

func (ctrl *AdminDashboardController) Users(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}
	items, err := ctrl.store.ListUsers(c.Request.Context(), t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load users"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": items}})
}

func (ctrl *AdminDashboardController) Devices(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}
	items, err := ctrl.store.ListDevices(c.Request.Context(), t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load devices"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": items}})
}

func (ctrl *AdminDashboardController) SigningJobs(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}
	items, err := ctrl.store.ListSigningJobs(c.Request.Context(), t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load signing jobs"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": items}})
}

func (ctrl *AdminDashboardController) Certificates(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}
	items, err := ctrl.store.ListCertificates(c.Request.Context(), t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load certificates"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": items}})
}

func (ctrl *AdminDashboardController) Audit(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}
	items, err := ctrl.store.ListAudit(c.Request.Context(), t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load audit log"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": items}})
}

func (ctrl *AdminDashboardController) Admins(c *gin.Context) {
	items, err := ctrl.store.ListAdmins(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to load admins"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": items}})
}

func (ctrl *AdminDashboardController) UserDetail(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "tenant context missing"}})
		return
	}
	detail, err := ctrl.store.UserDetail(c.Request.Context(), t.ID, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": gin.H{"message": "user not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": detail})
}
