package controller

import (
	"errors"
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
// @Description  Retrieves current balance, total income, total expense, and related wallets summary.
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200            {object}  dto.Response{data=dto.SummaryResponse}
// @Failure      401            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transaction/summary [get]
func (t *TransactionController) GetSummary(ctx *gin.Context) {
	email, ok := pkg.CurrentUserEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", errors.New("missing user context")))
		return
	}
	summary, err := t.transactionService.GetSummary(ctx.Request.Context(), email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to fetch summary", err))
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
// @Description  Fetches graphical data for income/expenses across specified temporal intervals.
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        range          query     string  false  "Time span filter" Enums(7days, 30days) default(7days)
// @Param        type           query     string  false  "Financial stream categorization" Enums(income, expense, both) default(both)
// @Success      200            {object}  dto.Response{data=dto.TransactionReportResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transaction/report [get]
func (t *TransactionController) GetTransactionReport(ctx *gin.Context) {
	email, ok := pkg.CurrentUserEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", errors.New("missing user context")))
		return
	}
	rangeParam := ctx.DefaultQuery("range", "7days")
	if rangeParam != "7days" && rangeParam != "30days" {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid range", errors.New("range must be '7days' or '30days'")))
		return
	}
	typeFilter := ctx.DefaultQuery("type", "both")
	if typeFilter != "income" && typeFilter != "expense" && typeFilter != "both" {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid type", errors.New("type must be 'income', 'expense', or 'both'")))
		return
	}
	points, err := t.transactionService.GetTransactionReport(ctx.Request.Context(), email, rangeParam, typeFilter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to fetch report", err))
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

// GetHistory godoc
// @Summary      Get paginated transaction history
// @Description  Returns the unified user transaction history feed. Supports pagination, filters, and simple search.
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page           query     int     false  "Page number" default(1)
// @Param        limit          query     int     false  "Items per page" default(10)
// @Param        wallet_id      query     string  false  "Filter by wallet UUID"
// @Param        source         query     string  false  "Filter source" Enums(transaction, topup)
// @Param        type           query     string  false  "Filter type" Enums(TOPUP, EXPENSE, WITHDRAWAL, TRANSFER_IN, TRANSFER_OUT)
// @Param        status         query     string  false  "Filter status" Enums(PENDING, SUCCESS, FAILED, CANCELLED)
// @Param        direction      query     string  false  "Filter direction" Enums(income, expense)
// @Param        start_date     query     string  false  "Start date YYYY-MM-DD"
// @Param        end_date       query     string  false  "End date YYYY-MM-DD"
// @Param        q              query     string  false  "Search text"
// @Success      200            {object}  dto.Response{data=dto.HistoryListResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transaction/history [get]
func (t *TransactionController) GetHistory(ctx *gin.Context) {
	email, ok := pkg.CurrentUserEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", errors.New("missing user context")))
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

	filter := model.HistoryFilter{
		Page:      page,
		Limit:     limit,
		WalletID:  ctx.Query("wallet_id"),
		Source:    ctx.Query("source"),
		Type:      ctx.Query("type"),
		Status:    ctx.Query("status"),
		Direction: ctx.Query("direction"),
		StartDate: ctx.Query("start_date"),
		EndDate:   ctx.Query("end_date"),
		Query:     ctx.Query("q"),
	}
	if filter.Query == "" {
		filter.Query = ctx.Query("query")
	}

	if filter.WalletID != "" {
		if _, err := uuid.Parse(filter.WalletID); err != nil {
			ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid wallet_id", errors.New("wallet_id must be a valid UUID")))
			return
		}
	}

	items, total, err := t.transactionService.GetAllHistory(ctx.Request.Context(), email, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to fetch history", err))
		return
	}

	resp := make([]dto.HistoryItem, 0, len(items))
	for _, h := range items {
		resp = append(resp, dto.HistoryItem{
			ID:            h.ID,
			Source:        h.Source,
			Type:          h.Type,
			Direction:     h.Direction,
			Title:         historyTitle(h),
			Amount:        h.Amount,
			AdminFee:      h.AdminFee,
			Status:        h.Status,
			PaymentMethod: h.PaymentMethod,
			Note:          h.Note,
			WalletID:      h.WalletID,
			WalletLabel:   h.WalletLabel,
			CreatedAt:     h.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	ctx.JSON(http.StatusOK, dto.NewSuccess("History retrieved successfully", dto.HistoryListResponse{
		Data:       resp,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}))
}
