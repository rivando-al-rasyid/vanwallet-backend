package service

import (
	"context"
	"errors"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/repository"
)

type AuthService struct {
	authRepo *repository.Authrepo
}

func NewAuthService(authRepo *repository.Authrepo) *AuthService {
	return &AuthService{authRepo: authRepo}
}

func (a *AuthService) Register(ctx context.Context, user dto.NewUser) (dto.User, error) {
	var hc pkg.HashConfig
	hc.UseRecommended()
	hashedPwd := hc.GenHash(user.Password)

	newUser, err := a.authRepo.Register(ctx, user.Email, hashedPwd)
	if err != nil {
		return dto.User{}, err
	}

	return dto.User{
		Id:        newUser.Id,
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt,
	}, nil
}

func (a *AuthService) Login(ctx context.Context, user dto.NewUser) (dto.User, error) {
	existingUser, err := a.authRepo.Login(ctx, user.Email)
	if err != nil {
		// Jangan expose apakah email tidak ditemukan atau password salah
		return dto.User{}, errors.New("email atau password salah")
	}

	var hc pkg.HashConfig
	if err := hc.Compare(user.Password, existingUser.Password); err != nil {
		return dto.User{}, errors.New("email atau password salah")
	}

	return dto.User{
		Id:        existingUser.Id,
		Email:     existingUser.Email,
		CreatedAt: existingUser.CreatedAt,
	}, nil
}
