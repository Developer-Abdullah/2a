package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/signing"
	"platform/internal/tenant"
	"platform/pkg/response"
	"platform/pkg/storage"
)

type InstallRepo interface {
	GetSignedVersion(ctx context.Context, tenantID uuid.UUID, versionID string) (*domain.InstallTarget, error)
	GetLatestSignedVersionID(ctx context.Context, tenantID uuid.UUID, appID string) (string, error)
}

type InstallController struct {
	repo InstallRepo
	s3   *storage.S3Client
}

func NewInstallController(repo InstallRepo, s3 *storage.S3Client) *InstallController {
	return &InstallController{repo: repo, s3: s3}
}

// InstallInfo (auth) hands the iOS client the manifest URL it feeds to itms-services://.
func (ctrl *InstallController) InstallInfo(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	if currentTenant == nil {
		response.Fail(c, http.StatusBadRequest, "tenant_missing", "tenant context missing")
		return
	}

	versionID, err := ctrl.repo.GetLatestSignedVersionID(c.Request.Context(), currentTenant.ID, c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusNotFound, "no_signed_version", "no signed version available for this app")
		return
	}

	response.Success(c, gin.H{
		"manifest_url": absoluteBaseURL(c) + "/v1/install/" + versionID + "/manifest.plist",
	})
}

// Manifest (public) serves the Apple OTA plist. The device fetches this directly with no auth
// headers, so the route is unauthenticated; the embedded download URL points back at this API.
func (ctrl *InstallController) Manifest(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	if currentTenant == nil {
		c.String(http.StatusBadRequest, "tenant context missing")
		return
	}

	versionID := c.Param("version_id")
	target, err := ctrl.repo.GetSignedVersion(c.Request.Context(), currentTenant.ID, versionID)
	if err != nil {
		c.String(http.StatusNotFound, "version not found or not signed")
		return
	}

	downloadURL := absoluteBaseURL(c) + "/v1/install/" + versionID + "/download"
	iconURL := ""
	if target.IconKey != "" {
		if url, err := ctrl.s3.GeneratePresignedURL(c.Request.Context(), target.IconKey, 1*time.Hour); err == nil {
			iconURL = url
		}
	}

	plist, err := signing.GenerateManifestPlist(target.AppName, target.BundleID, target.Version, downloadURL, iconURL, iconURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to render manifest")
		return
	}

	c.Data(http.StatusOK, "application/xml; charset=utf-8", plist)
}

// Download (public) redirects to a short-lived presigned URL for the signed IPA.
func (ctrl *InstallController) Download(c *gin.Context) {
	currentTenant := tenant.GetFromContext(c)
	if currentTenant == nil {
		c.String(http.StatusBadRequest, "tenant context missing")
		return
	}

	versionID := c.Param("version_id")
	target, err := ctrl.repo.GetSignedVersion(c.Request.Context(), currentTenant.ID, versionID)
	if err != nil || target.SignedKey == "" {
		c.String(http.StatusNotFound, "signed package not found")
		return
	}

	url, err := ctrl.s3.GeneratePresignedURL(c.Request.Context(), target.SignedKey, 15*time.Minute)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to presign download url")
		return
	}

	c.Redirect(http.StatusFound, url)
}

// absoluteBaseURL reconstructs this API's externally reachable origin for embedding in manifests.
func absoluteBaseURL(c *gin.Context) string {
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		return proto + "://" + c.Request.Host
	}
	scheme := "https"
	host := c.Request.Host
	if c.Request.TLS == nil && (strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") || strings.HasPrefix(host, "0.0.0.0")) {
		scheme = "http"
	}
	return scheme + "://" + host
}
