package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"onracha/internal/auth"
)

// AuthRepository stores authentication data in PostgreSQL.
type AuthRepository struct {
	db *sql.DB
}

// NewAuthRepository creates a PostgreSQL authentication repository backed by an existing connection pool.
func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// CreateUser persists a public user profile and its password hash.
func (r *AuthRepository) CreateUser(ctx context.Context, user auth.User, passwordHash string) (auth.User, error) {
	const query = `INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id, name, email, created_at`
	err := r.db.QueryRowContext(ctx, query, user.Name, user.Email, passwordHash).
		Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return auth.User{}, auth.ErrEmailAlreadyRegistered
		}
		return auth.User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

// FindUserByEmail returns a public profile and its password hash for credential verification.
func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (auth.User, string, error) {
	const query = `SELECT id, name, email, created_at, password_hash FROM users WHERE email = $1`
	var user auth.User
	var passwordHash string
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, "", auth.ErrUserNotFound
	}
	if err != nil {
		return auth.User{}, "", fmt.Errorf("find user by email: %w", err)
	}
	return user, passwordHash, nil
}

// CreateSession stores only the hash of an opaque session token.
func (r *AuthRepository) CreateSession(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	const query = `INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`
	if _, err := r.db.ExecContext(ctx, query, tokenHash, userID, expiresAt); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// FindUserBySession resolves an unexpired session hash to its account.
func (r *AuthRepository) FindUserBySession(ctx context.Context, tokenHash []byte, now time.Time) (auth.User, error) {
	const query = `
		SELECT users.id, users.name, users.email, users.created_at
		FROM sessions
		JOIN users ON users.id = sessions.user_id
		WHERE sessions.token_hash = $1 AND sessions.expires_at > $2`
	var user auth.User
	err := r.db.QueryRowContext(ctx, query, tokenHash, now).
		Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, auth.ErrInvalidSession
	}
	if err != nil {
		return auth.User{}, fmt.Errorf("find user by session: %w", err)
	}
	return user, nil
}
