package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

type AuthRepository struct {
	database *sql.DB
}

func NewAuthRepository(database *sql.DB) *AuthRepository {
	return &AuthRepository{database: database}
}

func (repository *AuthRepository) CreateUserWithToken(ctx context.Context, input domain.NewUser, tokenHash string, createdAt, expiresAt time.Time) (domain.User, error) {
	transaction, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return domain.User{}, fmt.Errorf("iniciar cadastro: %w", err)
	}
	defer transaction.Rollback()

	result, err := transaction.ExecContext(ctx, `
		INSERT INTO users(email, name, password_hash, created_at)
		VALUES (?, ?, ?, ?)
	`, input.Email, input.Name, input.PasswordHash, createdAt.Unix())
	if err != nil {
		exists, lookupErr := userEmailExists(ctx, transaction, input.Email)
		if lookupErr == nil && exists {
			return domain.User{}, domain.ErrEmailAlreadyExists
		}
		return domain.User{}, fmt.Errorf("criar usuário: %w", err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return domain.User{}, fmt.Errorf("obter id do usuário: %w", err)
	}
	if err := insertToken(ctx, transaction, userID, tokenHash, createdAt, expiresAt); err != nil {
		return domain.User{}, err
	}
	if err := transaction.Commit(); err != nil {
		return domain.User{}, fmt.Errorf("confirmar cadastro: %w", err)
	}
	return domain.User{ID: userID, Email: input.Email, Name: input.Name}, nil
}

func (repository *AuthRepository) FindUserByEmail(ctx context.Context, email string) (domain.UserCredentials, bool, error) {
	var credentials domain.UserCredentials
	err := repository.database.QueryRowContext(ctx, `
		SELECT id, email, name, password_hash FROM users WHERE email = ?
	`, email).Scan(&credentials.ID, &credentials.Email, &credentials.Name, &credentials.PasswordHash)
	if err == sql.ErrNoRows {
		return domain.UserCredentials{}, false, nil
	}
	if err != nil {
		return domain.UserCredentials{}, false, fmt.Errorf("buscar credenciais: %w", err)
	}
	return credentials, true, nil
}

func (repository *AuthRepository) CreateToken(ctx context.Context, userID int64, tokenHash string, createdAt, expiresAt time.Time) error {
	if err := insertToken(ctx, repository.database, userID, tokenHash, createdAt, expiresAt); err != nil {
		return err
	}
	return nil
}

func (repository *AuthRepository) FindUserByTokenHash(ctx context.Context, tokenHash string, now time.Time) (domain.User, bool, error) {
	var user domain.User
	err := repository.database.QueryRowContext(ctx, `
		SELECT users.id, users.email, users.name
		FROM auth_tokens
		JOIN users ON users.id = auth_tokens.user_id
		WHERE auth_tokens.token_hash = ? AND auth_tokens.expires_at > ?
	`, tokenHash, now.Unix()).Scan(&user.ID, &user.Email, &user.Name)
	if err == sql.ErrNoRows {
		return domain.User{}, false, nil
	}
	if err != nil {
		return domain.User{}, false, fmt.Errorf("validar token: %w", err)
	}
	return user, true, nil
}

type tokenExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func insertToken(ctx context.Context, executor tokenExecutor, userID int64, tokenHash string, createdAt, expiresAt time.Time) error {
	_, err := executor.ExecContext(ctx, `
		INSERT INTO auth_tokens(user_id, token_hash, created_at, expires_at)
		VALUES (?, ?, ?, ?)
	`, userID, tokenHash, createdAt.Unix(), expiresAt.Unix())
	if err != nil {
		return fmt.Errorf("criar token: %w", err)
	}
	return nil
}

func userEmailExists(ctx context.Context, transaction *sql.Tx, email string) (bool, error) {
	var exists bool
	err := transaction.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
	return exists, err
}
