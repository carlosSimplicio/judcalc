package httpapi

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startedAt := time.Now()

		ctx.Next()

		logger.InfoContext(
			ctx.Request.Context(),
			"requisição HTTP",
			"method", ctx.Request.Method,
			"path", ctx.Request.URL.Path,
			"status", ctx.Writer.Status(),
			"duration", time.Since(startedAt),
			"client_ip", ctx.ClientIP(),
		)
	}
}
