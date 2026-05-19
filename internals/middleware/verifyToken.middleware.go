package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
)

func VerifyToken(ctx *gin.Context) {
	bearerToken := ctx.GetHeader("Authorization")
	if bearerToken == "" {
		ctx.AbortWithStatusJSON(
			http.StatusUnauthorized, dto.Response{
				Message: "token expired",
				Success: false,
				Error:   "login Lagi",
			},
		)
		return
	}
	splitedbearer := strings.Split(bearerToken, " ")
	if len(splitedbearer) != 2 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Message: "Unauthorized Access, Please Login",
			Success: false,
			Error:   "invalid token",
		})
		return
	}
	token := splitedbearer[1]
	var claims pkg.Claims
	if err := claims.VerifyJWT(token); err != nil {
		log.Println("Error: ", err.Error())
		if errors.Is(err, jwt.ErrTokenInvalidIssuer) || errors.Is(err, jwt.ErrTokenExpired) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Message: "Unauthorized Access, Please Login",
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

	// menempelkan (attach) claims ke context request
	ctx.Set("claims", claims)
	ctx.Next()
}
