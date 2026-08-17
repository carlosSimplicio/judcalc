package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

type metadata struct {
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

type listResponse[T any] struct {
	Data []T      `json:"data"`
	Meta metadata `json:"meta"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

func newListResponse[T any](items []T, total int64, options domain.ListOptions) listResponse[T] {
	totalPages := int64(0)
	if total > 0 {
		totalPages = total / options.PageSize
		if total%options.PageSize != 0 {
			totalPages++
		}
	}
	return listResponse[T]{
		Data: items,
		Meta: metadata{
			Page:       options.Page,
			PageSize:   options.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

func WriteError(ctx *gin.Context, status int, code, message string) {
	writeError(ctx, status, code, message)
}

func writeError(ctx *gin.Context, status int, code, message string) {
	ctx.AbortWithStatusJSON(status, errorResponse{Error: errorBody{Code: code, Message: message}})
}
