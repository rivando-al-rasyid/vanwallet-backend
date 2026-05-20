package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/controller"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/middleware"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/repository"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

func DashboardRouter(router *gin.Engine, db *pgxpool.Pool) {
	dashboardRouter := router.Group("/dashboard")

	dashRepo := repository.NewDashboardRepo(db)
	dashServ := service.NewDashboardService(dashRepo)
	dashCont := controller.NewDashboardController(dashServ)

	dashboardRouter.GET("/", middleware.VerifyToken, dashCont.GetData)

}
