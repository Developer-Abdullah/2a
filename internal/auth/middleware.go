package auth

import (
	"crypto/ecdsa"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strings"
)

const ContextClaimsKey = "jwt_claims"

func RequireAuth(publicKey *ecdsa.PublicKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authorization missing"})
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid format"})
			return
		}
		claims, err := ValidateToken(parts[1], publicKey)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}
		reqDeviceID := c.GetHeader("X-Device-ID")
		if reqDeviceID == "" && claims.DeviceID != "admin-dashboard" {
			c.AbortWithStatusJSON(400, gin.H{"error": "X-Device-ID missing"})
			return
		}
		if reqDeviceID != claims.DeviceID && claims.DeviceID != "admin-dashboard" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Device mismatch"})
			return
		}
		c.Set(ContextClaimsKey, claims)
		log.Ctx(c.Request.Context()).UpdateContext(func(ctx zerolog.Context) zerolog.Context { return ctx.Str("user_id", claims.UserID) })
		c.Next()
	}
}
func GetClaims(c *gin.Context) *CustomClaims {
	val, ok := c.Get(ContextClaimsKey)
	if !ok {
		return nil
	}
	return val.(*CustomClaims)
}

// RequirePlatformOwner gates a route to the platform owner: a super_admin whose token carries no
// tenant binding. It must run after RequireAuth. This centralizes the owner check that the platform
// controllers also enforce per-handler, giving defense in depth on the whole owner route group.
func RequirePlatformOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil || claims.Role != "super_admin" || claims.TenantID != "" {
			c.AbortWithStatusJSON(403, gin.H{"success": false, "error": gin.H{"message": "platform owner access required"}})
			return
		}
		c.Next()
	}
}
