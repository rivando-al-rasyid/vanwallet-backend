package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type TransactionRepository interface {
	GetSummary(ctx context.Context, email string) (model.TransactionSummary, error)
	GetTransactionReport(ctx context.Context, email string, rangeParam string) ([]model.ChartPoint, error)
	GetTransactionsByWallet(ctx context.Context, email string, walletID uuid.UUID, page, limit int) ([]model.Transaction, int, error)
	GetAllTransactions(ctx context.Context, email string, page, limit int) ([]model.Transaction, int, error)
	GetTransactionByID(ctx context.Context, email string, transactionID uuid.UUID) (model.Transaction, error)

	CreateTopup(ctx context.Context, req model.Topup) (model.Topup, error)
	ConfirmTopup(ctx context.Context, topupID uuid.UUID) (model.Topup, error)

	CreateWithdrawal(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, bank model.Withdrawal) (model.Transaction, error)

	CreateTransfer(ctx context.Context, senderWalletID, recipientWalletID uuid.UUID, amount, adminFee int64, note *string) (model.Transfer, model.Transaction, model.Transaction, error)

	CreateExpense(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, category, merchantName, note *string) (model.Transaction, error)

	SearchReceivers(ctx context.Context, callerEmail, query string, page, limit int) ([]model.ReceiverResult, int, error)
}

type TransactionService struct {
	repo TransactionRepository
}

func NewTransactionService(repo TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) GetSummary(ctx context.Context, email string) (model.TransactionSummary, error) {
	return s.repo.GetSummary(ctx, email)
}

func (s *TransactionService) GetTransactionReport(ctx context.Context, email string, rangeParam string) ([]model.ChartPoint, error) {
	return s.repo.GetTransactionReport(ctx, email, rangeParam)
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

func (s *TransactionService) CreateTopup(ctx context.Context, req model.Topup) (model.Topup, error) {
	return s.repo.CreateTopup(ctx, req)
}

func (s *TransactionService) ConfirmTopup(ctx context.Context, topupID uuid.UUID) (model.Topup, error) {
	return s.repo.ConfirmTopup(ctx, topupID)
}

func (s *TransactionService) CreateWithdrawal(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, bank model.Withdrawal) (model.Transaction, error) {
	return s.repo.CreateWithdrawal(ctx, walletID, amount, adminFee, bank)
}

func (s *TransactionService) CreateTransfer(ctx context.Context, senderWalletID, recipientWalletID uuid.UUID, amount, adminFee int64, note *string) (model.Transfer, model.Transaction, model.Transaction, error) {
	return s.repo.CreateTransfer(ctx, senderWalletID, recipientWalletID, amount, adminFee, note)
}

func (s *TransactionService) CreateExpense(ctx context.Context, walletID uuid.UUID, amount, adminFee int64, category, merchantName, note *string) (model.Transaction, error) {
	return s.repo.CreateExpense(ctx, walletID, amount, adminFee, category, merchantName, note)
}

func (s *TransactionService) SearchReceivers(ctx context.Context, callerEmail, query string, page, limit int) ([]model.ReceiverResult, int, error) {
	return s.repo.SearchReceivers(ctx, callerEmail, query, page, limit)
}
