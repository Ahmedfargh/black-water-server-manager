package authentication

import (
	"fmt"
	"net/http"
	"strings"

	"errors"

	service "github.com/ahmedfargh/server-manager/Authentication/Service" // Assuming Claims is defined here
	config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func CheckRole(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format: Bearer <token>"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &service.Claims{} // Use Claims from AuthService

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.JwtSecret), nil
		})
		if err != nil {
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token signature"})
				c.Abort()
				return
			}
		}
		if token.Valid {
			var user models.User
			err := config.DB.Preload("Role").Preload("Role.Permissions").Find(&user, claims.UserID)
			if err.Error != nil {
				c.JSON(404, gin.H{"error": "User not found"})
				c.Abort()
				return
			}
			if user.HasPermission(permission) {
				c.Next()
				return
			}
		}
		c.JSON(403, gin.H{"error": "Forbidden"})
	}
}
