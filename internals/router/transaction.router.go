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
	transactionRouter := router.Group("/transaction")

	txRepo := repository.NewTransactionRepo(db)
	txServ := service.NewTransactionService(txRepo)
	txCont := controller.NewTransactionController(txServ)

	// GET /transaction/summary      → balance, income, expense summary
	transactionRouter.GET("/summary", middleware.VerifyToken, txCont.GetSummary)

	// GET /transaction/report?range=7days|30days  → chart data
	transactionRouter.GET("/report", middleware.VerifyToken, txCont.GetTransactionReport)

	transactionRouter.POST("/", middleware.VerifyToken, txCont.CreateTransaction)
	transactionRouter.GET("/", middleware.VerifyToken, txCont.GetTransactions)
	transactionRouter.GET("/:id", middleware.VerifyToken, txCont.GetTransactionByID)
}
