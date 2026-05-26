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
	return &TransactionController{
		transactionService: transactionService,
	}
}

func claimsEmail(ctx *gin.Context) (string, bool) {
	c, exists := ctx.Get("claims")
	if !exists {
		return "", false
	}
	return c.(pkg.Claims).Email, true
}

func txToResponse(tx model.Transaction) dto.TransactionResponse {
	note := ""
	if tx.Note != nil {
		note = *tx.Note
	}
	ikey := ""
	if tx.IdempotencyKey != nil {
		ikey = *tx.IdempotencyKey
	}
	return dto.TransactionResponse{
		ID:             tx.ID.String(),
		WalletID:       tx.WalletID.String(),
		Type:           string(tx.Type),
		Amount:         tx.Amount,
		AdminFee:       tx.AdminFee,
		Status:         string(tx.Status),
		IdempotencyKey: ikey,
		Note:           note,
		CreatedAt:      tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// GetSummary godoc
//
//	@Summary		Dashboard — financial summary
//	@Description	Returns the authenticated user's total wallet balance, total income (TRANSFER_IN), total expense (EXPENSE + WITHDRAWAL + TRANSFER_OUT), and a list of all wallets with their individual balances.
//	@Tags			Transaction
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.Response{data=dto.SummaryResponse}	"Summary data"
//	@Failure		401	{object}	dto.Response							"Unauthorized or missing token"
//	@Failure		500	{object}	dto.Response							"Internal server error"
//	@Router			/transaction/summary [get]
func (t *TransactionController) GetSummary(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	summary, err := t.transactionService.GetSummary(ctx.Request.Context(), email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Failed to fetch summary", Success: false, Error: err.Error()})
		return
	}

	wallets := make([]dto.WalletItem, 0, len(summary.Wallets))
	for _, w := range summary.Wallets {
		wallets = append(wallets, dto.WalletItem{
			ID:      w.ID.String(),
			Label:   w.Label,
			Balance: w.Balance,
		})
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data: dto.SummaryResponse{
			CurrentBalance: summary.CurrentBalance,
			TotalIncome:    summary.TotalIncome,
			TotalExpense:   summary.TotalExpense,
			Wallets:        wallets,
		},
		Message: "Summary successfully retrieved",
		Success: true,
	})
}

// GetTransactionReport godoc
//
//	@Summary		Dashboard — graph data
//	@Description	Returns chart data for the dashboard graph. Filter by type (income/expense/both) and date range (7days/30days). `7days` returns daily buckets labelled Mon–Sun; `30days` returns weekly buckets labelled W01, W02, etc.
//	@Tags			Transaction
//	@Produce		json
//	@Security		BearerAuth
//	@Param			range	query		string	false	"Date range"	Enums(7days, 30days)		default(7days)
//	@Param			type	query		string	false	"Data type"		Enums(income, expense, both)	default(both)
//	@Success		200		{object}	dto.Response{data=dto.TransactionReportResponse}	"Chart data"
//	@Failure		400		{object}	dto.Response										"Invalid query parameter"
//	@Failure		401		{object}	dto.Response										"Unauthorized or missing token"
//	@Failure		500		{object}	dto.Response										"Internal server error"
//	@Router			/transaction/report [get]
func (t *TransactionController) GetTransactionReport(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	rangeParam := ctx.DefaultQuery("range", "7days")
	if rangeParam != "7days" && rangeParam != "30days" {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid range", Success: false, Error: "range must be '7days' or '30days'"})
		return
	}

	typeFilter := ctx.DefaultQuery("type", "both")
	if typeFilter != "income" && typeFilter != "expense" && typeFilter != "both" {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid type", Success: false, Error: "type must be 'income', 'expense', or 'both'"})
		return
	}

	points, err := t.transactionService.GetTransactionReport(ctx.Request.Context(), email, rangeParam, typeFilter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Failed to fetch report", Success: false, Error: err.Error()})
		return
	}

	resp := make([]dto.ChartPointResponse, 0, len(points))
	for _, p := range points {
		item := dto.ChartPointResponse{Label: p.Label}
		switch typeFilter {
		case "income":
			item.Income = p.Income
		case "expense":
			item.Expense = p.Expense
		default:
			item.Income = p.Income
			item.Expense = p.Expense
		}
		resp = append(resp, item)
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data: dto.TransactionReportResponse{
			Range:  rangeParam,
			Type:   typeFilter,
			Points: resp,
		},
		Message: "Report successfully retrieved",
		Success: true,
	})
}

// GetTransactions godoc
//
//	@Summary		List transactions
//	@Description	Returns a paginated list of transactions across all wallets. Optionally filter by `wallet_id` to scope results to a single wallet. `page` and `limit` default to 1 and 10 respectively; `limit` is capped at 100.
//	@Tags			Transaction
//	@Produce		json
//	@Security		BearerAuth
//	@Param			wallet_id	query		string	false	"Filter by wallet UUID"
//	@Param			page		query		int		false	"Page number (min 1)"			default(1)
//	@Param			limit		query		int		false	"Items per page (1–100)"		default(10)
//	@Success		200			{object}	dto.Response{data=dto.TransactionListResponse}	"Transaction list"
//	@Failure		400			{object}	dto.Response									"Invalid wallet_id UUID"
//	@Failure		401			{object}	dto.Response									"Unauthorized or missing token"
//	@Failure		500			{object}	dto.Response									"Internal server error"
//	@Router			/transaction/ [get]
func (t *TransactionController) GetTransactions(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var (
		txs   []model.Transaction
		total int
		err   error
	)

	if walletIDStr := ctx.Query("wallet_id"); walletIDStr != "" {
		walletID, parseErr := uuid.Parse(walletIDStr)
		if parseErr != nil {
			ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid wallet_id", Success: false, Error: "wallet_id must be a valid UUID"})
			return
		}
		txs, total, err = t.transactionService.GetTransactionsByWallet(ctx.Request.Context(), email, walletID, page, limit)
	} else {
		txs, total, err = t.transactionService.GetAllTransactions(ctx.Request.Context(), email, page, limit)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Failed to fetch transactions", Success: false, Error: err.Error()})
		return
	}

	responses := make([]dto.TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		responses = append(responses, txToResponse(tx))
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data:    dto.TransactionListResponse{Data: responses, Total: total, Page: page, Limit: limit},
		Message: "Transactions retrieved successfully",
		Success: true,
	})
}

// GetTransactionByID godoc
//
//	@Summary		Get transaction detail
//	@Description	Fetch a single transaction by UUID. Returns 404 if the transaction does not exist or does not belong to the authenticated user.
//	@Tags			Transaction
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Transaction UUID"
//	@Success		200	{object}	dto.Response{data=dto.TransactionResponse}	"Transaction detail"
//	@Failure		400	{object}	dto.Response								"Invalid UUID format"
//	@Failure		401	{object}	dto.Response								"Unauthorized or missing token"
//	@Failure		404	{object}	dto.Response								"Transaction not found"
//	@Failure		500	{object}	dto.Response								"Internal server error"
//	@Router			/transaction/{id} [get]
func (t *TransactionController) GetTransactionByID(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	txID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid transaction id", Success: false, Error: "id must be a valid UUID"})
		return
	}

	tx, err := t.transactionService.GetTransactionByID(ctx.Request.Context(), email, txID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.Response{Message: "Transaction not found", Success: false, Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Data: txToResponse(tx), Message: "Transaction retrieved successfully", Success: true})
}

// CreateTopup godoc
//
//	@Summary		Initiate a top-up
//	@Description	Creates a PENDING top-up record. The wallet balance is NOT credited yet. Call PATCH /transaction/topup/{id}/confirm to complete the top-up and credit the balance. No PIN required.
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.TopupRequest					true	"Top-up payload"
//	@Success		201		{object}	dto.Response{data=dto.TopupResponse}	"Top-up initiated; status = PENDING"
//	@Failure		400		{object}	dto.Response						"Invalid payload or wallet_id UUID"
//	@Failure		401		{object}	dto.Response						"Unauthorized or missing token"
//	@Failure		500		{object}	dto.Response						"Internal server error"
//	@Router			/transaction/topup [post]
func (t *TransactionController) CreateTopup(ctx *gin.Context) {
	_, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	var body dto.TopupRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid request body", Success: false, Error: err.Error()})
		return
	}

	walletID, err := uuid.Parse(body.WalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid wallet_id", Success: false, Error: "wallet_id must be a valid UUID"})
		return
	}

	pm := model.PaymentMethod(body.PaymentMethod)
	topup, err := t.transactionService.CreateTopup(ctx.Request.Context(), model.Topup{
		WalletID:      walletID,
		Amount:        body.Amount,
		PaymentMethod: &pm,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Failed to create top-up", Success: false, Error: err.Error()})
		return
	}

	extRef := ""
	if topup.ExternalReference != nil {
		extRef = *topup.ExternalReference
	}
	pmStr := ""
	if topup.PaymentMethod != nil {
		pmStr = string(*topup.PaymentMethod)
	}

	ctx.JSON(http.StatusCreated, dto.Response{
		Data: dto.TopupResponse{
			ID:                topup.ID.String(),
			WalletID:          topup.WalletID.String(),
			Amount:            topup.Amount,
			PaymentMethod:     pmStr,
			ExternalReference: extRef,
			Status:            string(topup.Status),
			CreatedAt:         topup.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Message: "Top-up initiated successfully",
		Success: true,
	})
}

// ConfirmTopup godoc
//
//	@Summary		Confirm a top-up
//	@Description	Sets the top-up status to SUCCESS and atomically credits the wallet balance. Only works on PENDING top-ups; returns 404 if the top-up does not exist or has already been processed.
//	@Tags			Transaction
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Top-up UUID"
//	@Success		200	{object}	dto.Response{data=dto.TopupResponse}	"Top-up confirmed; status = SUCCESS"
//	@Failure		400	{object}	dto.Response						"Invalid UUID format"
//	@Failure		401	{object}	dto.Response						"Unauthorized or missing token"
//	@Failure		404	{object}	dto.Response						"Top-up not found or already processed"
//	@Router			/transaction/topup/{id}/confirm [patch]
func (t *TransactionController) ConfirmTopup(ctx *gin.Context) {
	_, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	topupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid topup id", Success: false, Error: "id must be a valid UUID"})
		return
	}

	topup, err := t.transactionService.ConfirmTopup(ctx.Request.Context(), topupID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.Response{Message: "Failed to confirm top-up", Success: false, Error: err.Error()})
		return
	}

	pmStr := ""
	if topup.PaymentMethod != nil {
		pmStr = string(*topup.PaymentMethod)
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data: dto.TopupResponse{
			ID:            topup.ID.String(),
			WalletID:      topup.WalletID.String(),
			Amount:        topup.Amount,
			PaymentMethod: pmStr,
			Status:        string(topup.Status),
			CreatedAt:     topup.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Message: "Top-up confirmed successfully",
		Success: true,
	})
}

// CreateWithdrawal godoc
//
//	@Summary		Withdraw to bank account
//	@Description	Debits the wallet and records a WITHDRAWAL transaction atomically. A fixed admin fee of Rp 6,500 is deducted in addition to the requested amount. PIN required.
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.WithdrawalRequest					true	"Withdrawal payload"
//	@Success		201		{object}	dto.Response{data=dto.WithdrawalResponse}	"Withdrawal submitted; status = PENDING"
//	@Failure		400		{object}	dto.Response							"Invalid payload or wallet_id UUID"
//	@Failure		401		{object}	dto.Response							"Unauthorized or incorrect PIN"
//	@Failure		422		{object}	dto.Response							"Insufficient balance"
//	@Failure		500		{object}	dto.Response							"Internal server error"
//	@Router			/transaction/withdrawal [post]
func (t *TransactionController) CreateWithdrawal(ctx *gin.Context) {
	_, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	var body dto.WithdrawalRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid request body", Success: false, Error: err.Error()})
		return
	}

	walletID, err := uuid.Parse(body.WalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid wallet_id", Success: false, Error: "wallet_id must be a valid UUID"})
		return
	}

	const withdrawalAdminFee int64 = 6500

	bank := model.Withdrawal{
		BankName:      body.BankName,
		AccountNumber: body.AccountNumber,
		AccountHolder: body.AccountHolder,
	}

	tx, err := t.transactionService.CreateWithdrawal(ctx.Request.Context(), walletID, body.Amount, withdrawalAdminFee, bank)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "insufficient balance" {
			status = http.StatusUnprocessableEntity
		}
		ctx.JSON(status, dto.Response{Message: "Withdrawal failed", Success: false, Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, dto.Response{
		Data: dto.WithdrawalResponse{
			TransactionID: tx.ID.String(),
			WalletID:      tx.WalletID.String(),
			Amount:        tx.Amount,
			AdminFee:      tx.AdminFee,
			BankName:      body.BankName,
			AccountNumber: body.AccountNumber,
			AccountHolder: body.AccountHolder,
			Status:        string(tx.Status),
			CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Message: "Withdrawal submitted successfully",
		Success: true,
	})
}

// CreateTransfer godoc
//
//	@Summary		Transfer funds between wallets
//	@Description	Atomically debits the sender wallet and credits the recipient wallet. Creates a TRANSFER_OUT entry on the sender and a TRANSFER_IN entry on the recipient, both linked by a shared transfer_code. PIN required.
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.TransferRequest					true	"Transfer payload"
//	@Success		201		{object}	dto.Response						"Transfer completed"
//	@Failure		400		{object}	dto.Response						"Invalid payload or wallet UUID"
//	@Failure		401		{object}	dto.Response						"Unauthorized or incorrect PIN"
//	@Failure		422		{object}	dto.Response						"Insufficient balance"
//	@Failure		500		{object}	dto.Response						"Internal server error"
//	@Router			/transaction/transfer [post]
func (t *TransactionController) CreateTransfer(ctx *gin.Context) {
	_, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	var body dto.TransferRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid request body", Success: false, Error: err.Error()})
		return
	}

	senderWalletID, err := uuid.Parse(body.SenderWalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid sender_wallet_id", Success: false, Error: "must be a valid UUID"})
		return
	}
	recipientWalletID, err := uuid.Parse(body.RecipientWalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid recipient_wallet_id", Success: false, Error: "must be a valid UUID"})
		return
	}

	var note *string
	if body.Note != "" {
		note = &body.Note
	}

	const transferAdminFee int64 = 0
	transfer, _, _, err := t.transactionService.CreateTransfer(
		ctx.Request.Context(), senderWalletID, recipientWalletID, body.Amount, transferAdminFee, note,
	)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "insufficient balance" {
			status = http.StatusUnprocessableEntity
		}
		ctx.JSON(status, dto.Response{Message: "Transfer failed", Success: false, Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, dto.Response{
		Data:    transfer,
		Message: "Transfer completed successfully",
		Success: true,
	})
}

// CreateExpense godoc
//
//	@Summary		Record an expense
//	@Description	Debits the specified wallet and records an EXPENSE transaction. Optional fields: category and merchant_name. PIN required.
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.ExpenseRequest					true	"Expense payload"
//	@Success		201		{object}	dto.Response{data=dto.ExpenseResponse}	"Expense recorded"
//	@Failure		400		{object}	dto.Response						"Invalid payload or wallet_id UUID"
//	@Failure		401		{object}	dto.Response						"Unauthorized or incorrect PIN"
//	@Failure		422		{object}	dto.Response						"Insufficient balance"
//	@Failure		500		{object}	dto.Response						"Internal server error"
//	@Router			/transaction/expense [post]
func (t *TransactionController) CreateExpense(ctx *gin.Context) {
	_, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	var body dto.ExpenseRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid request body", Success: false, Error: err.Error()})
		return
	}

	walletID, err := uuid.Parse(body.WalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid wallet_id", Success: false, Error: "wallet_id must be a valid UUID"})
		return
	}

	var category, merchantName *string
	if body.Category != "" {
		category = &body.Category
	}
	if body.MerchantName != "" {
		merchantName = &body.MerchantName
	}

	tx, err := t.transactionService.CreateExpense(ctx.Request.Context(), walletID, body.Amount, body.AdminFee, category, merchantName, body.Note)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "insufficient balance" {
			status = http.StatusUnprocessableEntity
		}
		ctx.JSON(status, dto.Response{Message: "Expense recording failed", Success: false, Error: err.Error()})
		return
	}

	cat, merch := "", ""
	if category != nil {
		cat = *category
	}
	if merchantName != nil {
		merch = *merchantName
	}

	ctx.JSON(http.StatusCreated, dto.Response{
		Data: dto.ExpenseResponse{
			TransactionID: tx.ID.String(),
			WalletID:      tx.WalletID.String(),
			Amount:        tx.Amount,
			AdminFee:      tx.AdminFee,
			Category:      cat,
			MerchantName:  merch,
			Status:        string(tx.Status),
			CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Message: "Expense recorded successfully",
		Success: true,
	})
}

// FindReceivers godoc
//
//	@Summary		Search transfer receivers
//	@Description	Search users by full name or phone number (case-insensitive partial match). Excludes the calling user. Only users who have set their full_name or phone are searchable. `q` must be at least 2 characters. `limit` is capped at 50.
//	@Tags			Transaction
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q		query		string	true	"Search query — full name or phone number (min 2 chars)"
//	@Param			page	query		int		false	"Page number (min 1)"		default(1)
//	@Param			limit	query		int		false	"Items per page (1–50)"		default(10)
//	@Success		200		{object}	dto.Response{data=dto.ReceiverListResponse}	"Matching receivers"
//	@Failure		400		{object}	dto.Response								"Query too short (< 2 characters)"
//	@Failure		401		{object}	dto.Response								"Unauthorized or missing token"
//	@Failure		500		{object}	dto.Response								"Internal server error"
//	@Router			/transaction/receiver [get]
func (t *TransactionController) FindReceivers(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	query := ctx.Query("q")
	if len(query) < 2 {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid search query", Success: false, Error: "q must be at least 2 characters"})
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	results, total, err := t.transactionService.SearchReceivers(ctx.Request.Context(), email, query, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Search failed", Success: false, Error: "Internal server error"})
		return
	}

	resp := make([]dto.ReceiverResult, 0, len(results))
	for _, r := range results {
		item := dto.ReceiverResult{
			UserID:      r.UserID.String(),
			Email:       r.Email,
			WalletID:    r.WalletID.String(),
			WalletLabel: r.WalletLabel,
		}
		if r.FullName != nil {
			item.FullName = *r.FullName
		}
		if r.Phone != nil {
			item.Phone = *r.Phone
		}
		if r.Photo != nil {
			item.Photo = *r.Photo
		}
		resp = append(resp, item)
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data:    dto.ReceiverListResponse{Data: resp, Total: total, Page: page, Limit: limit},
		Message: "Receivers found",
		Success: true,
	})
}
