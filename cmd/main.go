package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/config"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/router"
)

// @title						Vanwallet
// @version						1.0
// @description					Backend Vanwallet  using Gin

// @license.name				MIT

// @host						localhost:8080
// @BasePath					/

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load .env — tidak fatal jika tidak ada (untuk production compatibility)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// Connect DB dulu sebelum register router apapun
	db, err := config.ConnectPsql()
	if err != nil {
		log.Fatalf("DB connection error.\ncause: %s", err.Error())
	}
	defer db.Close()
	log.Println("DB Connected")

	// Inisialisasi Gin
	app := gin.Default()

	// Register semua router setelah DB siap
	router.MainRouter(app, db)

	// Run server dengan fallback default
	host := os.Getenv("APP_HOST")
	port := os.Getenv("APP_PORT")
	addr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Server running at %s", addr)

	if err := app.Run(addr); err != nil {
		log.Fatalf("Failed to start server.\ncause: %s", err.Error())
	}
}
