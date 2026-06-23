package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type TopupRepository interface {
	WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error)
	GetWalletOwnerEmail(ctx context.Context, walletID uuid.UUID) (string, error)
	CreateTopup(ctx context.Context, req model.Topup) (model.Topup, error)
}

type TopupService struct {
	repo TopupRepository
	rdb  *redis.Client
}

func NewTopupService(repo TopupRepository, rdb *redis.Client) *TopupService {
	return &TopupService{repo: repo, rdb: rdb}
}

func (s *TopupService) WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error) {
	return s.repo.WalletBelongsToUser(ctx, email, walletID)
}

func (s *TopupService) CreateTopup(ctx context.Context, req model.Topup) (model.Topup, error) {
	topup, err := s.repo.CreateTopup(ctx, req)
	if err != nil {
		return model.Topup{}, err
	}
	bumpHistoryVersionByWalletID(ctx, s.rdb, s.repo, topup.WalletID)
	return topup, nil
}
