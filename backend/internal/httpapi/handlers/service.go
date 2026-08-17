package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

type Service struct {
	repository domain.ServiceRepository
	logger     *slog.Logger
}

func NewService(repository domain.ServiceRepository, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{repository: repository, logger: logger}
}

func (handler *Service) List(ctx *gin.Context) {
	options, err := parseListOptions(ctx)
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}

	result, err := handler.repository.ListServices(ctx.Request.Context(), options)
	if err != nil {
		handler.logger.ErrorContext(ctx.Request.Context(), "falha ao listar serviços", "error", err)
		writeError(ctx, http.StatusInternalServerError, "internal_error", "Não foi possível processar a solicitação.")
		return
	}
	ctx.JSON(http.StatusOK, newListResponse(result.Items, result.Total, options))
}
