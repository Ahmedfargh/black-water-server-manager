package functionalscontrollers

import (
	"net/http"
	"strconv"

	"github.com/ahmedfargh/server-manager/Services/PackageManagers"
	"github.com/gin-gonic/gin"
)

type PackageManagerController struct {
	service *PackageManagers.PackageManagerService
}

func NewPackageManagerController(service *PackageManagers.PackageManagerService) *PackageManagerController {
	if service == nil {
		service = PackageManagers.NewPackageManagerService(nil, nil)
	}
	return &PackageManagerController{
		service: service,
	}
}

// GetOverview godoc
// @Summary Get system OS and detected package managers overview
func (ctrl *PackageManagerController) GetOverview(c *gin.Context) {
	refresh := c.Query("refresh") == "true"
	overview, err := ctrl.service.GetSystemOverview(c.Request.Context(), refresh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to collect package managers overview",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   overview,
	})
}

// GetUpdates godoc
// @Summary Get pending updates for a specific package manager
func (ctrl *PackageManagerController) GetUpdates(c *gin.Context) {
	manager := c.Param("manager")
	if manager == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Manager parameter is required",
		})
		return
	}

	updates, err := ctrl.service.GetManagerUpdates(c.Request.Context(), manager)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch updates",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"manager": manager,
		"count":   len(updates),
		"data":    updates,
	})
}

// GetPackages godoc
// @Summary Get installed packages for a specific package manager with search and pagination
func (ctrl *PackageManagerController) GetPackages(c *gin.Context) {
	manager := c.Param("manager")
	if manager == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Manager parameter is required",
		})
		return
	}

	query := c.Query("query")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	items, total, err := ctrl.service.GetInstalledPackages(c.Request.Context(), manager, query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch installed packages",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"manager": manager,
		"page":    page,
		"limit":   limit,
		"total":   total,
		"data":    items,
	})
}

// CleanCache godoc
// @Summary Clean package manager cache
func (ctrl *PackageManagerController) CleanCache(c *gin.Context) {
	manager := c.Param("manager")
	if manager == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Manager parameter is required",
		})
		return
	}

	result, err := ctrl.service.CleanCache(c.Request.Context(), manager)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to clean package manager cache",
			"error":   err.Error(),
			"data":    result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Package cache cleaned successfully",
		"data":    result,
	})
}

// RefreshRepositories godoc
// @Summary Refresh repository metadata
func (ctrl *PackageManagerController) RefreshRepositories(c *gin.Context) {
	manager := c.Param("manager")
	if manager == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Manager parameter is required",
		})
		return
	}

	result, err := ctrl.service.RefreshRepositories(c.Request.Context(), manager)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to refresh repository metadata",
			"error":   err.Error(),
			"data":    result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Repositories refreshed successfully",
		"data":    result,
	})
}

// UpgradeSystem godoc
// @Summary Perform system or package manager upgrade
func (ctrl *PackageManagerController) UpgradeSystem(c *gin.Context) {
	manager := c.Param("manager")
	if manager == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Manager parameter is required",
		})
		return
	}

	result, err := ctrl.service.UpgradeSystem(c.Request.Context(), manager)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "System upgrade failed",
			"error":   err.Error(),
			"data":    result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "System upgrade executed successfully",
		"data":    result,
	})
}

// InstallPackage godoc
// @Summary Install a package via package manager
func (ctrl *PackageManagerController) InstallPackage(c *gin.Context) {
	manager := c.Param("manager")
	if manager == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Manager parameter is required",
		})
		return
	}

	var req PackageManagers.PackageActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body: package name is required",
			"error":   err.Error(),
		})
		return
	}

	result, err := ctrl.service.InstallPackage(c.Request.Context(), manager, req.Package)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to install package",
			"error":   err.Error(),
			"data":    result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Package installed successfully",
		"data":    result,
	})
}

// RemovePackage godoc
// @Summary Remove/uninstall a package
func (ctrl *PackageManagerController) RemovePackage(c *gin.Context) {
	manager := c.Param("manager")
	if manager == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Manager parameter is required",
		})
		return
	}

	var req PackageManagers.PackageActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body: package name is required",
			"error":   err.Error(),
		})
		return
	}

	result, err := ctrl.service.RemovePackage(c.Request.Context(), manager, req.Package, req.Purge)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to remove package",
			"error":   err.Error(),
			"data":    result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Package removed successfully",
		"data":    result,
	})
}

// UpgradePackage godoc
// @Summary Upgrade a single package
func (ctrl *PackageManagerController) UpgradePackage(c *gin.Context) {
	manager := c.Param("manager")
	if manager == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Manager parameter is required",
		})
		return
	}

	var req PackageManagers.PackageActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body: package name is required",
			"error":   err.Error(),
		})
		return
	}

	result, err := ctrl.service.UpgradePackage(c.Request.Context(), manager, req.Package)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to upgrade package",
			"error":   err.Error(),
			"data":    result,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Package upgraded successfully",
		"data":    result,
	})
}
