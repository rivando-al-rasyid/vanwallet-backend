package service

import (
	"context"

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

func (a *AuthService) Register(ctx context.Context, user dto.RegisterRequest) (dto.UserResponse, error) {
	var hc pkg.HashConfig
	hc.UseRecommended()
	hashedPwd := hc.GenHash(user.Password)

	registerResult, err := a.authRepo.Register(ctx, user.Email, hashedPwd)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:    registerResult.ID,
		Email: registerResult.Email,
	}, nil
}

func (a *AuthService) Login(ctx context.Context, user dto.LoginRequest) (string, error) {
	login, err := a.authRepo.Login(ctx, user.Email)
	if err != nil {
		return "", err
	}

	var hc pkg.HashConfig
	if err := hc.Compare(user.Password, login.Password); err != nil {
		return "", err
	}

	claims := pkg.NewClaims(login.ID, user.Email)
	token, err := claims.GenJWT()
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *AuthService) Logout(ctx context.Context, userID string) error {
	return a.authRepo.ClearToken(ctx, userID)
}
