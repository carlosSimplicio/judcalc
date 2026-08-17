package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

type Area struct {
	repository domain.AreaRepository
	logger     *slog.Logger
}

func NewArea(repository domain.AreaRepository, logger *slog.Logger) *Area {
	if logger == nil {
		logger = slog.Default()
	}
	return &Area{repository: repository, logger: logger}
}

func (handler *Area) List(ctx *gin.Context) {
	options, err := parseListOptions(ctx)
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}

	result, err := handler.repository.ListAreas(ctx.Request.Context(), options)
	if err != nil {
		handler.logger.ErrorContext(ctx.Request.Context(), "falha ao listar áreas", "error", err)
		writeError(ctx, http.StatusInternalServerError, "internal_error", "Não foi possível processar a solicitação.")
		return
	}
	ctx.JSON(http.StatusOK, newListResponse(result.Items, result.Total, options))
}
