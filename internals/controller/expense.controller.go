package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

type ExpenseController struct {
	expenseService *service.ExpenseService
}

func NewExpenseController(expenseService *service.ExpenseService) *ExpenseController {
	return &ExpenseController{expenseService: expenseService}
}

// CreateExpense godoc
// @Summary      Record an expense
// @Description  Deducts wallet balance and stores expense metadata after PIN verification.
// @Tags         Expenses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body           body      dto.ExpenseRequest  true  "Expense payload"
// @Success      201            {object}  dto.Response{data=dto.ExpenseResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      403            {object}  dto.Response{error}
// @Failure      422            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transaction/expense [post]
func (t *ExpenseController) CreateExpense(ctx *gin.Context) {
	email, ok := pkg.CurrentUserEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", errors.New("missing user context")))
		return
	}

	var body dto.ExpenseRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err))
		return
	}

	if !checkPIN(ctx, t.expenseService, email, body.Pin) {
		return
	}

	walletID, err := uuid.Parse(body.WalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid wallet_id", errors.New("wallet_id must be a valid UUID")))
		return
	}

	if !checkWalletOwnership(ctx, t.expenseService, email, walletID) {
		return
	}

	var category, merchantName *string
	if body.Category != "" {
		category = &body.Category
	}
	if body.MerchantName != "" {
		merchantName = &body.MerchantName
	}

	const expenseAdminFee int64 = 0
	tx, err := t.expenseService.CreateExpense(ctx.Request.Context(), walletID, body.Amount, expenseAdminFee, category, merchantName, body.Note)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "insufficient balance" {
			status = http.StatusUnprocessableEntity
		}
		ctx.JSON(status, dto.NewError("Expense recording failed", err))
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
		TransactionID: tx.ID.String(),
		WalletID:      tx.WalletID.String(),
		Amount:        tx.Amount,
		AdminFee:      tx.AdminFee,
		Category:      cat,
		MerchantName:  merch,
		Status:        string(tx.Status),
		CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}))
}
