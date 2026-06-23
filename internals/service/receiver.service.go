package service

import (
	"context"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type ReceiverRepository interface {
	SearchReceivers(ctx context.Context, callerEmail, query string, page, limit int) ([]model.ReceiverResult, int, error)
}

type ReceiverService struct {
	repo ReceiverRepository
}

func NewReceiverService(repo ReceiverRepository) *ReceiverService {
	return &ReceiverService{repo: repo}
}

func (s *ReceiverService) SearchReceivers(ctx context.Context, callerEmail, query string, page, limit int) ([]model.ReceiverResult, int, error) {
	return s.repo.SearchReceivers(ctx, callerEmail, query, page, limit)
}
