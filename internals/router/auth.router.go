package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/controller"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/repository"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

func RegisterAuthRouter(router *gin.Engine, db *pgxpool.Pool) {
	authRouter := router.Group("/auth")

	authRepo := repository.NewAuthRepo(db)
	authServ := service.NewAuthService(authRepo)
	authCont := controller.NewAuthController(authServ)

	authRouter.POST("/register", authCont.Register)
	authRouter.POST("/login", authCont.Login)
}
