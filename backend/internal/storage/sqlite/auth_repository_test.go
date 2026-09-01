package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

const authTestSchema = `
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    email TEXT NOT NULL COLLATE NOCASE UNIQUE,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at INTEGER NOT NULL
);
CREATE TABLE auth_tokens (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL
);`

func TestAuthRepositoryCreatesUserAndTokenAtomically(t *testing.T) {
	repository, database := newAuthTestRepository(t)
	createdAt := time.Unix(1_700_000_000, 0).UTC()
	expiresAt := createdAt.Add(30 * 24 * time.Hour)

	user, err := repository.CreateUserWithToken(context.Background(), domain.NewUser{
		Email: "maria@example.com", Name: "Maria", PasswordHash: "$2a$10$hash",
	}, "a-token-hash", createdAt, expiresAt)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID <= 0 || user.Email != "maria@example.com" || user.Name != "Maria" {
		t.Fatalf("unexpected user: %#v", user)
	}

	var storedUserID int64
	if err := database.QueryRow("SELECT user_id FROM auth_tokens WHERE token_hash = ?", "a-token-hash").Scan(&storedUserID); err != nil {
		t.Fatal(err)
	}
	if storedUserID != user.ID {
		t.Fatalf("token user_id = %d, want %d", storedUserID, user.ID)
	}
}

func TestAuthRepositoryRejectsDuplicateEmailWithoutCreatingToken(t *testing.T) {
	repository, database := newAuthTestRepository(t)
	now := time.Unix(1_700_000_000, 0).UTC()
	_, err := repository.CreateUserWithToken(context.Background(), domain.NewUser{
		Email: "maria@example.com", Name: "Maria", PasswordHash: "hash",
	}, "first-token", now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.CreateUserWithToken(context.Background(), domain.NewUser{
		Email: "MARIA@EXAMPLE.COM", Name: "Outra", PasswordHash: "hash",
	}, "second-token", now, now.Add(time.Hour))
	if !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Fatalf("error = %v, want ErrEmailAlreadyExists", err)
	}
	var count int
	if err := database.QueryRow("SELECT count(*) FROM auth_tokens").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("token count = %d, want 1", count)
	}
}

func TestAuthRepositoryRollsBackUserWhenTokenCannotBeCreated(t *testing.T) {
	repository, database := newAuthTestRepository(t)
	now := time.Unix(1_700_000_000, 0).UTC()
	_, err := repository.CreateUserWithToken(context.Background(), domain.NewUser{
		Email: "first@example.com", Name: "First", PasswordHash: "hash",
	}, "same-token", now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.CreateUserWithToken(context.Background(), domain.NewUser{
		Email: "second@example.com", Name: "Second", PasswordHash: "hash",
	}, "same-token", now, now.Add(time.Hour))
	if err == nil {
		t.Fatal("duplicate token should fail")
	}
	var count int
	if err := database.QueryRow("SELECT count(*) FROM users WHERE email = 'second@example.com'").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rolled back user count = %d, want 0", count)
	}
}

func TestAuthRepositoryFindsCredentialsAndValidToken(t *testing.T) {
	repository, _ := newAuthTestRepository(t)
	now := time.Unix(1_700_000_000, 0).UTC()
	created, err := repository.CreateUserWithToken(context.Background(), domain.NewUser{
		Email: "maria@example.com", Name: "Maria", PasswordHash: "stored-hash",
	}, "valid-token", now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	credentials, found, err := repository.FindUserByEmail(context.Background(), "maria@example.com")
	if err != nil || !found || credentials.PasswordHash != "stored-hash" || credentials.User != created {
		t.Fatalf("credentials = %#v, found = %v, err = %v", credentials, found, err)
	}
	authenticated, found, err := repository.FindUserByTokenHash(context.Background(), "valid-token", now.Add(30*time.Minute))
	if err != nil || !found || authenticated != created {
		t.Fatalf("user = %#v, found = %v, err = %v", authenticated, found, err)
	}
}

func TestAuthRepositoryRejectsExpiredTokenAndCreatesAdditionalSession(t *testing.T) {
	repository, database := newAuthTestRepository(t)
	now := time.Unix(1_700_000_000, 0).UTC()
	user, err := repository.CreateUserWithToken(context.Background(), domain.NewUser{
		Email: "maria@example.com", Name: "Maria", PasswordHash: "hash",
	}, "expired-token", now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	if authenticated, found, err := repository.FindUserByTokenHash(context.Background(), "expired-token", now.Add(time.Hour)); err != nil || found || authenticated.ID != 0 {
		t.Fatalf("expired token returned user = %#v, found = %v, err = %v", authenticated, found, err)
	}
	if err := repository.CreateToken(context.Background(), user.ID, "new-token", now.Add(time.Hour), now.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := database.QueryRow("SELECT count(*) FROM auth_tokens WHERE user_id = ?", user.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("token count = %d, want 2", count)
	}
}

func newAuthTestRepository(t *testing.T) (*AuthRepository, *sql.DB) {
	t.Helper()
	database, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec("PRAGMA foreign_keys = ON"); err != nil {
		database.Close()
		t.Fatal(err)
	}
	if _, err := database.Exec(authTestSchema); err != nil {
		database.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return NewAuthRepository(database), database
}
