package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/repository"
)

// RequirePin is a middleware that:
//  1. Reads the authenticated user's email from the JWT claims (set by VerifyTokenWithDB).
//  2. Looks up the user's PIN record.
//  3. If the user has no PIN yet (empty pin_hash) → 403 with a prompt to create one first.
//  4. If a PIN exists → reads the X-Pin header, verifies it against the stored argon2id
//     hash using pkg.HashConfig.Compare, and aborts with 401 on mismatch.
//
// Place this middleware AFTER VerifyTokenWithDB on any route that requires PIN
// confirmation (e.g. transfers, withdrawals).
func RequirePin(db *pgxpool.Pool) gin.HandlerFunc {
	pinRepo := repository.NewPinRepo(db)

	return func(ctx *gin.Context) {
		// --- 1. Pull claims that VerifyTokenWithDB already stored ---
		raw, exists := ctx.Get("claims")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Message: "Unauthorized",
				Success: false,
				Error:   "Missing JWT claims — place RequirePin after VerifyTokenWithDB",
			})
			return
		}
		claims, ok := raw.(pkg.Claims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.Response{
				Message: "Error",
				Success: false,
				Error:   "Internal Server Error",
			})
			return
		}

		// --- 2. Fetch the PIN record for this user ---
		userPin, err := pinRepo.GetPinByEmail(context.Background(), claims.Email)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, dto.Response{
				Message: "Error",
				Success: false,
				Error:   "Failed to fetch PIN record",
			})
			return
		}

		// --- 3. No PIN set yet → reject and tell the client to set one ---
		if userPin.PinHash == nil || *userPin.PinHash == "" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, dto.Response{
				Message: "PIN not set",
				Success: false,
				Error:   "You must create a PIN before performing this action",
			})
			return
		}

		// --- 4. PIN exists → require X-Pin header and verify ---
		pin := ctx.GetHeader("X-Pin")
		if pin == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Message: "PIN required",
				Success: false,
				Error:   "Provide your 6-digit PIN in the X-Pin header",
			})
			return
		}

		var hc pkg.HashConfig
		hc.UseRecommended()
		if err := hc.Compare(pin, *userPin.PinHash); err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Message: "Invalid PIN",
				Success: false,
				Error:   "The PIN you entered is incorrect",
			})
			return
		}

		ctx.Next()
	}
}
