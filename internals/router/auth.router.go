package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/controller"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/middleware"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/repository"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

func AuthRouter(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client) {
	authRepo := repository.NewAuthRepo(db)
	authServ := service.NewAuthService(authRepo, rdb)
	authCont := controller.NewAuthController(authServ)

	auth := router.Group("/auth")

	// Public
	auth.POST("/register", authCont.Register)
	auth.POST("/login", authCont.Login)

	// Protected
	protected := auth.Group("/", middleware.VerifyTokenWithDB(db))
	protected.POST("/logout", authCont.Logout)
	protected.GET("/pin", authCont.GetPIN)
	protected.POST("/pin/verify", authCont.VerifyPIN)
}
