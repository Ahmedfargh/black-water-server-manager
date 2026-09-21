package functionalscontrollers

import (
	"net/http"
	"strconv"

	nginxService "github.com/ahmedfargh/server-manager/Services/Nginx"
	"github.com/gin-gonic/gin"
)

type NginxController struct {
	Service *nginxService.NginxService
}

func NewNginxController(s *nginxService.NginxService) *NginxController {
	if s == nil {
		s = nginxService.NewNginxService()
	}
	return &NginxController{Service: s}
}

// GetOverview handles GET /nginx/overview
func (ctrl *NginxController) GetOverview(c *gin.Context) {
	overview := ctrl.Service.GetOverview()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    overview,
	})
}

// GetSites handles GET /nginx/sites
func (ctrl *NginxController) GetSites(c *gin.Context) {
	includeContent := c.DefaultQuery("include_content", "false") == "true"
	sites, err := ctrl.Service.ListSites(includeContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sites,
		"total":   len(sites),
	})
}

// GetSite handles GET /nginx/sites/:name
func (ctrl *NginxController) GetSite(c *gin.Context) {
	name := c.Param("name")
	site, err := ctrl.Service.GetSite(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    site,
	})
}

type SaveSiteRequest struct {
	Filename string `json:"filename" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// SaveSite handles POST /nginx/sites
func (ctrl *NginxController) SaveSite(c *gin.Context) {
	var req SaveSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := ctrl.Service.SaveSite(req.Filename, req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Site configuration saved and validated successfully",
	})
}

type ToggleSiteRequest struct {
	Enable bool `json:"enable"`
}

// ToggleSite handles POST /nginx/sites/:name/toggle
func (ctrl *NginxController) ToggleSite(c *gin.Context) {
	name := c.Param("name")
	var req ToggleSiteRequest
	_ = c.ShouldBindJSON(&req)

	if err := ctrl.Service.ToggleSite(name, req.Enable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	statusStr := "disabled"
	if req.Enable {
		statusStr = "enabled"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Site " + name + " " + statusStr + " successfully",
	})
}

// DeleteSite handles DELETE /nginx/sites/:name
func (ctrl *NginxController) DeleteSite(c *gin.Context) {
	name := c.Param("name")
	if err := ctrl.Service.DeleteSite(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Site configuration deleted successfully",
	})
}

// TestConfig handles POST /nginx/test
func (ctrl *NginxController) TestConfig(c *gin.Context) {
	valid, output := ctrl.Service.TestConfiguration()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"valid":   valid,
		"output":  output,
	})
}

// Reload handles POST /nginx/reload
func (ctrl *NginxController) Reload(c *gin.Context) {
	if err := ctrl.Service.Reload(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Nginx service reloaded successfully",
	})
}

// Restart handles POST /nginx/restart
func (ctrl *NginxController) Restart(c *gin.Context) {
	if err := ctrl.Service.Restart(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Nginx service restarted successfully",
	})
}

// GetLogFiles handles GET /nginx/logs/files
func (ctrl *NginxController) GetLogFiles(c *gin.Context) {
	accessLogs, errorLogs := ctrl.Service.DiscoverLogFiles()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_logs": accessLogs,
			"error_logs":  errorLogs,
		},
	})
}

// GetAccessLogs handles GET /nginx/logs/access
func (ctrl *NginxController) GetAccessLogs(c *gin.Context) {
	file := c.DefaultQuery("file", "")
	limitStr := c.DefaultQuery("limit", "100")
	statusStr := c.DefaultQuery("status", "0")
	search := c.DefaultQuery("search", "")

	limit, _ := strconv.Atoi(limitStr)
	status, _ := strconv.Atoi(statusStr)

	entries, err := ctrl.Service.ParseAccessLogs(file, limit, status, search)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    entries,
		"count":   len(entries),
	})
}

// GetErrorLogs handles GET /nginx/logs/error
func (ctrl *NginxController) GetErrorLogs(c *gin.Context) {
	file := c.DefaultQuery("file", "")
	limitStr := c.DefaultQuery("limit", "100")
	level := c.DefaultQuery("level", "")
	search := c.DefaultQuery("search", "")

	limit, _ := strconv.Atoi(limitStr)

	entries, err := ctrl.Service.ParseErrorLogs(file, limit, level, search)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    entries,
		"count":   len(entries),
	})
}

// GetLogAnalytics handles GET /nginx/logs/analytics
func (ctrl *NginxController) GetLogAnalytics(c *gin.Context) {
	file := c.DefaultQuery("file", "")
	analytics, err := ctrl.Service.GetLogAnalytics(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}
