// internals/service/profile_service.go
package service

import (
	"context"
	"mime/multipart"
	"path"
	"strings"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/config"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type ProfileRepository interface {
	UserProfile(ctx context.Context, email string) (model.Profile, error)
	EditProfile(ctx context.Context, email string, updates map[string]any) (model.Profile, error)
	EditPin(ctx context.Context, email string, newPin string) (model.UserPin, error)
	EditPassword(ctx context.Context, email string, newPassword string) (model.User, error)
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

func (s *ProfileService) EditProfile(ctx context.Context, email string, updates map[string]any) (model.Profile, error) {
	return s.repo.EditProfile(ctx, email, updates)
}

func (s *ProfileService) EditPin(ctx context.Context, email string, newPin string) (model.UserPin, error) {
	return s.repo.EditPin(ctx, email, newPin)
}

func (s *ProfileService) EditPassword(ctx context.Context, email string, newPassword string) (model.User, error) {
	return s.repo.EditPassword(ctx, email, newPassword)
}

func (s *ProfileService) ValidateUpload(maxSize int64, photo *multipart.FileHeader) error {
	if photo.Size > maxSize {
		return config.ErrFileTooLarge
	}

	ext := strings.ToLower(path.Ext(photo.Filename))
	if !config.AllowedPhotoExt[ext] {
		return config.ErrExtNotAllowed
	}

	return nil
}
