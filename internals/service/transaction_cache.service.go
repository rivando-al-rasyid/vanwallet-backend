package service

import (
	"context"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type walletOwnerEmailGetter interface {
	GetWalletOwnerEmail(ctx context.Context, walletID uuid.UUID) (string, error)
}

func historyVersionKey(email string) string {
	return "vando:history-version:" + strings.ToLower(strings.TrimSpace(email))
}

func bumpHistoryVersion(ctx context.Context, rdb *redis.Client, email string) {
	if rdb == nil || strings.TrimSpace(email) == "" {
		return
	}

	if err := rdb.Incr(ctx, historyVersionKey(email)).Err(); err != nil {
		log.Println("history cache invalidation error:", err)
	}
}

func bumpHistoryVersionByWalletID(ctx context.Context, rdb *redis.Client, repo walletOwnerEmailGetter, walletID uuid.UUID) {
	email, err := repo.GetWalletOwnerEmail(ctx, walletID)
	if err != nil {
		log.Println("history owner lookup error:", err)
		return
	}
	bumpHistoryVersion(ctx, rdb, email)
}
