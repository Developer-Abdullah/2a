package controllers

import (
	"context"
	"crypto/ecdsa"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"platform/internal/auth"
	"platform/internal/tenant"
	"platform/pkg/response"
)

type ActivationRepo interface {
	GenerateCodes(ctx context.Context, tenantID uuid.UUID, count int, codeType, deviceType string, maxDevices, maxUses int, adminID string) ([]string, error)
	ValidateAndConsume(ctx context.Context, tenantID uuid.UUID, code, deviceID, hardwareType string) (uuid.UUID, error)
}

type ActivationController struct {
	repo       ActivationRepo
	privateKey *ecdsa.PrivateKey
}

func NewActivationController(repo ActivationRepo, privateKey *ecdsa.PrivateKey) *ActivationController {
	return &ActivationController{repo: repo, privateKey: privateKey}
}

type validateActivationRequest struct {
	Code       string `json:"code" binding:"required"`
	DeviceID   string `json:"device_id" binding:"required"`
	DeviceType string `json:"device_type" binding:"required"`
}

type generateActivationRequest struct {
	Count      int    `json:"count" binding:"required"`
	CodeType   string `json:"code_type"`
	DeviceType string `json:"device_type"`
	MaxDevices int    `json:"max_devices"`
	MaxUses    int    `json:"max_uses"`
}

func (ctrl *ActivationController) Validate(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	if currentTenant == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "tenant context missing")
		return
	}

	var req validateActivationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_request", "invalid activation request")
		return
	}

	userID, err := ctrl.repo.ValidateAndConsume(c.Request.Context(), currentTenant.ID, req.Code, req.DeviceID, req.DeviceType)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "activation_failed", "activation code is invalid or exhausted")
		return
	}

	accessTTL := 24 * time.Hour
	refreshTTL := 30 * 24 * time.Hour
	accessToken, err := auth.GenerateToken(auth.CustomClaims{
		UserID:   userID.String(),
		TenantID: currentTenant.ID.String(),
		Role:     "user",
		DeviceID: req.DeviceID,
	}, ctrl.privateKey, accessTTL)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "token_failed", "failed to issue access token")
		return
	}
	refreshToken, err := auth.GenerateToken(auth.CustomClaims{
		UserID:   userID.String(),
		TenantID: currentTenant.ID.String(),
		Role:     "refresh",
		DeviceID: req.DeviceID,
	}, ctrl.privateKey, refreshTTL)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "token_failed", "failed to issue refresh token")
		return
	}

	response.Success(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user_id":       userID.String(),
		"expires_in":    int(accessTTL.Seconds()),
	})
}

func (ctrl *ActivationController) Generate(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	if currentTenant == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "tenant context missing")
		return
	}

	var req generateActivationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_request", "invalid code generation request")
		return
	}

	adminID := ""
	if claims := auth.GetClaims(c); claims != nil {
		adminID = claims.UserID
	}

	codes, err := ctrl.repo.GenerateCodes(c.Request.Context(), currentTenant.ID, req.Count, req.CodeType, req.DeviceType, req.MaxDevices, req.MaxUses, adminID)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "generate_codes_failed", err.Error())
		return
	}

	response.Success(c, gin.H{"codes": codes})
}
