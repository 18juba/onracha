package auth

import (
	"context"
	"time"
)

// Repository exposes the persistence operations required by authentication use cases.
type Repository interface {
	CreateUser(ctx context.Context, user User, passwordHash string) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, string, error)
	CreateSession(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error
	FindUserBySession(ctx context.Context, tokenHash []byte, now time.Time) (User, error)
	DeleteSession(ctx context.Context, tokenHash []byte) error
}
