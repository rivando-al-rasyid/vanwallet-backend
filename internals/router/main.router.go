package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/middleware"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

func MainRouter(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client, mail *service.MailService, midtrans *service.MidtransService) {
	router.Use(middleware.CORSMiddleware)
	router.Static("/img", "public/img")

	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/readyz", func(ctx *gin.Context) {
		if err := db.Ping(ctx.Request.Context()); err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "database": "down"})
			return
		}

		redisStatus := "disabled"
		if rdb != nil {
			if err := rdb.Ping(ctx.Request.Context()).Err(); err != nil {
				ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "database": "up", "redis": "down"})
				return
			}
			redisStatus = "up"
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "ready", "database": "up", "redis": redisStatus})
	})

	AuthRouter(router, db, rdb, mail)
	ProfileRouter(router, db, rdb)
	TransactionRouter(router, db, rdb)
	ReceiverRouter(router, db)
	TopupRouter(router, db, rdb, midtrans)
	WithdrawalRouter(router, db, rdb)
	TransferRouter(router, db, rdb)
	ExpenseRouter(router, db, rdb)
}
