package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	service "github.com/ahmedfargh/server-manager/Authentication/Service"
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

func main() {
	_ = godotenv.Load()

	// 1. Initialize Notification Driver
	driver := &NotificationDrivers.TelegramNotificationDriver{
		BotToken: config.GetKey("TELEGRAM_BOT_TOKEN"),
		ChatID:   config.GetKey("MASTER_CHAT_ID"),
	}

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

	// 5. Start Server in background
	go func() {
		fmt.Println("🚀 Server starting on :8080...")
		if err := router.Run(":8080"); err != nil {
			fmt.Printf("Router shutdown: %v\n", err)
		}
	}()

	// 6. Notify Startup
	go driver.Send("System", "✅ BlackWater Instance is UP", nil)

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
	_, _ = driver.Send("System", fmt.Sprintf("⚠️ BlackWater is DOWN (Signal: %s)", sig), nil)

	// Final wait for network flush
	time.Sleep(2 * time.Second)
	fmt.Println("Cleanup complete. Goodbye!")
}
