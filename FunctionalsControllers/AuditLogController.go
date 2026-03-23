package functionalscontrollers

import (
	"net/http"
	"strconv"

	Config "github.com/ahmedfargh/server-manager/Config"
	Crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	Repository "github.com/ahmedfargh/server-manager/Database/Repository"
	"github.com/gin-gonic/gin"
)

func getAuditCrudService() *Crud.AuditLogCRUD {
	return Crud.NewAuditLogCRUD(Repository.NewAuditRepository(Config.DB))
}

func GetAuditLogHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "10")
		Type := c.Query("type")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			page = 1
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 10
		}

		audits, err := getAuditCrudService().GetAudits(page, limit, Type)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  audits,
			"page":  page,
			"limit": limit,
		})
	}
}
