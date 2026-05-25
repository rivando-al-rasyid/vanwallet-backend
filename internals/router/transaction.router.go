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

func TransactionRouter(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client) {
	txRepo := repository.NewTransactionRepo(db)
	txServ := service.NewTransactionService(txRepo)

	authRepo := repository.NewAuthRepo(db)
	authServ := service.NewAuthService(authRepo, rdb)

	txCont := controller.NewTransactionController(txServ, authServ)

	g := router.Group("/transaction", middleware.VerifyTokenWithDB(db))

	// Receiver search — used before initiating a transfer
	g.GET("/receiver", txCont.FindReceivers)

	// Read
	g.GET("/summary", txCont.GetSummary)
	g.GET("/report", txCont.GetTransactionReport)
	g.GET("/", txCont.GetTransactions)
	g.GET("/:id", txCont.GetTransactionByID)

	// Write
	g.POST("/topup", txCont.CreateTopup)
	g.PATCH("/topup/:id/confirm", txCont.ConfirmTopup)
	g.POST("/withdrawal", txCont.CreateWithdrawal)
	g.POST("/transfer", txCont.CreateTransfer)
	g.POST("/expense", txCont.CreateExpense)
}
