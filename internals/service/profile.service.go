package service

import (
	"context"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type ProfileRepository interface {
	UserProfile(ctx context.Context, email string) (model.Profile, error)
}

type ProfileService struct {
	repo ProfileRepository
}

func NewProfileService(repo ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) GetProfile(ctx context.Context, email string) (model.Profile, error) {
	return s.repo.UserProfile(ctx, email)
}

func (s *ProfileService) EditProfile(ctx context.Context, email string) (model.Profile, error) {
	return s.repo.UserProfile(ctx, email)
}
