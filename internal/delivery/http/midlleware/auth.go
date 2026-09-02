package midlleware

import (
	"strings"

	"task-api/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, response.CodeInvalidCredential, "Missing authorization header")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, response.CodeInvalidCredential, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, response.CodeInvalidCredential, "Invalid or expired token")
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, ok := claims["user_id"].(string)
			if !ok {
				response.Unauthorized(c, response.CodeInvalidCredential, "Invalid token claims")
				c.Abort()
				return
			}
			c.Set("user_id", userID)
			c.Next()
		} else {
			response.Unauthorized(c, response.CodeInvalidCredential, "Invalid token")
			c.Abort()
		}
	}
}
