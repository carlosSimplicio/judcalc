package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/carlosSimplicio/judcalc/backend/internal/auth"
	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

type AuthenticationService interface {
	SignUp(context.Context, string, string, string) (auth.Session, error)
	SignIn(context.Context, string, string) (auth.Session, error)
}

type Authentication struct {
	service AuthenticationService
	logger  *slog.Logger
}

func NewAuthentication(service AuthenticationService, logger *slog.Logger) *Authentication {
	if logger == nil {
		logger = slog.Default()
	}
	return &Authentication{service: service, logger: logger}
}

type signUpRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type sessionResponse struct {
	User        domain.User `json:"user"`
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresAt   time.Time   `json:"expires_at"`
}

func (handler *Authentication) SignUp(ctx *gin.Context) {
	var request signUpRequest
	if err := decodeStrictJSON(ctx, &request); err != nil {
		writeJSONDecodeError(ctx, err)
		return
	}
	session, err := handler.service.SignUp(ctx.Request.Context(), request.Email, request.Name, request.Password)
	if err != nil {
		handler.writeSignUpError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"data": newSessionResponse(session)})
}

func (handler *Authentication) SignIn(ctx *gin.Context) {
	var request signInRequest
	if err := decodeStrictJSON(ctx, &request); err != nil {
		writeJSONDecodeError(ctx, err)
		return
	}
	session, err := handler.service.SignIn(ctx.Request.Context(), request.Email, request.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(ctx, http.StatusUnauthorized, "invalid_credentials", "Email ou senha inválidos.")
			return
		}
		handler.logger.ErrorContext(ctx.Request.Context(), "falha ao realizar sign-in", "error", err)
		writeError(ctx, http.StatusInternalServerError, "internal_error", "Não foi possível processar a solicitação.")
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": newSessionResponse(session)})
}

func (handler *Authentication) writeSignUpError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidRegistration):
		writeError(ctx, http.StatusBadRequest, "invalid_body", "Email, nome ou senha inválidos.")
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		writeError(ctx, http.StatusConflict, "email_already_exists", "Já existe um usuário com este email.")
	default:
		handler.logger.ErrorContext(ctx.Request.Context(), "falha ao cadastrar usuário", "error", err)
		writeError(ctx, http.StatusInternalServerError, "internal_error", "Não foi possível processar a solicitação.")
	}
}

func newSessionResponse(session auth.Session) sessionResponse {
	return sessionResponse{User: session.User, AccessToken: session.AccessToken, TokenType: "Bearer", ExpiresAt: session.ExpiresAt}
}
