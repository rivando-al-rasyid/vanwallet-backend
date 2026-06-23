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

type TransferController struct {
	transferService *service.TransferService
}

func NewTransferController(transferService *service.TransferService) *TransferController {
	return &TransferController{transferService: transferService}
}

// CreateTransfer godoc
// @Summary      Transfer to another wallet
// @Description  Creates balanced sender and recipient ledger entries after PIN verification.
// @Tags         Transfers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body           body      dto.TransferRequest  true  "Transfer payload"
// @Success      201            {object}  dto.Response{data=dto.TransferResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      403            {object}  dto.Response{error}
// @Failure      422            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transaction/transfer [post]
func (t *TransferController) CreateTransfer(ctx *gin.Context) {
	email, ok := pkg.CurrentUserEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", errors.New("missing user context")))
		return
	}

	var body dto.TransferRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err))
		return
	}

	if !checkPIN(ctx, t.transferService, email, body.Pin) {
		return
	}

	senderWalletID, err := uuid.Parse(body.SenderWalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid sender_wallet_id", errors.New("must be a valid UUID")))
		return
	}

	if !checkWalletOwnership(ctx, t.transferService, email, senderWalletID) {
		return
	}

	recipientWalletID, err := uuid.Parse(body.RecipientWalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid recipient_wallet_id", errors.New("must be a valid UUID")))
		return
	}

	var note *string
	if body.Note != "" {
		note = &body.Note
	}

	const transferAdminFee int64 = 0
	transfer, senderTx, recipientTx, err := t.transferService.CreateTransfer(
		ctx.Request.Context(), senderWalletID, recipientWalletID, body.Amount, transferAdminFee, note,
	)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "insufficient balance":
			status = http.StatusUnprocessableEntity
		case "cannot transfer to the same wallet":
			status = http.StatusBadRequest
		}
		ctx.JSON(status, dto.NewError("Transfer failed", err))
		return
	}

	transferCode := ""
	if transfer.TransferCode != nil {
		transferCode = *transfer.TransferCode
	}

	ctx.JSON(http.StatusCreated, dto.NewSuccess("Transfer completed successfully", dto.TransferResponse{
		TransferCode: transferCode,
		SenderTx:     transactionToResponse(senderTx),
		RecipientInfo: dto.RecipientInfo{
			WalletID: recipientTx.WalletID.String(),
			Amount:   recipientTx.Amount,
		},
	}))
}
