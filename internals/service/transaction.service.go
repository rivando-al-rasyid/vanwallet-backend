package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, tx model.Transaction) (model.Transaction, error)
	GetTransactionsByWallet(ctx context.Context, email string, walletID uuid.UUID, page, limit int) ([]model.Transaction, int, error)
	GetAllTransactions(ctx context.Context, email string, page, limit int) ([]model.Transaction, int, error)
	GetTransactionByID(ctx context.Context, email string, transactionID uuid.UUID) (model.Transaction, error)
}

type TransactionService struct {
	repo TransactionRepository
}

func NewTransactionService(repo TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, tx model.Transaction) (model.Transaction, error) {
	return s.repo.CreateTransaction(ctx, tx)
}

func (s *TransactionService) GetTransactionsByWallet(ctx context.Context, email string, walletID uuid.UUID, page, limit int) ([]model.Transaction, int, error) {
	return s.repo.GetTransactionsByWallet(ctx, email, walletID, page, limit)
}

func (s *TransactionService) GetAllTransactions(ctx context.Context, email string, page, limit int) ([]model.Transaction, int, error) {
	return s.repo.GetAllTransactions(ctx, email, page, limit)
}

func (s *TransactionService) GetTransactionByID(ctx context.Context, email string, transactionID uuid.UUID) (model.Transaction, error) {
	return s.repo.GetTransactionByID(ctx, email, transactionID)
}
