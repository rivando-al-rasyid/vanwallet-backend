package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/middleware"
)

func MainRouter(router *gin.Engine, db *pgxpool.Pool) {
	router.Use(middleware.CORSMiddleware)
	AuthRouter(router, db)
	ProfileRouter(router, db)
}
