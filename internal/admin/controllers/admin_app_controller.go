package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"platform/internal/domain"
	"platform/internal/tenant"
	"platform/pkg/storage"
)

// AdminAppStore is the persistence surface the admin app controller depends on.
type AdminAppStore interface {
	GetAppByID(ctx context.Context, tenantID uuid.UUID, appID string) (*domain.AppItemDTO, error)
	CreateApp(ctx context.Context, tenantID uuid.UUID, bundleID, name, category, description string, features []string, publish bool) (string, error)
	ListApps(ctx context.Context, tenantID uuid.UUID) ([]domain.AppItemDTO, error)
	ListVersions(ctx context.Context, tenantID uuid.UUID, appID string) ([]domain.VersionDTO, error)
	ActiveCertID(ctx context.Context, tenantID uuid.UUID) (string, error)
	CreateVersionWithJob(ctx context.Context, tenantID uuid.UUID, appID, version, build, rawKey string, sizeBytes int64, certID string) (string, string, error)
}

type AdminAppController struct {
	store    AdminAppStore
	s3Client *storage.S3Client
	queue    *asynq.Client
}

func NewAdminAppController(store AdminAppStore, s3Client *storage.S3Client, queue *asynq.Client) *AdminAppController {
	return &AdminAppController{store: store, s3Client: s3Client, queue: queue}
}

func adminFail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"success": false, "error": gin.H{"message": message}})
}

func adminOK(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// ListApps returns every app (published or draft) for the dashboard.
func (ctrl *AdminAppController) ListApps(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	apps, err := ctrl.store.ListApps(c.Request.Context(), t.ID)
	if err != nil {
		adminFail(c, http.StatusInternalServerError, "failed to list apps")
		return
	}
	adminOK(c, gin.H{"items": apps})
}

type createAppRequest struct {
	BundleIdentifier string   `json:"bundle_identifier" binding:"required"`
	Name             string   `json:"name" binding:"required"`
	Category         string   `json:"category"`
	Description      string   `json:"description"`
	Features         []string `json:"features"`
	Publish          bool     `json:"publish"`
}

// CreateApp registers a new application.
func (ctrl *AdminAppController) CreateApp(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	var req createAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		adminFail(c, http.StatusBadRequest, err.Error())
		return
	}
	id, err := ctrl.store.CreateApp(c.Request.Context(), t.ID, req.BundleIdentifier, req.Name, req.Category, req.Description, req.Features, req.Publish)
	if err != nil {
		adminFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	adminOK(c, gin.H{"id": id})
}

// ListVersions returns an app's version history (the "signed apps" view).
func (ctrl *AdminAppController) ListVersions(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	versions, err := ctrl.store.ListVersions(c.Request.Context(), t.ID, c.Param("id"))
	if err != nil {
		adminFail(c, http.StatusInternalServerError, "failed to list versions")
		return
	}
	adminOK(c, gin.H{"items": versions})
}

// GetUploadURL hands back a presigned PUT URL so the dashboard can upload a raw IPA to S3.
func (ctrl *AdminAppController) GetUploadURL(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	appID := c.Param("id")
	if _, err := ctrl.store.GetAppByID(c.Request.Context(), t.ID, appID); err != nil {
		adminFail(c, http.StatusNotFound, "app not found")
		return
	}

	uploadID := uuid.NewString()
	expectedKey := fmt.Sprintf("%s/raw-ipa/%s/%s.ipa", t.ID.String(), appID, uploadID)
	presignedURL, err := ctrl.s3Client.GeneratePresignedPutURL(c.Request.Context(), expectedKey, 1*time.Hour)
	if err != nil {
		adminFail(c, http.StatusInternalServerError, "failed to generate upload URL")
		return
	}

	adminOK(c, gin.H{"presigned_url": presignedURL, "expected_key": expectedKey})
}

type createVersionRequest struct {
	Version     string `json:"version" binding:"required"`
	BuildNumber string `json:"build_number"`
	RawIPAS3Key string `json:"raw_ipa_s3_key" binding:"required"`
	SizeBytes   int64  `json:"size_bytes"`
}

// CreateVersion persists a version and its queued signing job in the pending_review state. Signing
// is NOT enqueued here — a platform owner must approve the version first (see AdminReviewController).
// This is the anti-piracy gate: there is no path from upload to a signed build that skips review.
func (ctrl *AdminAppController) CreateVersion(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		adminFail(c, http.StatusBadRequest, "tenant context missing")
		return
	}
	appID := c.Param("id")

	var req createVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		adminFail(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.BuildNumber == "" {
		req.BuildNumber = "1"
	}

	certID, err := ctrl.store.ActiveCertID(c.Request.Context(), t.ID)
	if err != nil || certID == "" {
		adminFail(c, http.StatusPreconditionFailed, "no active signing certificate configured for this tenant")
		return
	}

	versionID, jobID, err := ctrl.store.CreateVersionWithJob(c.Request.Context(), t.ID, appID, req.Version, req.BuildNumber, req.RawIPAS3Key, req.SizeBytes, certID)
	if err != nil {
		adminFail(c, http.StatusInternalServerError, err.Error())
		return
	}

	adminOK(c, gin.H{"version_id": versionID, "signing_job_id": jobID, "review_status": "pending_review"})
}
