package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type TransferRepository interface {
	VerifyPIN(ctx context.Context, email, rawPin string) error
	WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error)
	GetWalletOwnerEmail(ctx context.Context, walletID uuid.UUID) (string, error)
	CreateTransfer(ctx context.Context, senderWalletID, recipientWalletID uuid.UUID, amount, adminFee int64, note *string) (model.Transfer, model.Transaction, model.Transaction, error)
}

type TransferService struct {
	repo TransferRepository
	rdb  *redis.Client
}

func NewTransferService(repo TransferRepository, rdb *redis.Client) *TransferService {
	return &TransferService{repo: repo, rdb: rdb}
}

func (s *TransferService) VerifyPIN(ctx context.Context, email, rawPin string) error {
	return s.repo.VerifyPIN(ctx, email, rawPin)
}

func (s *TransferService) WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error) {
	return s.repo.WalletBelongsToUser(ctx, email, walletID)
}

func (s *TransferService) CreateTransfer(ctx context.Context, senderWalletID, recipientWalletID uuid.UUID, amount, adminFee int64, note *string) (model.Transfer, model.Transaction, model.Transaction, error) {
	transfer, senderTx, recipientTx, err := s.repo.CreateTransfer(ctx, senderWalletID, recipientWalletID, amount, adminFee, note)
	if err != nil {
		return model.Transfer{}, model.Transaction{}, model.Transaction{}, err
	}
	bumpHistoryVersionByWalletID(ctx, s.rdb, s.repo, senderWalletID)
	bumpHistoryVersionByWalletID(ctx, s.rdb, s.repo, recipientWalletID)
	return transfer, senderTx, recipientTx, nil
}
