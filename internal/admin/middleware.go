package admin

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"platform/internal/auth"
)

func EnforceAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := auth.GetClaims(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if claims.Role != "admin" && claims.Role != "super_admin" {
			log.Warn().Str("user_id", claims.UserID).Str("role", claims.Role).Msg("Blocked unauthorized admin access attempt")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "forbidden",
					"message": "You do not have administrative privileges.",
				},
			})
			return
		}
		c.Next()
	}
}