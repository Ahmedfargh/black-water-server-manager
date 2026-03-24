package functionalscontrollers

import (
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
