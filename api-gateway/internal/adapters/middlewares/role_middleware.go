package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole returns a Gin middleware that restricts access to users whose
// "role" (set in context by JWTMiddleware) matches one of the allowedRoles.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: No role assigned"})
			return
		}

		userRole, ok := roleVal.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Invalid role format"})
			return
		}

		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Insufficient permissions"})
	}
}
