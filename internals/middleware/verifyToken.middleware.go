package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/repository"
)

// VerifyTokenWithDB validates the JWT signature and checks the tokens table
// (token must exist, is_revoked = false, expires_at > now()).
func VerifyTokenWithDB(db *pgxpool.Pool) gin.HandlerFunc {
	authRepo := repository.NewAuthRepo(db)

	return func(ctx *gin.Context) {
		bearerToken := ctx.GetHeader("Authorization")
		if bearerToken == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing Authorization header"))
			return
		}

		parts := strings.Split(bearerToken, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Invalid token format, use: Bearer <token>"))
			return
		}
		rawToken := parts[1]

		var claims pkg.Claims
		if err := claims.VerifyJWT(rawToken); err != nil {
			log.Println("[VerifyToken] JWT error:", err)
			if errors.Is(err, jwt.ErrTokenInvalidIssuer) || errors.Is(err, jwt.ErrTokenExpired) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewError("Unauthorized, please login again", err.Error()))
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.NewError("Error", "Internal server error"))
			return
		}

		valid, err := authRepo.IsTokenValid(context.Background(), rawToken)
		if err != nil {
			log.Println("[VerifyToken] DB token check error:", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.NewError("Error", "Internal server error"))
			return
		}
		if !valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewError("Token has been revoked or expired, please login again", "Token invalid"))
			return
		}

		ctx.Set("claims", claims)
		ctx.Set("raw_token", rawToken)
		ctx.Next()
	}
}
