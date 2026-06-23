package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/controller"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/middleware"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/repository"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

func ReceiverRouter(router *gin.Engine, db *pgxpool.Pool) {
	receiverRepo := repository.NewTransactionRepo(db)
	receiverServ := service.NewReceiverService(receiverRepo)
	receiverCont := controller.NewReceiverController(receiverServ)

	g := router.Group("/transaction", middleware.AuthRequired(db))
	g.GET("/receiver", receiverCont.FindReceivers)
}
