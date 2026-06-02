package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(ctx *gin.Context) {

	allowedOrigin := []string{
		"http://127.0.0.1:5500",
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://localhost:5500",
		"http://localhost",       
		"http://localhost:80",
		"http://vanwallet_frontend", 
	}
	currentOrigin := ctx.GetHeader("Origin")

	if slices.Contains(allowedOrigin, currentOrigin) {
		ctx.Header("Access-Control-Allow-Origin", currentOrigin)
	} else if currentOrigin == "" {
		// Same-host proxy (Nginx → backend): no Origin header, allow freely.
		ctx.Header("Access-Control-Allow-Origin", "*")
	}

	allowedMethods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions}
	ctx.Header("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
	ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	ctx.Header("Access-Control-Allow-Credentials", "true")

	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	ctx.Next()
}
