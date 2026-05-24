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
	authService        *service.AuthService
}

func NewTransactionController(transactionService *service.TransactionService, authService *service.AuthService) *TransactionController {
	return &TransactionController{
		transactionService: transactionService,
		authService:        authService,
	}
}

func claimsEmail(ctx *gin.Context) (string, bool) {
	c, exists := ctx.Get("claims")
	if !exists {
		return "", false
	}
	return c.(pkg.Claims).Email, true
}

func (t *TransactionController) verifyPin(ctx *gin.Context, email, pin string) bool {
	userPin, err := t.authService.GetUserPin(ctx.Request.Context(), email)
	if err != nil || userPin.PinHash == nil {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "PIN verification failed",
			Success: false,
			Error:   "PIN not set",
		})
		return false
	}
	var hc pkg.HashConfig
	if err := hc.Compare(pin, *userPin.PinHash); err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "PIN verification failed",
			Success: false,
			Error:   "Invalid PIN",
		})
		return false
	}
	return true
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
//	@Summary		Get financial summary
//	@Description	Returns the authenticated user's total wallet balance, total income (TRANSFER_IN), and total expense (EXPENSE + WITHDRAWAL + TRANSFER_OUT)
//	@Tags			Transaction
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	dto.Response{data=dto.SummaryResponse}
//	@Failure		401	{object}	dto.Response	"Unauthorized"
//	@Failure		500	{object}	dto.Response
//	@Router			/transaction/summary [get]
func (t *TransactionController) GetSummary(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	summary, err := t.transactionService.GetSummary(ctx.Request.Context(), email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Failed to fetch transaction summary", Success: false, Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data:    dto.SummaryResponse{CurrentBalance: summary.CurrentBalance, TotalIncome: summary.TotalIncome, TotalExpense: summary.TotalExpense},
		Message: "Transaction summary successfully retrieved",
		Success: true,
	})
}

// GetTransactionReport godoc
//
//	@Summary		Get chart report
//	@Description	Returns income vs expense chart data. Use range=7days for daily buckets (last 7 days) or range=30days for weekly buckets (last 30 days).
//	@Tags			Transaction
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			range	query		string	false	"Time range"	Enums(7days, 30days)	default(7days)
//	@Success		200		{object}	dto.Response{data=dto.TransactionReportResponse}
//	@Failure		400		{object}	dto.Response	"Invalid range parameter"
//	@Failure		401		{object}	dto.Response	"Unauthorized"
//	@Failure		500		{object}	dto.Response
//	@Router			/transaction/report [get]
func (t *TransactionController) GetTransactionReport(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	rangeParam := ctx.DefaultQuery("range", "7days")
	if rangeParam != "7days" && rangeParam != "30days" {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid range parameter", Success: false, Error: "range must be '7days' or '30days'"})
		return
	}

	points, err := t.transactionService.GetTransactionReport(ctx.Request.Context(), email, rangeParam)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Failed to fetch transaction report", Success: false, Error: err.Error()})
		return
	}

	resp := make([]dto.ChartPointResponse, 0, len(points))
	for _, p := range points {
		resp = append(resp, dto.ChartPointResponse{Label: p.Label, Income: p.Income, Expense: p.Expense})
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data:    dto.TransactionReportResponse{Range: rangeParam, Points: resp},
		Message: "Transaction report successfully retrieved",
		Success: true,
	})
}

// GetTransactions godoc
//
//	@Summary		List transactions
//	@Description	Paginated list of transactions across all wallets. Filter by wallet_id to scope to one wallet.
//	@Tags			Transaction
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			wallet_id	query		string	false	"Filter by wallet UUID"
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			limit		query		int		false	"Items per page (max 100)"	default(10)
//	@Success		200			{object}	dto.Response{data=dto.TransactionListResponse}
//	@Failure		400			{object}	dto.Response	"Invalid wallet_id"
//	@Failure		401			{object}	dto.Response	"Unauthorized"
//	@Failure		500			{object}	dto.Response
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
//	@Description	Fetch a single transaction by UUID. Only accessible if the transaction belongs to the authenticated user.
//	@Tags			Transaction
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string	true	"Transaction UUID"
//	@Success		200	{object}	dto.Response{data=dto.TransactionResponse}
//	@Failure		400	{object}	dto.Response	"Invalid UUID"
//	@Failure		401	{object}	dto.Response	"Unauthorized"
//	@Failure		404	{object}	dto.Response	"Transaction not found"
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
//	@Description	Creates a PENDING top-up record. The wallet balance is NOT credited yet. Call PATCH /transaction/topup/{id}/confirm to complete it.
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			body	body		dto.TopupRequest				true	"Top-up payload"
//	@Success		201		{object}	dto.Response{data=dto.TopupResponse}
//	@Failure		400		{object}	dto.Response	"Invalid payload or wallet_id"
//	@Failure		401		{object}	dto.Response	"Unauthorized"
//	@Failure		500		{object}	dto.Response
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
//	@Description	Sets the top-up status to SUCCESS and credits the wallet balance atomically. Only works on PENDING top-ups.
//	@Tags			Transaction
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string	true	"Topup UUID"
//	@Success		200	{object}	dto.Response{data=dto.TopupResponse}
//	@Failure		400	{object}	dto.Response	"Invalid UUID"
//	@Failure		401	{object}	dto.Response	"Unauthorized"
//	@Failure		404	{object}	dto.Response	"Topup not found or already processed"
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
//	@Description	Debits the wallet and records a WITHDRAWAL transaction atomically. Admin fee of Rp 6.500 is applied. PIN required.
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			body	body		dto.WithdrawalRequest				true	"Withdrawal payload"
//	@Success		201		{object}	dto.Response{data=dto.WithdrawalResponse}
//	@Failure		400		{object}	dto.Response	"Invalid payload"
//	@Failure		401		{object}	dto.Response	"Unauthorized or wrong PIN"
//	@Failure		422		{object}	dto.Response	"Insufficient balance"
//	@Failure		500		{object}	dto.Response
//	@Router			/transaction/withdrawal [post]
func (t *TransactionController) CreateWithdrawal(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	var body dto.WithdrawalRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid request body", Success: false, Error: err.Error()})
		return
	}

	if !t.verifyPin(ctx, email, body.Pin) {
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
//	@Description	Atomically debits the sender wallet and credits the recipient wallet. Creates TRANSFER_OUT and TRANSFER_IN ledger entries linked by a transfer_code. PIN required.
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			body	body		dto.TransferRequest					true	"Transfer payload"
//	@Success		201		{object}	dto.Response{data=dto.TransferResponse}
//	@Failure		400		{object}	dto.Response	"Invalid payload"
//	@Failure		401		{object}	dto.Response	"Unauthorized or wrong PIN"
//	@Failure		422		{object}	dto.Response	"Insufficient balance"
//	@Failure		500		{object}	dto.Response
//	@Router			/transaction/transfer [post]
func (t *TransactionController) CreateTransfer(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	var body dto.TransferRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid request body", Success: false, Error: err.Error()})
		return
	}

	if !t.verifyPin(ctx, email, body.Pin) {
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

	transfer, senderTx, recipientTx, err := t.transactionService.CreateTransfer(
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

	transferCode := ""
	if transfer.TransferCode != nil {
		transferCode = *transfer.TransferCode
	}

	ctx.JSON(http.StatusCreated, dto.Response{
		Data: dto.TransferResponse{
			TransferCode:         transferCode,
			SenderTransaction:    txToResponse(senderTx),
			RecipientTransaction: txToResponse(recipientTx),
		},
		Message: "Transfer completed successfully",
		Success: true,
	})
}

// CreateExpense godoc
//
//	@Summary		Record an expense
//	@Description	Debits the wallet and records an EXPENSE transaction with optional category and merchant metadata. PIN required.
//	@Tags			Transaction
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			body	body		dto.ExpenseRequest				true	"Expense payload"
//	@Success		201		{object}	dto.Response{data=dto.ExpenseResponse}
//	@Failure		400		{object}	dto.Response	"Invalid payload"
//	@Failure		401		{object}	dto.Response	"Unauthorized or wrong PIN"
//	@Failure		422		{object}	dto.Response	"Insufficient balance"
//	@Failure		500		{object}	dto.Response
//	@Router			/transaction/expense [post]
func (t *TransactionController) CreateExpense(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	var body dto.ExpenseRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid request body", Success: false, Error: err.Error()})
		return
	}

	if !t.verifyPin(ctx, email, body.Pin) {
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

	cat := ""
	if category != nil {
		cat = *category
	}
	merch := ""
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
//	@Description	Search other users by full name or phone number (case-insensitive, partial match). The calling user is excluded. Only users who have set their full_name or phone are searchable.
//	@Tags			Transaction
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			q		query		string	true	"Search query (full name or phone)"
//	@Param			page	query		int		false	"Page number"				default(1)
//	@Param			limit	query		int		false	"Items per page (max 50)"	default(10)
//	@Success		200		{object}	dto.Response{data=dto.ReceiverListResponse}
//	@Failure		400		{object}	dto.Response	"Missing or too short query"
//	@Failure		401		{object}	dto.Response
//	@Failure		500		{object}	dto.Response
//	@Router			/transaction/receiver [get]
func (t *TransactionController) FindReceivers(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	query := ctx.Query("q")
	if len(query) < 2 {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid search query",
			Success: false,
			Error:   "q must be at least 2 characters",
		})
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
		Data: dto.ReceiverListResponse{
			Data:  resp,
			Total: total,
			Page:  page,
			Limit: limit,
		},
		Message: "Receivers found",
		Success: true,
	})
}
