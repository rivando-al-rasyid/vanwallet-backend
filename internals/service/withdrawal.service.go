package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type WithdrawalRepository interface {
	VerifyPIN(ctx context.Context, email, rawPin string) error
	WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error)
	GetWalletOwnerEmail(ctx context.Context, walletID uuid.UUID) (string, error)
	CreateWithdrawal(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, bank model.Withdrawal) (model.Transaction, error)
}

type WithdrawalService struct {
	repo WithdrawalRepository
	rdb  *redis.Client
}

func NewWithdrawalService(repo WithdrawalRepository, rdb *redis.Client) *WithdrawalService {
	return &WithdrawalService{repo: repo, rdb: rdb}
}

func (s *WithdrawalService) VerifyPIN(ctx context.Context, email, rawPin string) error {
	return s.repo.VerifyPIN(ctx, email, rawPin)
}

func (s *WithdrawalService) WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error) {
	return s.repo.WalletBelongsToUser(ctx, email, walletID)
}

func (s *WithdrawalService) CreateWithdrawal(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, bank model.Withdrawal) (model.Transaction, error) {
	tx, err := s.repo.CreateWithdrawal(ctx, walletID, amount, adminFee, bank)
	if err != nil {
		return model.Transaction{}, err
	}
	bumpHistoryVersionByWalletID(ctx, s.rdb, s.repo, walletID)
	return tx, nil
}
