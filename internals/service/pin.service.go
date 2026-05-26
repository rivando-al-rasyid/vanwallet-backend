package service

import (
	"context"
	"fmt"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
)

type PinRepository interface {
	GetPinByEmail(ctx context.Context, email string) (model.UserPin, error)
	SetPin(ctx context.Context, email, hashedPin string) error
}

type PinService struct {
	pinRepo PinRepository
}

func NewPinService(pinRepo PinRepository) *PinService {
	return &PinService{pinRepo: pinRepo}
}

func (s *PinService) HasPin(ctx context.Context, email string) (bool, error) {
	up, err := s.pinRepo.GetPinByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	return up.PinHash != nil && *up.PinHash != "", nil
}

func (s *PinService) SetPin(ctx context.Context, email, rawPin string) error {
	if len(rawPin) != 6 {
		return fmt.Errorf("SetPin: PIN must be exactly 6 digits")
	}
	var hc pkg.HashConfig
	hc.UseRecommended()
	hashed := hc.GenHash(rawPin)
	return s.pinRepo.SetPin(ctx, email, hashed)
}

func (s *PinService) VerifyPin(ctx context.Context, email, rawPin string) error {
	up, err := s.pinRepo.GetPinByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("VerifyPin: %w", err)
	}

	if up.PinHash == nil || *up.PinHash == "" {
		return fmt.Errorf("VerifyPin: no PIN set for this user")
	}

	var hc pkg.HashConfig
	hc.UseRecommended()
	if err := hc.Compare(rawPin, *up.PinHash); err != nil {
		return fmt.Errorf("VerifyPin: %w", err)
	}
	return nil
}
