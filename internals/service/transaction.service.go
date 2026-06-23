package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/cache"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type TransactionRepository interface {
	GetSummary(ctx context.Context, email string) (model.TransactionSummary, error)
	GetTransactionReport(ctx context.Context, email, rangeParam, typeFilter string) ([]model.ChartPoint, error)
	GetAllHistory(ctx context.Context, email string, filter model.HistoryFilter) ([]model.HistoryItem, int, error)
}

type TransactionService struct {
	repo TransactionRepository
	rdb  *redis.Client
}

func NewTransactionService(repo TransactionRepository, rdb *redis.Client) *TransactionService {
	return &TransactionService{repo: repo, rdb: rdb}
}

func (s *TransactionService) GetSummary(ctx context.Context, email string) (model.TransactionSummary, error) {
	return s.repo.GetSummary(ctx, email)
}

func (s *TransactionService) GetTransactionReport(ctx context.Context, email, rangeParam, typeFilter string) ([]model.ChartPoint, error) {
	return s.repo.GetTransactionReport(ctx, email, rangeParam, typeFilter)
}

func (s *TransactionService) historyCacheVersion(ctx context.Context, email string) string {
	if s.rdb == nil {
		return "0"
	}

	version, err := s.rdb.Get(ctx, historyVersionKey(email)).Result()
	if err == nil {
		return version
	}
	if !errors.Is(err, redis.Nil) {
		log.Println("history version redis error:", err)
	}
	return "0"
}

func (s *TransactionService) GetAllHistory(ctx context.Context, email string, filter model.HistoryFilter) ([]model.HistoryItem, int, error) {
	version := s.historyCacheVersion(ctx, email)
	rkey := fmt.Sprintf(
		"vando:history:%s:v%s:p%d:l%d:w%s:s%s:t%s:st%s:d%s:from%s:to%s:q%s",
		email,
		version,
		filter.Page,
		filter.Limit,
		filter.WalletID,
		filter.Source,
		filter.Type,
		filter.Status,
		filter.Direction,
		filter.StartDate,
		filter.EndDate,
		filter.Query,
	)

	type cacheEntry struct {
		Items []model.HistoryItem `json:"items"`
		Total int                 `json:"total"`
	}
	var entry cacheEntry
	if err := cache.GetFromCache(ctx, s.rdb, rkey, &entry); err == nil {
		log.Println("cache hit:", email)
		return entry.Items, entry.Total, nil
	} else if !errors.Is(err, redis.Nil) {
		log.Println("redis error:", err)
	}

	log.Println("cache miss:", email)

	fetched, total, err := s.repo.GetAllHistory(ctx, email, filter)
	if err != nil {
		return nil, 0, err
	}

	if err := cache.SaveToCache(ctx, s.rdb, rkey, cacheEntry{Items: fetched, Total: total}); err != nil {
		log.Println("cache save error:", err)
	}

	return fetched, total, nil
}
