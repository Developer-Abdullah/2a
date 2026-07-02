package controllers

import (
	"context"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"platform/internal/auth"
	"platform/internal/domain"
)

// PlatformStore is the platform-owner surface: it operates across ALL tenants (public schema),
// not scoped to one tenant.
type PlatformStore interface {
	ListTenants(ctx context.Context) ([]domain.TenantSummaryDTO, error)
	CreateTenant(ctx context.Context, slug, name, adminEmail, adminPassword string) error
	SetTenantStatus(ctx context.Context, slug, status string) error
}

type PlatformController struct {
	store PlatformStore
}

func NewPlatformController(store PlatformStore) *PlatformController {
	return &PlatformController{store: store}
}

var slugRe = regexp.MustCompile(`^[a-z0-9-]+$`)

// requireOwner gates the owner endpoints — only the platform owner (a super admin with no tenant
// binding) may manage merchants. A merchant admin's token carries a tenant slug and is rejected.
func requireOwner(c *gin.Context) bool {
	claims := auth.GetClaims(c)
	if claims == nil || claims.Role != "super_admin" || claims.TenantID != "" {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": gin.H{"message": "platform owner access required"}})
		return false
	}
	return true
}

func (ctrl *PlatformController) ListTenants(c *gin.Context) {
	if !requireOwner(c) {
		return
	}
	items, err := ctrl.store.ListTenants(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to list tenants"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": items}})
}

type createTenantRequest struct {
	Slug          string `json:"slug" binding:"required"`
	Name          string `json:"name" binding:"required"`
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminPassword string `json:"admin_password" binding:"required,min=6"`
}

func (ctrl *PlatformController) CreateTenant(c *gin.Context) {
	if !requireOwner(c) {
		return
	}
	var req createTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	if !slugRe.MatchString(req.Slug) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "slug must be lowercase letters, digits and hyphens"}})
		return
	}
	if err := ctrl.store.CreateTenant(c.Request.Context(), req.Slug, req.Name, req.AdminEmail, req.AdminPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"slug": req.Slug}})
}

type statusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (ctrl *PlatformController) SetStatus(c *gin.Context) {
	if !requireOwner(c) {
		return
	}
	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid request"}})
		return
	}
	switch req.Status {
	case "active", "suspended", "deactivated":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid status"}})
		return
	}
	if err := ctrl.store.SetTenantStatus(c.Request.Context(), c.Param("slug"), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"slug": c.Param("slug"), "status": req.Status}})
}
