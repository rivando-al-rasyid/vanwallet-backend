package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/controller"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/middleware"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/repository"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

func ProfileRouter(router *gin.Engine, db *pgxpool.Pool) {
	profileRouter := router.Group("/profile")

	profRepo := repository.NewProfileRepo(db)
	profServ := service.NewProfileService(profRepo)
	profCont := controller.NewProfileController(profServ)

	profileRouter.GET("/", middleware.VerifyToken, profCont.GetProfile)
}
