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
	return &AuthService{
		authRepo: authRepo,
	}
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
		Password:  newUser.Password,
		CreatedAt: newUser.CreatedAt,
	}, nil
}
