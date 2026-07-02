package controllers

import (
	"context"
	"crypto/ecdsa"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"platform/internal/auth"
	"platform/internal/domain"
)

type AdminAuthRepo interface {
	AuthenticateAdmin(ctx context.Context, email, password string) (*domain.AdminUser, error)
}

type AdminAuthController struct {
	repo    AdminAuthRepo
	privKey *ecdsa.PrivateKey
}

func NewAdminAuthController(repo AdminAuthRepo, privKey *ecdsa.PrivateKey) *AdminAuthController {
	return &AdminAuthController{repo: repo, privKey: privKey}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (ctrl *AdminAuthController) Login(c *gin.Context) {
	if ctrl.privKey == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "auth signing key is not configured"}})
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "invalid login request"}})
		return
	}

	user, err := ctrl.repo.AuthenticateAdmin(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": gin.H{"message": "invalid credentials"}})
		return
	}

	token, err := auth.GenerateToken(auth.CustomClaims{
		UserID:   user.ID.String(),
		Role:     string(user.Role),
		DeviceID: "admin-dashboard",
		TenantID: user.TenantSlug, // binds a merchant admin to their store; empty for the platform owner
	}, ctrl.privKey, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to issue access token"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token": token,
			"user": gin.H{
				"id":          user.ID.String(),
				"email":       user.Email,
				"role":        user.Role,
				"tenant_slug": user.TenantSlug,
				"is_owner":    user.TenantSlug == "",
			},
		},
	})
}
