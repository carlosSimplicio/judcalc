package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

type repositoryStub struct {
	createdInput domain.NewUser
	createdUser  domain.User
	credentials  domain.UserCredentials
	found        bool
	tokenHash    string
	tokenUserID  int64
	tokenUser    domain.User
	createdAt    time.Time
	expiresAt    time.Time
}

func (stub *repositoryStub) CreateUserWithToken(_ context.Context, input domain.NewUser, tokenHash string, createdAt, expiresAt time.Time) (domain.User, error) {
	stub.createdInput = input
	stub.tokenHash = tokenHash
	stub.createdAt = createdAt
	stub.expiresAt = expiresAt
	return stub.createdUser, nil
}

func (stub *repositoryStub) FindUserByEmail(context.Context, string) (domain.UserCredentials, bool, error) {
	return stub.credentials, stub.found, nil
}

func (stub *repositoryStub) CreateToken(_ context.Context, userID int64, tokenHash string, createdAt, expiresAt time.Time) error {
	stub.tokenUserID = userID
	stub.tokenHash = tokenHash
	stub.createdAt = createdAt
	stub.expiresAt = expiresAt
	return nil
}

func (stub *repositoryStub) FindUserByTokenHash(_ context.Context, tokenHash string, _ time.Time) (domain.User, bool, error) {
	stub.tokenHash = tokenHash
	return stub.tokenUser, stub.found, nil
}

func TestSignUpNormalizesIdentityHashesPasswordAndCreatesThirtyDaySession(t *testing.T) {
	now := time.Unix(1_700_000_000, 123_000_000).UTC()
	repository := &repositoryStub{createdUser: domain.User{ID: 7, Email: "maria@example.com", Name: "Maria Silva"}}
	service := newTestService(repository, now)

	session, err := service.SignUp(context.Background(), " MARIA@Example.COM ", " Maria Silva ", "password-123")
	if err != nil {
		t.Fatal(err)
	}
	if repository.createdInput.Email != "maria@example.com" || repository.createdInput.Name != "Maria Silva" {
		t.Fatalf("input was not normalized: %#v", repository.createdInput)
	}
	if repository.createdInput.PasswordHash == "password-123" || bcrypt.CompareHashAndPassword([]byte(repository.createdInput.PasswordHash), []byte("password-123")) != nil {
		t.Fatal("password was not stored as a valid bcrypt hash")
	}
	if session.User != repository.createdUser || session.AccessToken != "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" {
		t.Fatalf("unexpected session: %#v", session)
	}
	if repository.tokenHash == session.AccessToken || len(repository.tokenHash) != 64 {
		t.Fatalf("unsafe token hash: %q", repository.tokenHash)
	}
	expectedCreatedAt := time.Unix(1_700_000_000, 0).UTC()
	if !repository.createdAt.Equal(expectedCreatedAt) {
		t.Fatalf("created_at = %v, want %v", repository.createdAt, expectedCreatedAt)
	}
	if !session.ExpiresAt.Equal(expectedCreatedAt.Add(30*24*time.Hour)) || !repository.expiresAt.Equal(session.ExpiresAt) {
		t.Fatalf("expires_at = %v, stored = %v", session.ExpiresAt, repository.expiresAt)
	}
}

func TestSignUpRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name, email, displayName, password string
	}{
		{name: "invalid email", email: "not-an-email", displayName: "Maria", password: "password"},
		{name: "empty name", email: "maria@example.com", displayName: "   ", password: "password"},
		{name: "short password", email: "maria@example.com", displayName: "Maria", password: "1234567"},
		{name: "long password", email: "maria@example.com", displayName: "Maria", password: string(bytes.Repeat([]byte{'a'}, 73))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := newTestService(&repositoryStub{}, time.Unix(1_700_000_000, 0))
			if _, err := service.SignUp(context.Background(), test.email, test.displayName, test.password); !errors.Is(err, ErrInvalidRegistration) {
				t.Fatalf("error = %v, want ErrInvalidRegistration", err)
			}
		})
	}
}

func TestSignInCreatesAnAdditionalSessionForValidCredentials(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password-123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_000, 0).UTC()
	repository := &repositoryStub{
		found: true,
		credentials: domain.UserCredentials{
			User:         domain.User{ID: 9, Email: "maria@example.com", Name: "Maria"},
			PasswordHash: string(hash),
		},
	}
	service := newTestService(repository, now)

	session, err := service.SignIn(context.Background(), " MARIA@example.com ", "password-123")
	if err != nil {
		t.Fatal(err)
	}
	if session.User.ID != 9 || repository.tokenUserID != 9 || repository.tokenHash == session.AccessToken {
		t.Fatalf("unexpected session or token persistence: %#v, repository = %#v", session, repository)
	}
}

func TestSignInUsesSameErrorForUnknownEmailAndWrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		repository *repositoryStub
	}{
		{name: "unknown email", repository: &repositoryStub{}},
		{name: "wrong password", repository: &repositoryStub{found: true, credentials: domain.UserCredentials{PasswordHash: string(hash)}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := newTestService(test.repository, time.Unix(1_700_000_000, 0))
			if _, err := service.SignIn(context.Background(), "maria@example.com", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("error = %v, want ErrInvalidCredentials", err)
			}
		})
	}
}

func TestAuthenticateFindsUserFromOpaqueToken(t *testing.T) {
	repository := &repositoryStub{found: true, tokenUser: domain.User{ID: 11, Email: "user@example.com", Name: "User"}}
	service := newTestService(repository, time.Unix(1_700_000_000, 0))

	user, err := service.Authenticate(context.Background(), "opaque-token")
	if err != nil || user.ID != 11 || repository.tokenHash == "opaque-token" || len(repository.tokenHash) != 64 {
		t.Fatalf("user = %#v, hash = %q, err = %v", user, repository.tokenHash, err)
	}
	repository.found = false
	if _, err := service.Authenticate(context.Background(), "opaque-token"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("error = %v, want ErrInvalidToken", err)
	}
}

func newTestService(repository domain.AuthRepository, now time.Time) *Service {
	return &Service{repository: repository, now: func() time.Time { return now }, random: bytes.NewReader(make([]byte, 128))}
}
