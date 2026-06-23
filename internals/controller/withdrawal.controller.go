package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

type WithdrawalController struct {
	withdrawalService *service.WithdrawalService
}

func NewWithdrawalController(withdrawalService *service.WithdrawalService) *WithdrawalController {
	return &WithdrawalController{withdrawalService: withdrawalService}
}

// CreateWithdrawal godoc
// @Summary      Create withdrawal
// @Description  Deducts wallet balance and creates a withdrawal detail record after PIN verification.
// @Tags         Withdrawals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body           body      dto.WithdrawalRequest  true  "Withdrawal payload"
// @Success      201            {object}  dto.Response{data=dto.WithdrawalResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      403            {object}  dto.Response{error}
// @Failure      422            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transaction/withdrawal [post]
func (t *WithdrawalController) CreateWithdrawal(ctx *gin.Context) {
	email, ok := pkg.CurrentUserEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", errors.New("missing user context")))
		return
	}

	var body dto.WithdrawalRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err))
		return
	}

	if !checkPIN(ctx, t.withdrawalService, email, body.Pin) {
		return
	}

	walletID, err := uuid.Parse(body.WalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid wallet_id", errors.New("wallet_id must be a valid UUID")))
		return
	}

	if !checkWalletOwnership(ctx, t.withdrawalService, email, walletID) {
		return
	}

	const withdrawalAdminFee int64 = 6500
	bank := model.Withdrawal{BankName: body.BankName, AccountNumber: body.AccountNumber, AccountHolder: body.AccountHolder}
	tx, err := t.withdrawalService.CreateWithdrawal(ctx.Request.Context(), walletID, body.Amount, withdrawalAdminFee, bank)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "insufficient balance" {
			status = http.StatusUnprocessableEntity
		}
		ctx.JSON(status, dto.NewError("Withdrawal failed", err))
		return
	}

	ctx.JSON(http.StatusCreated, dto.NewSuccess("Withdrawal submitted successfully", dto.WithdrawalResponse{
		TransactionID: tx.ID.String(),
		WalletID:      tx.WalletID.String(),
		Amount:        tx.Amount,
		AdminFee:      tx.AdminFee,
		BankName:      body.BankName,
		AccountNumber: body.AccountNumber,
		AccountHolder: body.AccountHolder,
		Status:        string(tx.Status),
		CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}))
}
