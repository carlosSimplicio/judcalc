package httpapi

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const defaultCORSAllowedOrigin = "http://localhost:3000"

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := corsAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))

	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if _, allowed := allowedOrigins[origin]; !allowed {
			ctx.Next()
			return
		}

		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		ctx.Header("Vary", "Origin")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}

func corsAllowedOrigins(configured string) map[string]struct{} {
	if strings.TrimSpace(configured) == "" {
		configured = defaultCORSAllowedOrigin
	}

	origins := make(map[string]struct{})
	for origin := range strings.SplitSeq(configured, ",") {
		if normalized := strings.TrimSpace(origin); normalized != "" {
			origins[normalized] = struct{}{}
		}
	}
	return origins
}
