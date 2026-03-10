package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func JWTMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. Lấy token từ Header "Authorization: Bearer <token>" hoặc Fallback sang Cookie
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			var err error
			tokenString, err = c.Cookie("access_token")
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
				return
			}
		}


		// 2. Parse và Validate Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 3. Lấy UserID và Role từ Token và gán vào Context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userIDStr := claims["user_id"].(string)
			// Parse UUID để đảm bảo đúng định dạng
			userID, err := uuid.Parse(userIDStr)
			if err == nil {
				// QUAN TRỌNG: Gán ID thật vào context để các hàm sau dùng
				c.Set("userID", userID)
			}

			// Extract Role for RBAC
			if role, ok := claims["role"].(string); ok {
				c.Set("role", role)
			}
		}

		c.Next()
	}
}