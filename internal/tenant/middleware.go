package tenant

import (
	"context"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"platform/internal/domain"
)

const ContextKey = "current_tenant"

type TenantRepository interface {
	GetBySlugOrID(ctx context.Context, identifier string) (*domain.Tenant, error)
}

func ResolverMiddleware(repo TenantRepository) gin.HandlerFunc {
	// Single-store mode: when STORE_TENANT_SLUG is set, every request resolves to that one store,
	// regardless of Host/subdomain or headers. This is required so header-less callers — payment
	// webhooks and the device's itms-services OTA fetch — and subdomain APIs (api.example.com) all
	// map to the store instead of being misread as an "api" tenant.
	forced := strings.TrimSpace(os.Getenv("STORE_TENANT_SLUG"))
	return func(c *gin.Context) {
		identifier := forced
		if identifier == "" {
			identifier = ResolveSlug(c.Request)
		}
		if identifier == "" {
			c.AbortWithStatusJSON(400, gin.H{"error": "Tenant identifier missing"})
			return
		}
		t, err := repo.GetBySlugOrID(c.Request.Context(), identifier)
		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": "Tenant not found"})
			return
		}
		if t.Status != domain.TenantStatusActive {
			c.AbortWithStatusJSON(403, gin.H{"error": "Tenant inactive"})
			return
		}
		c.Set(ContextKey, t)
		log.Ctx(c.Request.Context()).UpdateContext(func(cx zerolog.Context) zerolog.Context { return cx.Str("tenant_id", t.ID.String()) })
		c.Next()
	}
}
func GetFromContext(c *gin.Context) *domain.Tenant {
	val, ok := c.Get(ContextKey)
	if !ok {
		return nil
	}
	return val.(*domain.Tenant)
}
