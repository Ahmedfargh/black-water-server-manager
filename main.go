package main

import (
	context "context"

	service "github.com/ahmedfargh/server-manager/Authentication/Service"
	config "github.com/ahmedfargh/server-manager/Config"
	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
	Mgrs "github.com/ahmedfargh/server-manager/Managers"
	routes "github.com/ahmedfargh/server-manager/Routes"
	"github.com/gin-gonic/gin"

	"log"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func main() {
	// start measurements
	startTime := time.Now()
	numCPU := runtime.NumCPU()
	cpuStartVals, _ := cpu.Percent(200*time.Millisecond, false)
	var cpuStart float64
	if len(cpuStartVals) > 0 {
		cpuStart = cpuStartVals[0]
	}
	memStart, _ := mem.VirtualMemory()

	log.Printf("Startup sample - CPU cores: %d, CPU%%: %.2f, RAM%%: %.2f, RAM used: %.2fMB", numCPU, cpuStart, memStart.UsedPercent, float64(memStart.Used)/1024/1024)

	dockerMgr := Mgrs.GetDockerManager()
	dockerMgr.DiscoverContainers(context.Background())
	config.ConnectDB()
	config.DB.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.Process{}, &models.AuditLog{}, &models.Site{}, &models.SiteHealthStatus{}, &models.Docker{})
	userRepo := repository.NewUserRepository(config.DB)
	userCRUD := crud.NewUserCRUD(userRepo)
	roleCRUD := crud.NewRoleCRUD(config.DB)
	// permissionCRUD := crud.NewPermissionCRUD(config.DB)
	authService := service.NewAuthService(userCRUD, roleCRUD)

	router := gin.Default()
	routes.AuthRoutes(router, userCRUD, authService, roleCRUD)

	routes.CpuRoute(router)
	routes.ProcessRoute(router)
	routes.RegisterRealTimeRoutes(router)
	routes.NetworkRoutes(router)
	routes.FireWallRoute(router)
	routes.SetupDockerRoutes(router)
	routes.AuditRoutes(router)
	routes.SiteRoutes(router)
	router.Static("/uploads", "./uploads")

	endTime := time.Now()
	cpuEndVals, _ := cpu.Percent(200*time.Millisecond, false)
	var cpuEnd float64
	if len(cpuEndVals) > 0 {
		cpuEnd = cpuEndVals[0]
	}
	memEnd, _ := mem.VirtualMemory()
	duration := endTime.Sub(startTime)

	log.Printf("Boot complete - duration: %s", duration)
	log.Printf("Boot sample - CPU%% start: %.2f -> end: %.2f, RAM%% start: %.2f -> end: %.2f", cpuStart, cpuEnd, memStart.UsedPercent, memEnd.UsedPercent)

	if err := router.Run(config.PortNumber()); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
