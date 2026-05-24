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
	profRepo := repository.NewProfileRepo(db)
	profServ := service.NewProfileService(profRepo)
	profCont := controller.NewProfileController(profServ)

	profileRouter := router.Group("/profile", middleware.VerifyTokenWithDB(db))

	profileRouter.GET("/", profCont.GetProfile)
	profileRouter.POST("/", profCont.EditProfile)
	profileRouter.POST("/change/pin", profCont.EditPin)
	profileRouter.POST("/change/password", profCont.EditPassword)
}
