package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"

	"platform/internal/auth"
	"platform/internal/domain"
	"platform/internal/repository"
	"platform/internal/signing"
)

// ReviewStore is the persistence surface for the manual review workflow. It operates across tenants
// (public schema + per-tenant schemas) and is only ever reached by the platform owner.
type ReviewStore interface {
	ListPendingReviews(ctx context.Context) ([]domain.PendingReviewDTO, error)
	ApproveVersion(ctx context.Context, tenantSlug, versionID, reviewerID string) (*domain.ApprovalResult, error)
	RejectVersion(ctx context.Context, tenantSlug, versionID, reviewerID, reason string) error
}

type AdminReviewController struct {
	store ReviewStore
	queue *asynq.Client
}

func NewAdminReviewController(store ReviewStore, queue *asynq.Client) *AdminReviewController {
	return &AdminReviewController{store: store, queue: queue}
}

// ListPending returns the owner's cross-tenant review queue.
func (ctrl *AdminReviewController) ListPending(c *gin.Context) {
	if !requireOwner(c) {
		return
	}
	items, err := ctrl.store.ListPendingReviews(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": "failed to list pending reviews"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": items}})
}

// Approve marks a version approved and enqueues its signing job. Signing cannot be triggered any
// other way — this is the only path from upload to a signed build.
func (ctrl *AdminReviewController) Approve(c *gin.Context) {
	if !requireOwner(c) {
		return
	}
	slug := c.Param("slug")
	versionID := c.Param("versionId")
	reviewerID := reviewerFromClaims(c)

	result, err := ctrl.store.ApproveVersion(c.Request.Context(), slug, versionID, reviewerID)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidReviewTransition) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "error": gin.H{"message": "version is not pending review"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}

	if result.JobID != "" && result.CertID != "" {
		if task, terr := signing.NewIPASignTask(signing.IPASignPayload{
			JobID: result.JobID, VersionID: result.VersionID, CertID: result.CertID, TenantID: result.TenantID,
		}); terr == nil {
			_, _ = ctrl.queue.EnqueueContext(c.Request.Context(), task, asynq.Queue("default"), asynq.MaxRetry(3))
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"version_id": versionID, "status": "approved", "signing_job_id": result.JobID}})
}

type rejectRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// Reject marks a version rejected with a required reason. The merchant resubmits as a new version.
func (ctrl *AdminReviewController) Reject(c *gin.Context) {
	if !requireOwner(c) {
		return
	}
	var req rejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"message": "a rejection reason is required"}})
		return
	}
	slug := c.Param("slug")
	versionID := c.Param("versionId")

	err := ctrl.store.RejectVersion(c.Request.Context(), slug, versionID, reviewerFromClaims(c), req.Reason)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidReviewTransition) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "error": gin.H{"message": "version is not pending review"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"version_id": versionID, "status": "rejected"}})
}

func reviewerFromClaims(c *gin.Context) string {
	if claims := auth.GetClaims(c); claims != nil {
		return claims.UserID
	}
	return ""
}
