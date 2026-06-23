package controller

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type pinVerifier interface {
	VerifyPIN(ctx context.Context, email, rawPin string) error
}

type walletOwnerChecker interface {
	WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error)
}

func transactionToResponse(tx model.Transaction) dto.TransactionResponse {
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

func checkPIN(ctx *gin.Context, verifier pinVerifier, email, pin string) bool {
	if pin == "" {
		ctx.JSON(http.StatusBadRequest, dto.NewError("PIN required", errors.New("pin field is required")))
		return false
	}
	if err := verifier.VerifyPIN(ctx.Request.Context(), email, pin); err != nil {
		status := http.StatusUnauthorized
		if strings.Contains(err.Error(), "temporarily locked") {
			status = http.StatusLocked
		}
		ctx.JSON(status, dto.NewError("Invalid PIN", err))
		return false
	}
	return true
}

func checkWalletOwnership(ctx *gin.Context, checker walletOwnerChecker, email string, walletID uuid.UUID) bool {
	owned, err := checker.WalletBelongsToUser(ctx.Request.Context(), email, walletID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Wallet validation failed", errors.New("internal server error")))
		return false
	}
	if !owned {
		ctx.JSON(http.StatusForbidden, dto.NewError("Wallet access denied", errors.New("wallet_id does not belong to the authenticated user")))
		return false
	}
	return true
}
