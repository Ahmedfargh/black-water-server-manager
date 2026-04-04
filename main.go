package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	service "github.com/ahmedfargh/server-manager/Authentication/Service"
	reports "github.com/ahmedfargh/server-manager/BackGround/Reports"
	config "github.com/ahmedfargh/server-manager/Config"
	crud "github.com/ahmedfargh/server-manager/Database/CRUD"
	models "github.com/ahmedfargh/server-manager/Database/Models"
	repository "github.com/ahmedfargh/server-manager/Database/Repository"
	"github.com/ahmedfargh/server-manager/Drivers/NotificationDrivers"
	Mgrs "github.com/ahmedfargh/server-manager/Managers"

	routes "github.com/ahmedfargh/server-manager/Routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func InitBackgroundTasks(mgr *Mgrs.BackgroundTaskManager) {
	hardwareReport := &reports.BackgroundHardwareReport{
		RunEachSeconds: 1, // Run every 60 seconds
	}
	fmt.Println("Init Background Task")
	mgr.AddTask(hardwareReport)
}
func StartBackgroundTasks(mgr *Mgrs.BackgroundTaskManager) {
	fmt.Println("Start Background Task")
	mgr.RunAllTasks()
}
func main() {
	fmt.Println("Starting BlackWater Server Manager...")

	_ = godotenv.Load()
	fmt.Println("Starting BlackWater Server Manager...")
	// 1. Initialize Notification Driver
	driver := &NotificationDrivers.DiscordDriver{}
	fmt.Println("Driver Initated Basic notification")
	// 2. Initialize Core Services (DB & Monitoring)
	initDatabase()
	go Mgrs.GetDockerManager().StartMonitoring()

	// 3. Setup App Dependencies
	userRepo := repository.NewUserRepository(config.DB)
	userCRUD := crud.NewUserCRUD(userRepo)
	roleCRUD := crud.NewRoleCRUD(config.DB)
	authService := service.NewAuthService(userCRUD, roleCRUD)

	// 4. Setup Router
	router := setupRouter(userCRUD, authService, roleCRUD)
	fmt.Println("Router initialized")
	// 5. Start Server in background
	go func() {
		fmt.Println("🚀 Server starting on :8080...")
		if err := router.Run(":8080"); err != nil {
			fmt.Printf("Router shutdown: %v\n", err)
		}
	}()

	// 6. Notify Startup
	Discord_meta_Data := map[string]string{
		"title":       "BlackWater Server Manager",
		"description": "✅ BlackWater Instance is UP",
		"bot_token":   config.GetKey("DISCORD_BOT_TOKEN"),
		"channel_id":  config.GetKey("DISCORD_CHANNEL_ID"),
	}
	go driver.Send("System", "✅ BlackWater Instance is UP", Discord_meta_Data)

	background_task_mgr := Mgrs.BackgroundTaskManager{}
	InitBackgroundTasks(&background_task_mgr)
	go StartBackgroundTasks(&background_task_mgr)
	// 7. Wait for Shutdown Signal
	handleGracefulShutdown(driver)

}

// --- Helper Functions to keep main() clean ---

func initDatabase() {
	config.ConnectDB()
	config.DB.AutoMigrate(
		&models.User{}, &models.Role{}, &models.Permission{},
		&models.Process{}, &models.AuditLog{}, &models.Site{},
		&models.SiteHealthStatus{}, &models.Docker{},
	)
}

func setupRouter(u *crud.UserCRUD, a *service.AuthService, r *crud.RoleCRUD) *gin.Engine {
	router := gin.Default()

	// Grouping routes visually
	routes.AuthRoutes(router, u, a, r)
	routes.CpuRoute(router)
	routes.ProcessRoute(router)
	routes.RegisterRealTimeRoutes(router)
	routes.NetworkRoutes(router)
	routes.FireWallRoute(router)
	routes.SetupDockerRoutes(router)
	routes.AuditRoutes(router)
	routes.SiteRoutes(router)

	router.Static("/uploads", "./uploads")
	return router
}

func handleGracefulShutdown(driver NotificationDrivers.NotificationInterface) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	fmt.Printf("\n📢 Signal [%v] received. Shutting down...\n", sig)

	// Final notification
	Discord_meta_Data := map[string]string{
		"title":       "BlackWater Server Manager",
		"description": fmt.Sprintf("⚠️ BlackWater is DOWN (Signal: %s)", sig),
	}
	_, _ = driver.Send("System", "⚠️ BlackWater is DOWN", Discord_meta_Data)

	// Final wait for network flush
	time.Sleep(2 * time.Second)
	fmt.Println("Cleanup complete. Goodbye!")
}
