package functionalscontrollers

import (
	"fmt"
	"net/http"

	"github.com/ahmedfargh/server-manager/Services"
	"github.com/gin-gonic/gin"
)

type IPBlockRequest struct {
	IP string `json:"ip" binding:"required"`
}

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

func BlockIPHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req IPBlockRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid request payload. 'ip' field is required.",
				"error":   err.Error(),
			})
			return
		}

		firewall := Services.NewFirewall()
		userID := c.GetInt("userID")
		msg, err := firewall.BlockIP(req.IP, userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Failed to block IP",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": msg,
			"ip":      req.IP,
		})
	}
}

func UnblockIPHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req IPBlockRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid request payload. 'ip' field is required.",
				"error":   err.Error(),
			})
			return
		}

		firewall := Services.NewFirewall()
		userID := c.GetInt("userID")
		msg, err := firewall.UnblockIP(req.IP, userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Failed to unblock IP",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": msg,
			"ip":      req.IP,
		})
	}
}

