package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/rivando-al-rasyid/vanwallet-backend/docs"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/middleware"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func MainRouter(router *gin.Engine, db *pgxpool.Pool) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.Use(middleware.CORSMiddleware)
	AuthRouter(router, db)
	ProfileRouter(router, db)
}
