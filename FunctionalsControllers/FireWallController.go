package functionalscontrollers

import (
	"github.com/ahmedfargh/server-manager/Services"
	"github.com/gin-gonic/gin"
)

func EnableFireWallHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		firewall := Services.NewFirewall()
		text, error := firewall.Enable()
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
		text, error := firewall.Disable()
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
		text, error := firewall.Status()
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
