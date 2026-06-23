package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/cache"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
)

type AuthRepo interface {
	Register(ctx context.Context, email, password string) (model.User, error)
	Login(ctx context.Context, email string) (model.User, error)
	GetUserInfo(ctx context.Context, email string) (model.Profile, error)
	GetUserByResetToken(ctx context.Context, rawToken string) (model.User, error)
	SaveToken(ctx context.Context, userID uuid.UUID, rawToken string, tokenType model.TokenType, expiresAt time.Time) error
	RevokeToken(ctx context.Context, rawToken string) error
	IsTokenValid(ctx context.Context, rawToken string) (bool, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
}

type AuthSession struct {
	Token string
	User  dto.UserResponse
}

type AuthService struct {
	authRepo AuthRepo
	rdb      *redis.Client
}

func NewAuthService(authRepo AuthRepo, rdb *redis.Client) *AuthService {
	return &AuthService{authRepo: authRepo, rdb: rdb}
}

func normalizeAuthEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func authUserCacheKey(email string) string {
	return "vando:user:" + normalizeAuthEmail(email)
}

func (a *AuthService) Register(ctx context.Context, user dto.RegisterRequest) (dto.UserResponse, error) {
	email := normalizeAuthEmail(user.Email)

	var hc pkg.HashConfig
	hc.UseRecommended()
	hashedPwd := hc.GenHash(user.Password)
	result, err := a.authRepo.Register(ctx, email, hashedPwd)
	if err != nil {
		return dto.UserResponse{}, err
	}
	return dto.UserResponse{ID: result.ID, Email: result.Email}, nil
}

func (a *AuthService) GetMe(ctx context.Context, email string) (model.Profile, error) {
	return a.authRepo.GetUserInfo(ctx, normalizeAuthEmail(email))
}

func (a *AuthService) Login(ctx context.Context, user dto.LoginRequest) (AuthSession, error) {
	email := normalizeAuthEmail(user.Email)
	login, err := a.getOrFetchUser(ctx, email)
	if err != nil {
		return AuthSession{}, err
	}

	var hc pkg.HashConfig
	if err := hc.Compare(user.Password, login.Password); err != nil {
		return AuthSession{}, err
	}

	claims := pkg.NewClaims(login.ID, login.Email)
	token, err := claims.GenJWT()
	if err != nil {
		return AuthSession{}, err
	}
	expiresAt := time.Now().Add(pkg.AccessTokenExpiry)
	if err := a.authRepo.SaveToken(
		ctx,
		login.ID,
		token,
		model.TokenTypeAccess,
		expiresAt,
	); err != nil {
		return AuthSession{}, err
	}

	return AuthSession{
		Token: token,
		User: dto.UserResponse{
			ID:    login.ID,
			Email: login.Email,
		},
	}, nil
}

func (a *AuthService) ResetPassword(ctx context.Context, user dto.ResetPasswordRequest) (string, error) {
	email := normalizeAuthEmail(user.Email)
	login, err := a.getOrFetchUser(ctx, email)
	if err != nil {
		return "", err
	}
	token, err := generateResetToken(32)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(pkg.ResetTokenExpiry)
	if err := a.authRepo.SaveToken(
		ctx,
		login.ID,
		token,
		model.TokenTypePasswordReset,
		expiresAt,
	); err != nil {
		return "", err
	}

	return token, nil
}

func (a *AuthService) ConfirmResetPassword(ctx context.Context, user dto.ConfirmResetPassword) (string, error) {
	foundUser, err := a.authRepo.GetUserByResetToken(ctx, user.Token)
	if err != nil {
		return "", err
	}

	// Issue a short-lived JWT scoped exclusively for the change-password endpoint.
	claims := pkg.NewResetClaims(foundUser.ID, foundUser.Email)
	resetJWT, err := claims.GenJWT()
	if err != nil {
		return "", err
	}

	return resetJWT, nil
}

// ChangeResetPassword hashes newPassword and persists it for the user identified by
// the password-reset JWT claims.
func (a *AuthService) ChangeResetPassword(ctx context.Context, userID uuid.UUID, email, newPassword string) error {
	var hc pkg.HashConfig
	hc.UseRecommended()
	hashed := hc.GenHash(newPassword)
	if err := a.authRepo.UpdatePassword(ctx, userID, hashed); err != nil {
		return err
	}

	if err := cache.DelFromCache(ctx, a.rdb, authUserCacheKey(email)); err != nil {
		log.Println("cache evict error after reset password change:", err)
	}

	return nil
}

func (a *AuthService) getOrFetchUser(ctx context.Context, email string) (*model.User, error) {
	email = normalizeAuthEmail(email)
	rkey := authUserCacheKey(email)

	var user model.User
	if err := cache.GetFromCache(ctx, a.rdb, rkey, &user); err == nil {
		log.Println("cache hit:", email)
		return &user, nil
	} else if !errors.Is(err, redis.Nil) {
		log.Println("redis error:", err)
	}

	log.Println("cache miss:", email)
	fetched, err := a.authRepo.Login(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := cache.SaveToCache(ctx, a.rdb, rkey, fetched); err != nil {
		log.Println("cache save error:", err) // non-fatal
	}

	return &fetched, nil
}

func (a *AuthService) Logout(ctx context.Context, rawToken, email string) error {
	if err := a.authRepo.RevokeToken(ctx, rawToken); err != nil {
		return err
	}
	if err := cache.DelFromCache(ctx, a.rdb, authUserCacheKey(email)); err != nil {
		log.Println("cache evict error on logout:", err) // non-fatal
	}
	return nil
}

func generateResetToken(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
