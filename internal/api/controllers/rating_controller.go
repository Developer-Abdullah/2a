package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"platform/internal/auth"
	"platform/internal/domain"
	"platform/internal/tenant"
	"platform/pkg/response"
)

type RatingRepo interface {
	SubmitRating(ctx context.Context, tenantID uuid.UUID, userID, deviceID *string, rating int, comment *string, ipHash string) error
	GetSummary(ctx context.Context, tenantID uuid.UUID) (*domain.RatingSummary, error)
}

type RatingController struct {
	repo RatingRepo
}

func NewRatingController(repo RatingRepo) *RatingController {
	return &RatingController{repo: repo}
}

type submitRatingRequest struct {
	Rating  int     `json:"rating" binding:"required,min=1,max=5"`
	Comment *string `json:"comment"`
}

// Submit records a 1-5 rating. Auth is optional: an activated user is attributed when a token
// is present, otherwise the rating is anonymous (the schema allows unactivated ratings).
func (ctrl *RatingController) Submit(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	if currentTenant == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "tenant context missing")
		return
	}

	var req submitRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid_request", "rating must be an integer between 1 and 5")
		return
	}

	var userID *string
	if claims := auth.GetClaims(c); claims != nil && claims.UserID != "" {
		uid := claims.UserID
		userID = &uid
	}

	ipHash := hashClientIP(c.ClientIP())
	if err := ctrl.repo.SubmitRating(c.Request.Context(), currentTenant.ID, userID, nil, req.Rating, req.Comment, ipHash); err != nil {
		response.Fail(c, http.StatusInternalServerError, "submit_rating_failed", "failed to submit rating")
		return
	}

	response.Success(c, gin.H{"submitted": true})
}

// Summary returns aggregate rating metrics for the storefront.
func (ctrl *RatingController) Summary(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	if currentTenant == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "tenant context missing")
		return
	}

	summary, err := ctrl.repo.GetSummary(c.Request.Context(), currentTenant.ID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "rating_summary_failed", "failed to load rating summary")
		return
	}

	response.Success(c, gin.H{
		"average": summary.Average,
		"count":   summary.Count,
		"breakdown": gin.H{
			"1": summary.R1,
			"2": summary.R2,
			"3": summary.R3,
			"4": summary.R4,
			"5": summary.R5,
		},
	})
}

// hashClientIP keeps PII out of the ratings table while preserving the daily-uniqueness key.
func hashClientIP(ip string) string {
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:])
}
