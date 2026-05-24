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
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Message: "Unauthorized",
				Success: false,
				Error:   "Missing Authorization header",
			})
			return
		}

		parts := strings.Split(bearerToken, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Message: "Unauthorized",
				Success: false,
				Error:   "Invalid token format, use: Bearer <token>",
			})
			return
		}
		rawToken := parts[1]

		// 1. Verify JWT signature + claims
		var claims pkg.Claims
		if err := claims.VerifyJWT(rawToken); err != nil {
			log.Println("[VerifyToken] JWT error:", err)
			if errors.Is(err, jwt.ErrTokenInvalidIssuer) || errors.Is(err, jwt.ErrTokenExpired) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
					Message: "Unauthorized, please login again",
					Success: false,
					Error:   err.Error(),
				})
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.Response{
				Message: "Error",
				Success: false,
				Error:   "Internal Server Error",
			})
			return
		}

		// 2. Check tokens table — must be active (not revoked, not expired)
		valid, err := authRepo.IsTokenValid(context.Background(), rawToken)
		if err != nil {
			log.Println("[VerifyToken] DB token check error:", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.Response{
				Message: "Error",
				Success: false,
				Error:   "Internal Server Error",
			})
			return
		}
		if !valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Message: "Token has been revoked or expired, please login again",
				Success: false,
				Error:   "Token invalid",
			})
			return
		}

		ctx.Set("claims", claims)
		ctx.Set("raw_token", rawToken)
		ctx.Next()
	}
}
