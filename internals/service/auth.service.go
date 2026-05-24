package service

import (
	"context"
	"time"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
)

type AuthRepo interface {
	Register(ctx context.Context, email, password string) (model.User, error)
	Login(ctx context.Context, email string) (model.User, error)
	GetUserPin(ctx context.Context, email string) (model.UserPin, error)
	SaveToken(ctx context.Context, userID, rawToken string, tokenType model.TokenType, expiresAt time.Time) error
	RevokeToken(ctx context.Context, rawToken string) error
	IsTokenValid(ctx context.Context, rawToken string) (bool, error)
}

type AuthService struct {
	authRepo AuthRepo
}

func NewAuthService(authRepo AuthRepo) *AuthService {
	return &AuthService{authRepo: authRepo}
}

func (a *AuthService) Register(ctx context.Context, user dto.RegisterRequest) (dto.UserResponse, error) {
	var hc pkg.HashConfig
	hc.UseRecommended()
	hashedPwd := hc.GenHash(user.Password)

	result, err := a.authRepo.Register(ctx, user.Email, hashedPwd)
	if err != nil {
		return dto.UserResponse{}, err
	}
	return dto.UserResponse{ID: result.ID, Email: result.Email}, nil
}

// Login authenticates the user and persists the issued JWT in the tokens table.
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

	expiresAt := time.Now().Add(pkg.AccessTokenExpiry)
	if err := a.authRepo.SaveToken(
		ctx,
		login.ID.String(),
		token,
		model.TokenTypeRefresh, // REFRESH tracks active sessions
		expiresAt,
	); err != nil {
		return "", err
	}

	return token, nil
}

// Logout revokes the current token (marks is_revoked = true).
func (a *AuthService) Logout(ctx context.Context, rawToken string) error {
	return a.authRepo.RevokeToken(ctx, rawToken)
}

// IsTokenValid checks the tokens table — token must exist, not revoked, and not expired.
func (a *AuthService) IsTokenValid(ctx context.Context, rawToken string) (bool, error) {
	return a.authRepo.IsTokenValid(ctx, rawToken)
}

func (a *AuthService) GetUserPin(ctx context.Context, email string) (model.UserPin, error) {
	return a.authRepo.GetUserPin(ctx, email)
}
