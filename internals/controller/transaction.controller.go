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

func (t *TransactionController) checkPIN(ctx *gin.Context, email, pin string) bool {
	if pin == "" {
		ctx.JSON(http.StatusBadRequest, dto.NewError("PIN required", "pin field is required"))
		return false
	}
	if err := t.transactionService.VerifyPIN(ctx.Request.Context(), email, pin); err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Invalid PIN", err.Error()))
		return false
	}
	return true
}

func historyDirection(txType string) string {
	switch txType {
	case "TOPUP", "TRANSFER_IN":
		return "income"
	default:
		return "expense"
	}
}

func historyTitle(h model.HistoryItem) string {
	if h.Note != "" {
		return h.Note
	}
	switch h.Type {
	case "TOPUP":
		return "Top Up"
	case "TRANSFER_IN":
		return "Transfer Received"
	case "TRANSFER_OUT":
		return "Transfer Sent"
	case "WITHDRAWAL":
		return "Withdrawal"
	case "EXPENSE":
		return "Expense"
	default:
		return h.Type
	}
}

// GetSummary godoc
// @Summary      Get user financial summary
// @Description  Retrieves current balance, total income, total expense, and related wallets summary[cite: 1].
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Success      200            {object}  dto.Response{data=dto.SummaryResponse}
// @Failure      401            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transactions/summary [get]
func (t *TransactionController) GetSummary(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	summary, err := t.transactionService.GetSummary(ctx.Request.Context(), email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to fetch summary", err.Error()))
		return
	}
	wallets := make([]dto.WalletItem, 0, len(summary.Wallets))
	for _, w := range summary.Wallets {
		wallets = append(wallets, dto.WalletItem{ID: w.ID.String(), Label: w.Label, Balance: w.Balance})
	}
	ctx.JSON(http.StatusOK, dto.NewSuccess("Summary successfully retrieved", dto.SummaryResponse{CurrentBalance: summary.CurrentBalance, TotalIncome: summary.TotalIncome, TotalExpense: summary.TotalExpense, Wallets: wallets}))
}

// GetTransactionReport godoc
// @Summary      Get chart report analytics
// @Description  Fetches graphical data for income/expenses across specified temporal intervals[cite: 1].
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        range          query     string  false  "Time span filter" Enums(7days, 30days) default(7days)
// @Param        type           query     string  false  "Financial stream categorization" Enums(income, expense, both) default(both)
// @Success      200            {object}  dto.Response{data=dto.TransactionReportResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transactions/report [get]
func (t *TransactionController) GetTransactionReport(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	rangeParam := ctx.DefaultQuery("range", "7days")
	if rangeParam != "7days" && rangeParam != "30days" {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid range", "range must be '7days' or '30days'"))
		return
	}
	typeFilter := ctx.DefaultQuery("type", "both")
	if typeFilter != "income" && typeFilter != "expense" && typeFilter != "both" {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid type", "type must be 'income', 'expense', or 'both'"))
		return
	}
	points, err := t.transactionService.GetTransactionReport(ctx.Request.Context(), email, rangeParam, typeFilter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to fetch report", err.Error()))
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
	ctx.JSON(http.StatusOK, dto.NewSuccess("Report successfully retrieved", dto.TransactionReportResponse{Range: rangeParam, Type: typeFilter, Points: resp}))
}

// GetTransactions godoc
// @Summary      List all technical ledger records
// @Description  Queries low-level ledger items with pagination, option to limit scope by wallet UUID[cite: 1].
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        page           query     int     false  "Page target number" default(1)
// @Param        limit          query     int     false  "Data slice size boundary" default(10)
// @Param        wallet_id      query     string  false  "Target operational entity UUID filter"
// @Success      200            {object}  dto.Response{data=dto.TransactionListResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transactions [get]
func (t *TransactionController) GetTransactions(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
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
			ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid wallet_id", "wallet_id must be a valid UUID"))
			return
		}
		txs, total, err = t.transactionService.GetTransactionsByWallet(ctx.Request.Context(), email, walletID, page, limit)
	} else {
		txs, total, err = t.transactionService.GetAllTransactions(ctx.Request.Context(), email, page, limit)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to fetch transactions", err.Error()))
		return
	}
	responses := make([]dto.TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		responses = append(responses, txToResponse(tx))
	}
	ctx.JSON(http.StatusOK, dto.NewSuccess("Transactions retrieved successfully", dto.TransactionListResponse{Data: responses, Total: total, Page: page, Limit: limit}))
}

// GetHistory godoc
// @Summary      Get user friendly historical logs
// @Description  Queries normalized user operational logs suitable for high-level transactional feeds[cite: 1].
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        page           query     int     false  "Page target number" default(1)
// @Param        limit          query     int     false  "Data slice size boundary" default(10)
// @Success      200            {object}  dto.Response{data=dto.HistoryListResponse}
// @Failure      401            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transactions/history [get]
func (t *TransactionController) GetHistory(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
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
	items, total, err := t.transactionService.GetAllHistory(ctx.Request.Context(), email, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to fetch history", err.Error()))
		return
	}
	resp := make([]dto.HistoryItem, 0, len(items))
	for _, h := range items {
		resp = append(resp, dto.HistoryItem{
			ID:            h.ID,
			Source:        h.Source,
			Type:          h.Type,
			Direction:     historyDirection(h.Type),
			Title:         historyTitle(h),
			Amount:        h.Amount,
			AdminFee:      h.AdminFee,
			Status:        h.Status,
			PaymentMethod: h.PaymentMethod,
			Note:          h.Note,
			WalletID:      h.WalletID,
			CreatedAt:     h.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	ctx.JSON(http.StatusOK, dto.NewSuccess("History retrieved successfully", dto.HistoryListResponse{Data: resp, Total: total, Page: page, Limit: limit}))
}

// GetTransactionByID godoc
// @Summary      Get isolated ledger item details
// @Description  Extracts data attributes of a specific transaction by explicit UUID entry[cite: 1].
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        id             path      string  true  "Transaction UUID identifier"
// @Success      200            {object}  dto.Response{data=dto.TransactionResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      404            {object}  dto.Response{error}
// @Router       /transactions/{id} [get]
func (t *TransactionController) GetTransactionByID(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	txID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid transaction id", "id must be a valid UUID"))
		return
	}
	tx, err := t.transactionService.GetTransactionByID(ctx.Request.Context(), email, txID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.NewError("Transaction not found", err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, dto.NewSuccess("Transaction retrieved successfully", txToResponse(tx)))
}

// CreateTopup godoc
// @Summary      Initiate financial deposit pipeline
// @Description  Spawns an internal pending top-up structure targeting an active wallet instance[cite: 1].
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        body           body      dto.TopupRequest  true  "Deposit payload specs"
// @Success      201            {object}  dto.Response{data=dto.TopupResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transactions/topup [post]
func (t *TransactionController) CreateTopup(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	var body dto.TopupRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err.Error()))
		return
	}
	if !t.checkPIN(ctx, email, body.Pin) {
		return
	}
	walletID, err := uuid.Parse(body.WalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid wallet_id", "wallet_id must be a valid UUID"))
		return
	}
	pm := model.PaymentMethod(body.PaymentMethod)
	topup, err := t.transactionService.CreateTopup(ctx.Request.Context(), model.Topup{
		WalletID:      walletID,
		Amount:        body.Amount,
		PaymentMethod: &pm,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to create top-up", err.Error()))
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
	ctx.JSON(http.StatusCreated, dto.NewSuccess("Top-up initiated successfully", dto.TopupResponse{
		ID: topup.ID.String(), WalletID: topup.WalletID.String(),
		Amount: topup.Amount, PaymentMethod: pmStr,
		ExternalReference: extRef, Status: string(topup.Status),
		CreatedAt: topup.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}))
}

// ConfirmTopup godoc
// @Summary      Finalize operational top-up event
// @Description  Validates internal states to switch a top-up balance allocation from pending to active[cite: 1].
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        id             path      string  true  "Top-up entity UUID pointer"
// @Success      200            {object}  dto.Response{data=dto.TopupResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      404            {object}  dto.Response{error}
// @Router       /transactions/topup/{id}/confirm [post]
func (t *TransactionController) ConfirmTopup(ctx *gin.Context) {
	_, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	topupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid topup id", "id must be a valid UUID"))
		return
	}
	topup, err := t.transactionService.ConfirmTopup(ctx.Request.Context(), topupID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.NewError("Failed to confirm top-up", err.Error()))
		return
	}
	pmStr := ""
	if topup.PaymentMethod != nil {
		pmStr = string(*topup.PaymentMethod)
	}
	ctx.JSON(http.StatusOK, dto.NewSuccess("Top-up confirmed successfully", dto.TopupResponse{
		ID: topup.ID.String(), WalletID: topup.WalletID.String(),
		Amount: topup.Amount, PaymentMethod: pmStr,
		Status: string(topup.Status), CreatedAt: topup.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}))
}

// CreateWithdrawal godoc
// @Summary      Execute capital exit pipeline
// @Description  Deducts assets out of the e-wallet matrix into external physical banking infrastructure[cite: 1].
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        body           body      dto.WithdrawalRequest  true  "Withdrawal parameters details"
// @Success      201            {object}  dto.Response{data=dto.WithdrawalResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      422            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transactions/withdraw [post]
func (t *TransactionController) CreateWithdrawal(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	var body dto.WithdrawalRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err.Error()))
		return
	}
	if !t.checkPIN(ctx, email, body.Pin) {
		return
	}
	walletID, err := uuid.Parse(body.WalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid wallet_id", "wallet_id must be a valid UUID"))
		return
	}
	const withdrawalAdminFee int64 = 6500
	bank := model.Withdrawal{BankName: body.BankName, AccountNumber: body.AccountNumber, AccountHolder: body.AccountHolder}
	tx, err := t.transactionService.CreateWithdrawal(ctx.Request.Context(), walletID, body.Amount, withdrawalAdminFee, bank)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "insufficient balance" {
			status = http.StatusUnprocessableEntity
		}
		ctx.JSON(status, dto.NewError("Withdrawal failed", err.Error()))
		return
	}
	ctx.JSON(http.StatusCreated, dto.NewSuccess("Withdrawal submitted successfully", dto.WithdrawalResponse{
		TransactionID: tx.ID.String(), WalletID: tx.WalletID.String(),
		Amount: tx.Amount, AdminFee: tx.AdminFee,
		BankName: body.BankName, AccountNumber: body.AccountNumber, AccountHolder: body.AccountHolder,
		Status: string(tx.Status), CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}))
}

// CreateTransfer godoc
// @Summary      Inter-wallet peer asset migration
// @Description  Executes nuclear ledger balanced migrations across separate user wallets inside the application[cite: 1].
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        body           body      dto.TransferRequest  true  "Transfer structural definition payload"
// @Success      201            {object}  dto.Response{data=dto.TransferResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      422            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transactions/transfer [post]
func (t *TransactionController) CreateTransfer(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	var body dto.TransferRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err.Error()))
		return
	}
	if !t.checkPIN(ctx, email, body.Pin) {
		return
	}
	senderWalletID, err := uuid.Parse(body.SenderWalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid sender_wallet_id", "must be a valid UUID"))
		return
	}
	recipientWalletID, err := uuid.Parse(body.RecipientWalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid recipient_wallet_id", "must be a valid UUID"))
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
		ctx.JSON(status, dto.NewError("Transfer failed", err.Error()))
		return
	}
	transferCode := ""
	if transfer.TransferCode != nil {
		transferCode = *transfer.TransferCode
	}
	ctx.JSON(http.StatusCreated, dto.NewSuccess("Transfer completed successfully", dto.TransferResponse{
		TransferCode: transferCode,
		SenderTx:     txToResponse(senderTx),
		RecipientInfo: dto.RecipientInfo{
			WalletID: recipientTx.WalletID.String(),
			Amount:   recipientTx.Amount,
		},
	}))
}

// CreateExpense godoc
// @Summary      Log an outward operational purchase
// @Description  Commits real-time commercial transactional operations by updating categorical metadata indexes[cite: 1].
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        body           body      dto.ExpenseRequest  true  "Expense context descriptors body"
// @Success      201            {object}  dto.Response{data=dto.ExpenseResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      422            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transactions/expense [post]
func (t *TransactionController) CreateExpense(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	var body dto.ExpenseRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err.Error()))
		return
	}
	if !t.checkPIN(ctx, email, body.Pin) {
		return
	}
	walletID, err := uuid.Parse(body.WalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid wallet_id", "wallet_id must be a valid UUID"))
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
		ctx.JSON(status, dto.NewError("Expense recording failed", err.Error()))
		return
	}
	cat, merch := "", ""
	if category != nil {
		cat = *category
	}
	if merchantName != nil {
		merch = *merchantName
	}
	ctx.JSON(http.StatusCreated, dto.NewSuccess("Expense recorded successfully", dto.ExpenseResponse{
		TransactionID: tx.ID.String(), WalletID: tx.WalletID.String(),
		Amount: tx.Amount, AdminFee: tx.AdminFee,
		Category: cat, MerchantName: merch,
		Status: string(tx.Status), CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}))
}

// FindReceivers godoc
// @Summary      Search system profiles for transactions
// @Description  Returns filtered structural matches for processing inter-user transactional lookups[cite: 1].
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        q              query     string  true   "Lookup parameter string minimum length 2 characters"
// @Param        page           query     int     false  "Page target number" default(1)
// @Param        limit          query     int     false  "Data slice size boundary" default(10)
// @Success      200            {object}  dto.Response{data=dto.ReceiverListResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transactions/receivers [get]
func (t *TransactionController) FindReceivers(ctx *gin.Context) {
	email, ok := claimsEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	query := ctx.Query("q")
	if len(query) < 2 {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid search query", "q must be at least 2 characters"))
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
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Search failed", "Internal server error"))
		return
	}
	resp := make([]dto.ReceiverResult, 0, len(results))
	for _, r := range results {
		item := dto.ReceiverResult{UserID: r.UserID.String(), Email: r.Email, WalletID: r.WalletID.String(), WalletLabel: r.WalletLabel}
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
	ctx.JSON(http.StatusOK, dto.NewSuccess("Receivers found", dto.ReceiverListResponse{Data: resp, Total: total, Page: page, Limit: limit}))
}
