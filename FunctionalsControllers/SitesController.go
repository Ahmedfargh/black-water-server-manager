package functionalscontrollers

import (
	"fmt"
	"net/http"
	"strconv"

	Config "github.com/ahmedfargh/server-manager/Config"
	service "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repo "github.com/ahmedfargh/server-manager/Database/Repository"
	functional_service "github.com/ahmedfargh/server-manager/Services"
	"github.com/gin-gonic/gin"
)

func createNewSiteCrud() *service.SiteCrud {
	siteCrud := service.SiteCrud{Rep: repo.NewSiteRepository(Config.DB)}
	return &siteCrud
}
func CreateSiteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var site models.Site
		if err := c.ShouldBindJSON(&site); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := createNewSiteCrud().CreateSite(site); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		results, err := http.Get(site.Health_Route)
		var health_message string
		if err != nil {
			if results.StatusCode >= 400 && results.StatusCode < 500 {
				health_message = "the site " + site.Name + " is not found"
			} else if results.StatusCode >= 500 {
				health_message = "the site " + site.Name + " is down"
			} else {
				health_message = "the site " + site.Name + " is health"
			}
		} else {
			if results.StatusCode >= 400 && results.StatusCode < 500 {
				health_message = "the site " + site.Name + " is not found"
			} else if results.StatusCode >= 500 {
				health_message = "the site " + site.Name + " is down"
			} else {
				health_message = "the site " + site.Name + " is health"
			}
		}
		c.JSON(http.StatusOK, gin.H{"message": "Site created successfully", "health_message": health_message})
	}
}
func GetSitesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		page_str := c.DefaultQuery("page", "1")
		limit_str := c.DefaultQuery("limit", "10")
		page, err := strconv.Atoi(page_str)
		if err != nil || page <= 0 {
			page = 1
		}
		limit, err := strconv.Atoi(limit_str)
		if err != nil || limit <= 0 {
			limit = 10
		}
		sites, total, err := createNewSiteCrud().GetSites(page, limit)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"sites": sites, "total": total, "page": page, "limit": limit})
	}
}
func GetFullSitesCheckUpHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		service := functional_service.NewSiteHealthService()
		results, err := service.RunCheckUp()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"results": results})
	}
}
func GetSiteHealthStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		site_id, err := strconv.Atoi(c.Param("site_id"))
		page_str := c.DefaultQuery("page", "1")
		limit_str := c.DefaultQuery("limit", "10")
		page, err := strconv.Atoi(page_str)
		if err != nil || page <= 0 {
			page = 1
		}
		limit, err := strconv.Atoi(limit_str)
		if err != nil || limit <= 0 {
			limit = 10
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site ID"})
			return
		}
		service := functional_service.NewSiteHealthService()
		results, total, err := service.GetSiteHealthStatus(uint(site_id), page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		c.JSON(http.StatusOK, gin.H{"results": results, "total": total, "page": page, "limit": limit})
	}
}
func GetSiteStatusReportHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		site_id, err := strconv.Atoi(c.Param("site_id"))
		start_date := c.Query("start_date")
		end_date := c.Query("end_date")
		if err != nil {
		}
		service := functional_service.NewSiteHealthService()
		report, err := service.GetSiteStatusReport(uint(site_id), start_date, end_date)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusOK, gin.H{"error": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"report": report})
	}
}
func UpdateSiteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid site ID"})
			return
		}
		var site models.Site
		if err := c.ShouldBindJSON(&site); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := createNewSiteCrud().UpdateSite(site, uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Site updated successfully"})
	}
}
