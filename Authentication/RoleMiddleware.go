package authentication

import (
	"net/http"

	config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	"github.com/gin-gonic/gin"
)

func CheckRole(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID not found in context"})
			c.Abort()
			return
		}

		var user models.User
		err := config.DB.Preload("Role").Preload("Role.Permissions").Preload("Permissions").First(&user, userID)
		if err.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		if user.HasPermission(permission) {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
		c.Abort()
	}
}
