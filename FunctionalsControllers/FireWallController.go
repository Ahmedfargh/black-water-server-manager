package functionalscontrollers

import (
	"fmt"

	"github.com/ahmedfargh/server-manager/Services"
	"github.com/gin-gonic/gin"
)

func EnableFireWallHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		firewall := Services.NewFirewall()
		user_id := c.GetInt("userID")
		fmt.Println(user_id)
		text, error := firewall.Enable(user_id)

		if error != nil {
			c.JSON(500, gin.H{"error": error.Error()})
			return
		}
		c.JSON(200, gin.H{"message": text})
	}
}
func DisableFireWallHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		firewall := Services.NewFirewall()
		fmt.Println(c.GetInt("userID"))
		text, error := firewall.Disable(c.GetInt("userID"))
		if error != nil {
			c.JSON(500, gin.H{"error": error.Error()})
			return
		}
		c.JSON(200, gin.H{"message": text})
	}
}
func StatusFireWallHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		firewall := Services.NewFirewall()
		text, error := firewall.Status(c.GetInt("userID"))
		if error != nil {
			c.JSON(500, gin.H{"error": error.Error()})
			return
		}
		c.JSON(200, gin.H{"message": text})
	}
}
func RulesFireWallHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		firewall := Services.NewFirewall()
		text, error := firewall.Rules()
		if error != nil {
			c.JSON(500, gin.H{"error": error.Error()})
			return
		}
		c.JSON(200, gin.H{"message": text})
	}
}
func ListRulesFireWallHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		firewall := Services.NewFirewall()
		text, error := firewall.ListRules()
		if error != nil {
			c.JSON(500, gin.H{"error": error.Error()})
			return
		}
		c.JSON(200, gin.H{"message": text})
	}
}
