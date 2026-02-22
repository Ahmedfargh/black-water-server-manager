package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB is the global database instance
var DB *gorm.DB

// JwtSecret is the JWT secret key
var JwtSecret string

// ConnectDB initializes the database connection
func ConnectDB() {
	// 1. Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// 2. Read variables
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	JwtSecret = os.Getenv("JWT_SECRET")

	// 3. Construct DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// 4. Open Connection
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Database connection successfully opened")
	DB = database
}

// DropTables drops the specified database tables. Use with caution!
func DropTables(models ...interface{}) {
	fmt.Println("Dropping tables...")
	err := DB.Migrator().DropTable(models...)
	if err != nil {
		log.Printf("Error dropping tables: %v\n", err)
	} else {
		fmt.Println("Tables dropped successfully (if they existed).")
	}
}

func PortNumber() string {
	return os.Getenv("APP_PORT")
}
func GetKey(key string) string {
	return os.Getenv(key)
}
