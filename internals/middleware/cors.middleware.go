package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(ctx *gin.Context) {
	allowedOrigin := []string{"http://127.0.0.1:5500"}
	currentOrigin := ctx.GetHeader("Origin")
	if slices.Contains(allowedOrigin, currentOrigin) {
		ctx.Header("Access-Control-Allow-Origin", currentOrigin)
	}

	allowedMethods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions}
	ctx.Header("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))

	ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	ctx.Next()
}
