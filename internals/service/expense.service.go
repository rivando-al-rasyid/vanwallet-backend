package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type ExpenseRepository interface {
	VerifyPIN(ctx context.Context, email, rawPin string) error
	WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error)
	GetWalletOwnerEmail(ctx context.Context, walletID uuid.UUID) (string, error)
	CreateExpense(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, category, merchantName, note *string) (model.Transaction, error)
}

type ExpenseService struct {
	repo ExpenseRepository
	rdb  *redis.Client
}

func NewExpenseService(repo ExpenseRepository, rdb *redis.Client) *ExpenseService {
	return &ExpenseService{repo: repo, rdb: rdb}
}

func (s *ExpenseService) VerifyPIN(ctx context.Context, email, rawPin string) error {
	return s.repo.VerifyPIN(ctx, email, rawPin)
}

func (s *ExpenseService) WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error) {
	return s.repo.WalletBelongsToUser(ctx, email, walletID)
}

func (s *ExpenseService) CreateExpense(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, category, merchantName, note *string) (model.Transaction, error) {
	tx, err := s.repo.CreateExpense(ctx, walletID, amount, adminFee, category, merchantName, note)
	if err != nil {
		return model.Transaction{}, err
	}
	bumpHistoryVersionByWalletID(ctx, s.rdb, s.repo, walletID)
	return tx, nil
}
