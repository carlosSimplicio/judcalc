package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

const (
	tokenBytes    = 32
	tokenValidity = 30 * 24 * time.Hour
)

var (
	ErrInvalidRegistration = errors.New("invalid registration")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidToken        = errors.New("invalid token")
)

type Session struct {
	User        domain.User
	AccessToken string
	ExpiresAt   time.Time
}

type Service struct {
	repository domain.AuthRepository
	now        func() time.Time
	random     io.Reader
}

func NewService(repository domain.AuthRepository) *Service {
	return &Service{repository: repository, now: time.Now, random: rand.Reader}
}

func (service *Service) SignUp(ctx context.Context, email, name, password string) (Session, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	normalizedName := strings.TrimSpace(name)
	if !validEmail(normalizedEmail) || normalizedName == "" || len(password) < 8 || len(password) > 72 {
		return Session{}, ErrInvalidRegistration
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Session{}, fmt.Errorf("gerar hash da senha: %w", err)
	}
	rawToken, tokenHash, err := service.newToken()
	if err != nil {
		return Session{}, err
	}
	createdAt := service.now().UTC().Truncate(time.Second)
	expiresAt := createdAt.Add(tokenValidity)
	user, err := service.repository.CreateUserWithToken(ctx, domain.NewUser{
		Email: normalizedEmail, Name: normalizedName, PasswordHash: string(passwordHash),
	}, tokenHash, createdAt, expiresAt)
	if err != nil {
		return Session{}, err
	}
	return Session{User: user, AccessToken: rawToken, ExpiresAt: expiresAt}, nil
}

func (service *Service) SignIn(ctx context.Context, email, password string) (Session, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if !validEmail(normalizedEmail) || password == "" {
		return Session{}, ErrInvalidCredentials
	}
	credentials, found, err := service.repository.FindUserByEmail(ctx, normalizedEmail)
	if err != nil {
		return Session{}, err
	}
	if !found || bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(password)) != nil {
		return Session{}, ErrInvalidCredentials
	}
	rawToken, tokenHash, err := service.newToken()
	if err != nil {
		return Session{}, err
	}
	createdAt := service.now().UTC().Truncate(time.Second)
	expiresAt := createdAt.Add(tokenValidity)
	if err := service.repository.CreateToken(ctx, credentials.ID, tokenHash, createdAt, expiresAt); err != nil {
		return Session{}, err
	}
	return Session{User: credentials.User, AccessToken: rawToken, ExpiresAt: expiresAt}, nil
}

func (service *Service) Authenticate(ctx context.Context, rawToken string) (domain.User, error) {
	if rawToken == "" {
		return domain.User{}, ErrInvalidToken
	}
	user, found, err := service.repository.FindUserByTokenHash(ctx, hashToken(rawToken), service.now().UTC())
	if err != nil {
		return domain.User{}, err
	}
	if !found {
		return domain.User{}, ErrInvalidToken
	}
	return user, nil
}

func (service *Service) newToken() (string, string, error) {
	buffer := make([]byte, tokenBytes)
	if _, err := io.ReadFull(service.random, buffer); err != nil {
		return "", "", fmt.Errorf("gerar token: %w", err)
	}
	rawToken := base64.RawURLEncoding.EncodeToString(buffer)
	return rawToken, hashToken(rawToken), nil
}

func hashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}

func validEmail(email string) bool {
	if len(email) == 0 || len(email) > 254 {
		return false
	}
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}
