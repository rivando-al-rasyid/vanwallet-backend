package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/controller"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/middleware"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/repository"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

func PinRouter(r *gin.Engine, db *pgxpool.Pool) {
	pinRepo := repository.NewPinRepo(db)
	pinSvc := service.NewPinService(pinRepo)
	pinCtrl := controller.NewPinController(pinSvc)

	pinGroup := r.Group("/pin")
	pinGroup.Use(middleware.VerifyTokenWithDB(db))
	{
		pinGroup.POST("", pinCtrl.SetPin)
		pinGroup.GET("/status", pinCtrl.CheckPin)
		pinGroup.POST("/verify", pinCtrl.VerifyPin)
	}
}
