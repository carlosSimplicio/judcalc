package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/carlosSimplicio/judcalc/backend/internal/auth"
	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
	"github.com/carlosSimplicio/judcalc/backend/internal/httpapi/handlers"
)

type AuthService interface {
	handlers.AuthenticationService
	Authenticate(context.Context, string) (domain.User, error)
}

func requireAuthentication(service AuthService, logger *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		parts := strings.Fields(ctx.GetHeader("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			handlers.WriteError(ctx, http.StatusUnauthorized, "unauthorized", "Autenticação obrigatória.")
			return
		}
		user, err := service.Authenticate(ctx.Request.Context(), parts[1])
		if err != nil {
			if errors.Is(err, auth.ErrInvalidToken) {
				handlers.WriteError(ctx, http.StatusUnauthorized, "unauthorized", "Token inválido ou expirado.")
				return
			}
			logger.ErrorContext(ctx.Request.Context(), "falha ao validar token", "error", err)
			handlers.WriteError(ctx, http.StatusInternalServerError, "internal_error", "Não foi possível processar a solicitação.")
			return
		}
		ctx.Set(handlers.AuthenticatedUserKey, user)
		ctx.Next()
	}
}
