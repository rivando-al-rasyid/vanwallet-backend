package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type TopupRepository interface {
	WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error)
	GetWalletOwnerEmail(ctx context.Context, walletID uuid.UUID) (string, error)
	CreateTopup(ctx context.Context, req model.Topup) (model.Topup, error)
	UpdateTopupPayment(ctx context.Context, topupID uuid.UUID, externalRef string, metadata []byte) error
	SettleTopup(ctx context.Context, topupID uuid.UUID, metadata []byte) (credited bool, walletID uuid.UUID, err error)
	UpdateTopupStatus(ctx context.Context, topupID uuid.UUID, status model.TransactionStatus, metadata []byte) error
	GetTopupWalletID(ctx context.Context, topupID uuid.UUID) (uuid.UUID, error)
}

type TopupService struct {
	repo     TopupRepository
	rdb      *redis.Client
	midtrans *MidtransService
}

func NewTopupService(repo TopupRepository, rdb *redis.Client, midtrans *MidtransService) *TopupService {
	return &TopupService{repo: repo, rdb: rdb, midtrans: midtrans}
}

func (s *TopupService) WalletBelongsToUser(ctx context.Context, email string, walletID uuid.UUID) (bool, error) {
	return s.repo.WalletBelongsToUser(ctx, email, walletID)
}

type TopupInitiation struct {
	Topup       model.Topup
	SnapToken   string
	RedirectURL string
	ClientKey   string
}

func (s *TopupService) CreateTopup(ctx context.Context, email string, req model.Topup) (TopupInitiation, error) {
	if s.midtrans == nil {
		return TopupInitiation{}, errors.New("midtrans is not configured")
	}

	topup, err := s.repo.CreateTopup(ctx, req)
	if err != nil {
		return TopupInitiation{}, err
	}

	orderID := topup.ID.String()
	pm := ""
	if req.PaymentMethod != nil {
		pm = string(*req.PaymentMethod)
	}

	snapResult, err := s.midtrans.CreateSnapTransaction(orderID, req.Amount, *req.PaymentMethod, email)
	if err != nil {
		return TopupInitiation{}, fmt.Errorf("create midtrans snap transaction: %w", err)
	}

	metadata, err := json.Marshal(map[string]string{
		"provider":       "midtrans",
		"order_id":       orderID,
		"snap_token":     snapResult.Token,
		"redirect_url":   snapResult.RedirectURL,
		"payment_method": pm,
	})
	if err != nil {
		return TopupInitiation{}, fmt.Errorf("marshal payment metadata: %w", err)
	}

	if err := s.repo.UpdateTopupPayment(ctx, topup.ID, orderID, metadata); err != nil {
		return TopupInitiation{}, err
	}

	bumpHistoryVersionByWalletID(ctx, s.rdb, s.repo, topup.WalletID)

	topup.ExternalReference = &orderID
	topup.PaymentMetadata = &metadata

	return TopupInitiation{
		Topup:       topup,
		SnapToken:   snapResult.Token,
		RedirectURL: snapResult.RedirectURL,
		ClientKey:   s.midtrans.ClientKey(),
	}, nil
}

func (s *TopupService) HandleMidtransNotification(ctx context.Context, notification dto.MidtransNotification, rawBody []byte) error {
	if s.midtrans == nil {
		return errors.New("midtrans is not configured")
	}

	if !s.midtrans.VerifyNotificationSignature(
		notification.OrderID,
		notification.StatusCode,
		notification.GrossAmount,
		notification.SignatureKey,
	) {
		return errors.New("invalid midtrans signature")
	}

	topupID, err := uuid.Parse(notification.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order_id: %w", err)
	}

	nextStatus, ok := MidtransStatusToTopupStatus(notification.TransactionStatus, notification.FraudStatus)
	if !ok {
		return nil
	}

	switch nextStatus {
	case model.TransactionStatusSuccess:
		credited, walletID, err := s.repo.SettleTopup(ctx, topupID, rawBody)
		if err != nil {
			return err
		}
		if credited {
			bumpHistoryVersionByWalletID(ctx, s.rdb, s.repo, walletID)
		}
		return nil
	case model.TransactionStatusFailed, model.TransactionStatusCancelled:
		if err := s.repo.UpdateTopupStatus(ctx, topupID, nextStatus, rawBody); err != nil {
			return err
		}
		if walletID, err := s.repo.GetTopupWalletID(ctx, topupID); err == nil {
			bumpHistoryVersionByWalletID(ctx, s.rdb, s.repo, walletID)
		}
		return nil
	default:
		return nil
	}
}
