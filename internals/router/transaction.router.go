package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/controller"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/middleware"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/repository"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

func TransactionRouter(router *gin.Engine, db *pgxpool.Pool) {
	txRepo := repository.NewTransactionRepo(db)
	txServ := service.NewTransactionService(txRepo)

	txCont := controller.NewTransactionController(txServ)

	g := router.Group("/transaction", middleware.VerifyTokenWithDB(db))

	g.GET("/receiver", txCont.FindReceivers)
	g.GET("/summary", txCont.GetSummary)
	g.GET("/report", txCont.GetTransactionReport)
	g.GET("/", txCont.GetTransactions)
	g.GET("/:id", txCont.GetTransactionByID)

	write := g.Group("", middleware.RequirePin(db))
	{
		write.POST("/topup", txCont.CreateTopup)
		write.PATCH("/topup/:id/confirm", txCont.ConfirmTopup)
		write.POST("/withdrawal", txCont.CreateWithdrawal)
		write.POST("/transfer", txCont.CreateTransfer)
		write.POST("/expense", txCont.CreateExpense)
	}
}
