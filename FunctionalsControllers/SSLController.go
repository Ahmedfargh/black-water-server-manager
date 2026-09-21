package functionalscontrollers

import (
	"net/http"
	"strconv"
	"strings"

	Config "github.com/ahmedfargh/server-manager/Config"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	sslservice "github.com/ahmedfargh/server-manager/Services/SSL"
	"github.com/gin-gonic/gin"
)

type InspectCertFileRequest struct {
	CertPath string `json:"cert_path" binding:"required"`
}

type IssueCertbotRequest struct {
	Domain     string `json:"domain" binding:"required"`
	Email      string `json:"email"`
	Webroot    string `json:"webroot"`
	Standalone bool   `json:"standalone"`
}

func InspectSSLHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		target := c.Query("domain")
		if target == "" {
			target = c.Query("target")
		}
		if target == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'domain' or 'target' is required"})
			return
		}

		svc := sslservice.NewSSLService()
		info, err := svc.InspectDomainSSL(target)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"target":      target,
			"certificate": info,
		})
	}
}

func InspectCertFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req InspectCertFileRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := sslservice.NewSSLService()
		info, err := svc.InspectCertFile(req.CertPath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"cert_path":   req.CertPath,
			"certificate": info,
		})
	}
}

func IssueCertbotHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req IssueCertbotRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		svc := sslservice.NewSSLService()
		output, err := svc.RequestCertbotCert(req.Domain, req.Email, req.Webroot, req.Standalone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  err.Error(),
				"output": output,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Let's Encrypt certificate generated/renewed successfully",
			"domain":  strings.TrimSpace(req.Domain),
			"output":  output,
		})
	}
}

func CheckSiteSSLHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site ID"})
			return
		}

		svc := sslservice.NewSSLService()
		info, err := svc.CheckSiteSSLAndUpdate(uint(id))
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "Failed to inspect SSL for site: " + err.Error(),
				"site_id": id,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Site SSL inspected and updated successfully",
			"site_id":     id,
			"certificate": info,
		})
	}
}

func ToggleSiteAutoRenewHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site ID"})
			return
		}

		var site models.Site
		if err := Config.DB.First(&site, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Site not found"})
			return
		}

		site.SSLAutoRenew = !site.SSLAutoRenew
		if err := Config.DB.Save(&site).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "Site SSL auto renew updated",
			"site_id":        site.ID,
			"ssl_auto_renew": site.SSLAutoRenew,
		})
	}
}
