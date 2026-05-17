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

func (a *AuthService) Login(ctx context.Context, user dto.NewUser) (string, error) {

	login, err := a.authRepo.Login(ctx, user.Email)
	if err != nil {
		return "", err
	}
	var hc pkg.HashConfig

	if err := hc.Compare(user.Password, login.Password); err != nil {
		return "", err
	}
	claims := pkg.NewClaims(login.Id, user.Email)
	token, err := claims.GenJWT()
	if err != nil {
		return "", err
	}
	return token, nil

}

func (a *AuthService) Logout(ctx context.Context, user dto.NewUser) (string, error) {

	login, err := a.authRepo.Login(ctx, user.Email)
	if err != nil {
		return "", err
	}
	var hc pkg.HashConfig

	if err := hc.Compare(user.Password, login.Password); err != nil {
		return "", err
	}
	claims := pkg.NewClaims(login.Id, user.Email)
	token, err := claims.GenJWT()
	if err != nil {
		return "", err
	}
	return token, nil

}
