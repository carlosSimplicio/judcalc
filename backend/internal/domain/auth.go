package domain

import (
	"context"
	"errors"
	"time"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type NewUser struct {
	Email        string
	Name         string
	PasswordHash string
}

type UserCredentials struct {
	User
	PasswordHash string
}

type AuthRepository interface {
	CreateUserWithToken(context.Context, NewUser, string, time.Time, time.Time) (User, error)
	FindUserByEmail(context.Context, string) (UserCredentials, bool, error)
	CreateToken(context.Context, int64, string, time.Time, time.Time) error
	FindUserByTokenHash(context.Context, string, time.Time) (User, bool, error)
}
