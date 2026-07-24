package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/config"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/router"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

// @title                       Vanwallet
// @version                     1.0
// @description                 Backend Vanwallet using Gin

// @license.name                MIT

// @host                        localhost:8080
// @BasePath                    /

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" followed by a space and your JWT. Example: "Bearer eyJhbGci..."
func main() {
	app := gin.Default()

	db, err := config.ConnectPsql()
	if err != nil {
		log.Fatalf("DB connection error. \ncause: %s", err.Error())
	}
	defer db.Close()

	log.Println("DB Connected")

	rc, err := config.ConnectRedis()
	if err != nil {
		log.Printf("Redis disabled: %s", err.Error())
	} else {
		defer rc.Close()
		log.Println("Redis Connected")
	}

	// Initialize mail service
	mailService, err := service.NewMailService()
	if err != nil {
		log.Fatalf("Mail service initialization error: %s", err.Error())
	}

	log.Println("Mail Service Connected")

	midtransCfg, err := config.MidtransConfigFromEnv()
	if err != nil {
		log.Fatalf("Midtrans config error: %s", err.Error())
	}
	midtransService := service.NewMidtransService(midtransCfg)

	log.Println("Midtrans Service Connected")

	router.MainRouter(
		app,
		db,
		rc,
		mailService,
		midtransService,
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serverAddr := fmt.Sprintf("0.0.0.0:%s", port)

	if err := app.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server %v", err)
	}
}
