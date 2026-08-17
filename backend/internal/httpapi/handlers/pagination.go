package handlers

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

const (
	defaultPage     int64 = 1
	defaultPageSize int64 = 20
	maximumPageSize int64 = 100
)

var errInvalidPagination = errors.New("parâmetros de paginação inválidos")

func parseListOptions(ctx *gin.Context) (domain.ListOptions, error) {
	page, err := parsePositiveParameter(ctx, "page", defaultPage)
	if err != nil {
		return domain.ListOptions{}, err
	}
	pageSize, err := parsePositiveParameter(ctx, "page_size", defaultPageSize)
	if err != nil {
		return domain.ListOptions{}, err
	}
	if pageSize > maximumPageSize {
		return domain.ListOptions{}, errors.New("page_size deve ser menor ou igual a 100")
	}
	if page-1 > math.MaxInt64/pageSize {
		return domain.ListOptions{}, errInvalidPagination
	}
	return domain.ListOptions{
		Page:     page,
		PageSize: pageSize,
		Query:    strings.TrimSpace(ctx.Query("q")),
	}, nil
}

func parsePositiveParameter(ctx *gin.Context, name string, fallback int64) (int64, error) {
	raw, exists := ctx.GetQuery(name)
	if !exists {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 1 {
		return 0, errors.New(name + " deve ser um número inteiro maior ou igual a 1")
	}
	return value, nil
}
