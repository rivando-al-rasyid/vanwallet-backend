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

func ProfileRouter(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client) {
	profRepo := repository.NewProfileRepo(db)
	profServ := service.NewProfileService(profRepo, rdb)
	profCont := controller.NewProfileController(profServ)

	profileRouter := router.Group("/profile", middleware.AuthRequired(db))

	// Full profile CRUD
	profileRouter.GET("/", profCont.GetProfile)
	profileRouter.PATCH("/edit", profCont.EditProfile)
	profileRouter.PATCH("/change/pin", profCont.EditPin)
	profileRouter.PATCH("/change/password", profCont.EditPassword)
}
