package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

type TransactionController struct {
	transactionService *service.TransactionService
}

func NewTransactionController(transactionService *service.TransactionService) *TransactionController {
	return &TransactionController{transactionService: transactionService}
}

// GetSummary godoc
//
//	@Summary		Get transaction summary
//	@Description	Returns the authenticated user's current balance, total income, and total expense
//	@Tags			Transaction
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	dto.Response{data=dto.SummaryResponse}
//	@Failure		401	{object}	dto.Response
//	@Failure		404	{object}	dto.Response
//	@Failure		500	{object}	dto.Response
//	@Router			/transaction/summary [get]
func (t *TransactionController) GetSummary(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Unauthorized",
			Success: false,
			Error:   "Missing claims",
		})
		return
	}

	email := claims.(pkg.Claims).Email

	summary, err := t.transactionService.GetSummary(ctx.Request.Context(), email)
	if err != nil {
		if err.Error() == "user profile not found" {
			ctx.JSON(http.StatusNotFound, dto.Response{
				Message: "Failed to fetch transaction summary",
				Success: false,
				Error:   "Data tidak ditemukan",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to fetch transaction summary",
			Success: false,
			Error:   "Internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data: dto.SummaryResponse{
			CurrentBalance: summary.CurrentBalance,
			TotalIncome:    summary.TotalIncome,
			TotalExpense:   summary.TotalExpense,
		},
		Message: "Transaction summary successfully retrieved",
		Success: true,
	})
}

// GetTransactionReport godoc
//
//	@Summary		Get transaction report (graph)
//	@Description	Returns income vs expense chart data grouped by day (7days) or by week (30days)
//	@Tags			Transaction
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			range	query		string	false	"Time range: 7days (default) or 30days"	Enums(7days, 30days)
//	@Success		200		{object}	dto.Response{data=dto.TransactionReportResponse}
//	@Failure		401		{object}	dto.Response
//	@Failure		500		{object}	dto.Response
//	@Router			/transaction/report [get]
func (t *TransactionController) GetTransactionReport(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Unauthorized",
			Success: false,
			Error:   "Missing claims",
		})
		return
	}

	email := claims.(pkg.Claims).Email

	rangeParam := ctx.DefaultQuery("range", "7days")
	if rangeParam != "7days" && rangeParam != "30days" {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid range parameter",
			Success: false,
			Error:   "range must be '7days' or '30days'",
		})
		return
	}

	points, err := t.transactionService.GetTransactionReport(ctx.Request.Context(), email, rangeParam)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to fetch transaction report",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	resp := make([]dto.ChartPointResponse, 0, len(points))
	for _, p := range points {
		resp = append(resp, dto.ChartPointResponse{
			Label:   p.Label,
			Income:  p.Income,
			Expense: p.Expense,
		})
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data: dto.TransactionReportResponse{
			Range:  rangeParam,
			Points: resp,
		},
		Message: "Transaction report successfully retrieved",
		Success: true,
	})
}

// CreateTransaction godoc
//
//	@Summary		Create a transaction
//	@Description	Create a new transaction entry for a wallet owned by the authenticated user
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			body	body		dto.CreateTransactionRequest	true	"Transaction payload"
//	@Success		201		{object}	dto.Response{data=dto.TransactionResponse}
//	@Failure		400		{object}	dto.Response
//	@Failure		401		{object}	dto.Response
//	@Failure		500		{object}	dto.Response
//	@Router			/transaction/ [post]
func (t *TransactionController) CreateTransaction(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Unauthorized",
			Success: false,
			Error:   "Missing claims",
		})
		return
	}
	_ = claims.(pkg.Claims).Email // verified user

	var body dto.CreateTransactionRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid request body",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	walletID, err := uuid.Parse(body.WalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid wallet_id",
			Success: false,
			Error:   "wallet_id must be a valid UUID",
		})
		return
	}

	tx := model.Transaction{
		WalletID:  walletID,
		Amount:    body.Amount,
		Direction: model.Direction(body.Direction),
		AdminFee:  body.AdminFee,
		Status:    model.TransactionStatusPending,
		Note:      body.Note,
	}

	created, err := t.transactionService.CreateTransaction(ctx.Request.Context(), tx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to create transaction",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	note := ""
	if created.Note != nil {
		note = *created.Note
	}

	ctx.JSON(http.StatusCreated, dto.Response{
		Data: dto.TransactionResponse{
			ID:        created.ID.String(),
			WalletID:  created.WalletID.String(),
			Amount:    created.Amount,
			Direction: string(created.Direction),
			AdminFee:  created.AdminFee,
			Status:    string(created.Status),
			Note:      note,
			CreatedAt: created.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Message: "Transaction created successfully",
		Success: true,
	})
}

// GetTransactions godoc
//
//	@Summary		List transactions
//	@Description	List all transactions for the authenticated user (optionally filtered by wallet_id)
//	@Tags			Transaction
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			wallet_id	query		string	false	"Filter by wallet UUID"
//	@Param			page		query		int		false	"Page number (default 1)"
//	@Param			limit		query		int		false	"Items per page (default 10)"
//	@Success		200			{object}	dto.Response{data=dto.TransactionListResponse}
//	@Failure		401			{object}	dto.Response
//	@Failure		500			{object}	dto.Response
//	@Router			/transaction/ [get]
func (t *TransactionController) GetTransactions(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Unauthorized",
			Success: false,
			Error:   "Missing claims",
		})
		return
	}
	email := claims.(pkg.Claims).Email

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	walletIDStr := ctx.Query("wallet_id")

	var (
		txs   []model.Transaction
		total int
		err   error
	)

	if walletIDStr != "" {
		walletID, parseErr := uuid.Parse(walletIDStr)
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, dto.Response{
				Message: "Invalid wallet_id",
				Success: false,
				Error:   "wallet_id must be a valid UUID",
			})
			return
		}
		txs, total, err = t.transactionService.GetTransactionsByWallet(ctx.Request.Context(), email, walletID, page, limit)
	} else {
		txs, total, err = t.transactionService.GetAllTransactions(ctx.Request.Context(), email, page, limit)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to fetch transactions",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	responses := make([]dto.TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		note := ""
		if tx.Note != nil {
			note = *tx.Note
		}
		responses = append(responses, dto.TransactionResponse{
			ID:        tx.ID.String(),
			WalletID:  tx.WalletID.String(),
			Amount:    tx.Amount,
			Direction: string(tx.Direction),
			AdminFee:  tx.AdminFee,
			Status:    string(tx.Status),
			Note:      note,
			CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data: dto.TransactionListResponse{
			Data:  responses,
			Total: total,
			Page:  page,
			Limit: limit,
		},
		Message: "Transactions retrieved successfully",
		Success: true,
	})
}

// GetTransactionByID godoc
//
//	@Summary		Get a transaction
//	@Description	Get a single transaction by ID for the authenticated user
//	@Tags			Transaction
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string	true	"Transaction UUID"
//	@Success		200	{object}	dto.Response{data=dto.TransactionResponse}
//	@Failure		400	{object}	dto.Response
//	@Failure		401	{object}	dto.Response
//	@Failure		404	{object}	dto.Response
//	@Router			/transaction/:id [get]
func (t *TransactionController) GetTransactionByID(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Unauthorized",
			Success: false,
			Error:   "Missing claims",
		})
		return
	}
	email := claims.(pkg.Claims).Email

	txID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid transaction id",
			Success: false,
			Error:   "id must be a valid UUID",
		})
		return
	}

	tx, err := t.transactionService.GetTransactionByID(ctx.Request.Context(), email, txID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.Response{
			Message: "Transaction not found",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	note := ""
	if tx.Note != nil {
		note = *tx.Note
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data: dto.TransactionResponse{
			ID:        tx.ID.String(),
			WalletID:  tx.WalletID.String(),
			Amount:    tx.Amount,
			Direction: string(tx.Direction),
			AdminFee:  tx.AdminFee,
			Status:    string(tx.Status),
			Note:      note,
			CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Message: "Transaction retrieved successfully",
		Success: true,
	})
}
