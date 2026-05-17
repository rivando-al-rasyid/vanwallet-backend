package service

import (
	"context"
	"errors"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type ProfileRepository interface {
	UserProfile(ctx context.Context, email string) (model.User, model.Profile, error)
}

type ProfileService struct {
	repo ProfileRepository
}

func NewProfileService(repo ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) GetProfileByEmail(ctx context.Context, email string) (model.User, model.Profile, error) {
	if email == "" {
		return model.User{}, model.Profile{}, errors.New("email cannot be empty")
	}

	return s.repo.UserProfile(ctx, email)
}
