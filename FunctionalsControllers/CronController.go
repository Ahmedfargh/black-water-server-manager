package functionalscontrollers

import (
	"encoding/json"
	"net/http"

	Config "github.com/ahmedfargh/server-manager/Config"
	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
	cron "github.com/ahmedfargh/server-manager/Services/Cron"
	"github.com/gin-gonic/gin"
)

type CronController struct {
	service   *cron.Service
	auditCRUD *crud.AuditLogCRUD
}

func NewCronController(service *cron.Service, auditCRUD *crud.AuditLogCRUD) *CronController {
	if service == nil {
		service = cron.NewService()
	}
	if auditCRUD == nil {
		auditCRUD = crud.NewAuditLogCRUD(repository.NewAuditRepository(Config.DB))
	}
	return &CronController{
		service:   service,
		auditCRUD: auditCRUD,
	}
}

// ListJobs godoc
// @Summary List all cron jobs with human-readable schedule translations
func (ctrl *CronController) ListJobs(c *gin.Context) {
	jobs, err := ctrl.service.ListJobs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to list crontab tasks",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   jobs,
	})
}

// CreateOrUpdateJob godoc
// @Summary Create or update a cron job
func (ctrl *CronController) CreateOrUpdateJob(c *gin.Context) {
	var req cron.CronMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body: expression and command are required",
			"error":   err.Error(),
		})
		return
	}

	var err error
	action := "CREATE"
	if req.ID != "" {
		action = "UPDATE"
		err = ctrl.service.UpdateJob(req.ID, req)
	} else {
		err = ctrl.service.AddJob(req)
	}

	// Record audit trail
	userID, _ := c.Get("userID")
	var uID *uint
	if id, ok := userID.(uint); ok {
		uID = &id
	} else if id, ok := userID.(float64); ok {
		uidVal := uint(id)
		uID = &uidVal
	}

	payloadJSON, _ := json.Marshal(map[string]interface{}{
		"expression": req.Expression,
		"command":    req.Command,
		"comment":    req.Comment,
		"enabled":    req.Enabled,
		"action":     action,
	})

	_ = ctrl.auditCRUD.CreateAudit(&models.AuditLog{
		UserID:      uID,
		ServiceType: "CRON",
		ServiceID:   req.ID,
		Action:      action,
		Results:     string(payloadJSON),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to save cron job",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cron job saved successfully",
	})
}

// DeleteJob godoc
// @Summary Remove a cron job by ID
func (ctrl *CronController) DeleteJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Cron job ID is required",
		})
		return
	}

	err := ctrl.service.DeleteJob(id)

	// Record audit trail
	userID, _ := c.Get("userID")
	var uID *uint
	if u, ok := userID.(uint); ok {
		uID = &u
	} else if u, ok := userID.(float64); ok {
		uidVal := uint(u)
		uID = &uidVal
	}

	_ = ctrl.auditCRUD.CreateAudit(&models.AuditLog{
		UserID:      uID,
		ServiceType: "CRON",
		ServiceID:   id,
		Action:      "DELETE",
		Results:     `{"id":"` + id + `"}`,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete cron job",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cron job deleted successfully",
	})
}

// ToggleJob godoc
// @Summary Enable or disable a cron job
func (ctrl *CronController) ToggleJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Cron job ID is required",
		})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Enabled boolean is required",
			"error":   err.Error(),
		})
		return
	}

	err := ctrl.service.ToggleJob(id, req.Enabled)

	// Record audit trail
	userID, _ := c.Get("userID")
	var uID *uint
	if u, ok := userID.(uint); ok {
		uID = &u
	} else if u, ok := userID.(float64); ok {
		uidVal := uint(u)
		uID = &uidVal
	}

	_ = ctrl.auditCRUD.CreateAudit(&models.AuditLog{
		UserID:      uID,
		ServiceType: "CRON",
		ServiceID:   id,
		Action:      "TOGGLE",
		Results:     `{"id":"` + id + `","enabled":` + boolToString(req.Enabled) + `}`,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to toggle cron job state",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cron job toggled successfully",
	})
}

// RunJob godoc
// @Summary Manually trigger a cron task execution
func (ctrl *CronController) RunJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Cron job ID is required",
		})
		return
	}

	res, err := ctrl.service.RunJob(id)

	// Record audit trail
	userID, _ := c.Get("userID")
	var uID *uint
	if u, ok := userID.(uint); ok {
		uID = &u
	} else if u, ok := userID.(float64); ok {
		uidVal := uint(u)
		uID = &uidVal
	}

	payloadJSON, _ := json.Marshal(map[string]interface{}{
		"id":          id,
		"command":     res.Command,
		"success":     res.Success,
		"duration_ms": res.DurationMs,
	})

	_ = ctrl.auditCRUD.CreateAudit(&models.AuditLog{
		UserID:      uID,
		ServiceType: "CRON",
		ServiceID:   id,
		Action:      "MANUAL_RUN",
		Results:     string(payloadJSON),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Cron job execution encountered an error",
			"error":   err.Error(),
			"data":    res,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cron job executed successfully",
		"data":    res,
	})
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
