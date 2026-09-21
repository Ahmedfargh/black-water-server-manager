package functionalscontrollers

import (
	"encoding/json"
	"net/http"

	Config "github.com/ahmedfargh/server-manager/Config"
	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
	systemd "github.com/ahmedfargh/server-manager/Services/Systemd"
	"github.com/gin-gonic/gin"
)

type SystemdController struct {
	service   *systemd.Service
	auditCRUD *crud.AuditLogCRUD
}

func NewSystemdController(service *systemd.Service, auditCRUD *crud.AuditLogCRUD) *SystemdController {
	if service == nil {
		service = systemd.NewService()
	}
	if auditCRUD == nil {
		auditCRUD = crud.NewAuditLogCRUD(repository.NewAuditRepository(Config.DB))
	}
	return &SystemdController{
		service:   service,
		auditCRUD: auditCRUD,
	}
}

// GetOverview godoc
// @Summary Get systemd availability and units summary
func (ctrl *SystemdController) GetOverview(c *gin.Context) {
	unitType := c.DefaultQuery("type", "service")
	search := c.Query("search")

	available := ctrl.service.IsSystemdAvailable()
	if !available {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": systemd.SystemdOverview{
				Available: false,
				Units:     []systemd.UnitInfo{},
			},
		})
		return
	}

	units, err := ctrl.service.ListUnits(c.Request.Context(), unitType, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to list systemd units",
			"error":   err.Error(),
		})
		return
	}

	activeCount := 0
	failedCount := 0
	serviceCount := 0
	timerCount := 0
	socketCount := 0

	for _, u := range units {
		if u.Active == "active" {
			activeCount++
		} else if u.Active == "failed" {
			failedCount++
		}
		switch u.Type {
		case "service":
			serviceCount++
		case "timer":
			timerCount++
		case "socket":
			socketCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": systemd.SystemdOverview{
			Available:    true,
			TotalUnits:   len(units),
			ActiveUnits:  activeCount,
			FailedUnits:  failedCount,
			ServiceUnits: serviceCount,
			TimerUnits:   timerCount,
			SocketUnits:  socketCount,
			Units:        units,
		},
	})
}

// GetUnitStatus godoc
// @Summary Get detailed status for a unit
func (ctrl *SystemdController) GetUnitStatus(c *gin.Context) {
	unitName := c.Param("unit")
	if unitName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Unit name is required",
		})
		return
	}

	status, err := ctrl.service.GetUnitStatus(c.Request.Context(), unitName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get unit status",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"unit":   unitName,
			"status": status,
		},
	})
}

// ExecuteAction godoc
// @Summary Perform lifecycle action (start, stop, restart, reload, enable, disable) on a unit
func (ctrl *SystemdController) ExecuteAction(c *gin.Context) {
	unitName := c.Param("unit")
	if unitName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Unit name is required",
		})
		return
	}

	var req systemd.UnitActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Action parameter is required in request body",
			"error":   err.Error(),
		})
		return
	}

	res, err := ctrl.service.ExecuteUnitAction(c.Request.Context(), unitName, req.Action)

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
		"unit":        unitName,
		"action":      req.Action,
		"success":     res.Success,
		"duration_ms": res.DurationMs,
		"output":      res.Output,
	})

	_ = ctrl.auditCRUD.CreateAudit(&models.AuditLog{
		UserID:      uID,
		ServiceType: "SYSTEMD",
		ServiceID:   unitName,
		Action:      req.Action,
		Results:     string(payloadJSON),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to execute unit action",
			"error":   err.Error(),
			"data":    res,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Unit action executed successfully",
		"data":    res,
	})
}
