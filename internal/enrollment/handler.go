package enrollment

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mozilla.org/pkcs7"
	"howett.net/plist"

	"platform/internal/tenant"
	"platform/pkg/mobileconfig"
)

const sessionTTL = 15 * time.Minute

type CertProvider interface {
	GetEnrollmentCert(ctx context.Context, tenantID uuid.UUID) ([]byte, []byte, string, error)
}

type Handler struct {
	repo         SessionRepository
	certProvider CertProvider
}

func NewHandler(repo SessionRepository, certProvider CertProvider) *Handler {
	return &Handler{repo: repo, certProvider: certProvider}
}

// HandleGetProfile issues a fresh enrollment session and returns the signed .mobileconfig that
// tells the device to POST its attributes (UDID, product, OS version) back to our callback.
func (h *Handler) HandleGetProfile(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant context missing"})
		return
	}

	token := randomToken()
	session := &Session{
		TenantID:     t.ID,
		OneTimeToken: token,
		IPAddress:    c.ClientIP(),
		ExpiresAt:    time.Now().Add(sessionTTL),
	}
	if err := h.repo.CreateSession(c.Request.Context(), t.SchemaName, session); err != nil {
		log.Ctx(c.Request.Context()).Error().Err(err).Msg("failed to create enrollment session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start enrollment"})
		return
	}

	callbackURL := baseURL(c) + "/v1/enroll/callback?token=" + token
	profile, err := h.buildProfile(c.Request.Context(), t.ID, t.Slug, callbackURL)
	if err != nil {
		log.Ctx(c.Request.Context()).Error().Err(err).Msg("failed to build enrollment profile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build enrollment profile"})
		return
	}

	c.Header("Content-Type", "application/x-apple-aspen-config")
	c.Header("Content-Disposition", `attachment; filename="enroll.mobileconfig"`)
	c.Writer.Write(profile)
}

// buildProfile signs the enrollment profile with the tenant's cert. If the tenant has no usable
// enrollment certificate configured yet, it falls back to an unsigned profile so the flow still
// works in development (iOS shows an "unsigned" warning but installation proceeds).
func (h *Handler) buildProfile(ctx context.Context, tenantID uuid.UUID, tenantName, callbackURL string) ([]byte, error) {
	certPEM, keyPEM, password, err := h.certProvider.GetEnrollmentCert(ctx, tenantID)
	if err == nil {
		if signed, serr := mobileconfig.GenerateAndSign(tenantName, callbackURL, certPEM, keyPEM, password); serr == nil {
			return signed, nil
		} else {
			log.Ctx(ctx).Warn().Err(serr).Msg("enrollment profile signing unavailable; serving unsigned profile")
		}
	} else {
		log.Ctx(ctx).Warn().Err(err).Msg("no enrollment certificate; serving unsigned profile")
	}
	return mobileconfig.BuildEnrollmentProfile(tenantName, callbackURL)
}

// deviceAttributes mirrors the keys Apple posts back in the signed profile-service response.
type deviceAttributes struct {
	UDID    string `plist:"UDID"`
	Product string `plist:"PRODUCT"`
	Version string `plist:"VERSION"`
}

// HandleAppleCallback receives the device's attributes after it installs the profile. Apple posts a
// PKCS#7-signed plist; we also accept a raw plist or a simple form post to support the web-form
// capture flow. The UDID is stored hashed — we never persist the raw identifier.
func (h *Handler) HandleAppleCallback(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant context missing"})
		return
	}
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing enrollment token"})
		return
	}

	attrs := h.parseCallback(c)
	if attrs.UDID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no device UDID in callback"})
		return
	}

	udidHash := sha256Hex(attrs.UDID)
	if err := h.repo.CompleteSession(c.Request.Context(), t.SchemaName, token, udidHash, deviceTypeFromProduct(attrs.Product)); err != nil {
		log.Ctx(c.Request.Context()).Warn().Err(err).Msg("failed to complete enrollment session")
		c.JSON(http.StatusConflict, gin.H{"error": "enrollment session expired or already completed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// parseCallback extracts the device attributes from a PKCS#7-signed plist, a raw plist, or form fields.
func (h *Handler) parseCallback(c *gin.Context) deviceAttributes {
	body, _ := io.ReadAll(c.Request.Body)

	if len(body) > 0 {
		payload := body
		if p7, err := pkcs7.Parse(body); err == nil && len(p7.Content) > 0 {
			payload = p7.Content
		}
		var attrs deviceAttributes
		if _, err := plist.Unmarshal(payload, &attrs); err == nil && attrs.UDID != "" {
			return attrs
		}
	}

	// Fallback: simple form/query capture flow.
	return deviceAttributes{
		UDID:    c.PostForm("udid"),
		Product: c.PostForm("product"),
		Version: c.PostForm("version"),
	}
}

// HandlePollStatus lets the enrollment web page poll for the device's callback.
func (h *Handler) HandlePollStatus(c *gin.Context) {
	t := tenant.GetFromContext(c)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant context missing"})
		return
	}
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing enrollment token"})
		return
	}
	session, err := h.repo.GetSessionByToken(c.Request.Context(), t.SchemaName, token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown enrollment token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"completed": session.Completed, "device_type": session.DeviceType})
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(strings.ToUpper(strings.TrimSpace(s))))
	return hex.EncodeToString(sum[:])
}

// deviceTypeFromProduct maps an Apple product identifier (e.g. "iPhone13,2", "iPad8,1") to the
// device_platform_type enum. Unknown products yield "" so the column stays NULL.
func deviceTypeFromProduct(product string) string {
	p := strings.ToLower(product)
	switch {
	case strings.HasPrefix(p, "iphone"):
		return "iphone"
	case strings.HasPrefix(p, "ipad"):
		return "ipad"
	default:
		return ""
	}
}

func baseURL(c *gin.Context) string {
	scheme := "https"
	if c.Request.TLS == nil && c.GetHeader("X-Forwarded-Proto") == "" {
		scheme = "http"
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	return scheme + "://" + c.Request.Host
}
