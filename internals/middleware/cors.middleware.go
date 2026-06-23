package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(ctx *gin.Context) {
	origin := ctx.GetHeader("Origin")
	allowedOrigins := getAllowedOrigins()

	if origin != "" && isOriginAllowed(origin, allowedOrigins) {
		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Vary", "Origin")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		requestHeaders := ctx.GetHeader("Access-Control-Request-Headers")
		if requestHeaders != "" {
			ctx.Header("Access-Control-Allow-Headers", requestHeaders)
		} else {
			ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		}
	}

	if ctx.Request.Method == http.MethodOptions {
		if origin != "" && !isOriginAllowed(origin, allowedOrigins) {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	ctx.Next()
}

func getAllowedOrigins() map[string]bool {
	rawOrigins := os.Getenv("ALLOWED_ORIGINS")

	allowedOrigins := make(map[string]bool)

	for _, origin := range strings.Split(rawOrigins, ",") {
		normalizedOrigin := strings.TrimSpace(origin)

		if normalizedOrigin == "" {
			continue
		}

		allowedOrigins[normalizedOrigin] = true
	}

	return allowedOrigins
}

func isOriginAllowed(origin string, allowedOrigins map[string]bool) bool {
	if origin == "" {
		return false
	}

	return allowedOrigins[origin]
}
