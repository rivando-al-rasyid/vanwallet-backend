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

func TopupRouter(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client, midtransService *service.MidtransService) {
	topupRepo := repository.NewTransactionRepo(db)
	topupServ := service.NewTopupService(topupRepo, rdb, midtransService)
	topupCont := controller.NewTopupController(topupServ)
	webhookCont := controller.NewMidtransWebhookController(topupServ)

	g := router.Group("/transaction", middleware.AuthRequired(db))
	g.POST("/topup", topupCont.CreateTopup)

	router.POST("/webhooks/midtrans", webhookCont.HandleNotification)
}
