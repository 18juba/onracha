package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

const sessionLifetime = 24 * time.Hour

// Repository exposes the persistence operations required by authentication use cases.
type Repository interface {
	CreateUser(ctx context.Context, user User, passwordHash string) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, string, error)
	CreateSession(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error
	FindUserBySession(ctx context.Context, tokenHash []byte, now time.Time) (User, error)
}

// PasswordHasher hashes passwords and verifies them against stored hashes.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) bool
}

// Session contains the opaque token returned after successful login.
type Session struct {
	Token     string
	ExpiresAt time.Time
}

// Service implements account registration and session-based authentication.
type Service struct {
	repository     Repository
	passwordHasher PasswordHasher
}

// NewService creates an authentication service with its persistence and hashing dependencies.
func NewService(repository Repository, passwordHasher PasswordHasher) *Service {
	return &Service{repository: repository, passwordHasher: passwordHasher}
}

// Register validates and persists a user while storing only a password hash.
func (s *Service) Register(ctx context.Context, name, email, password string) (User, error) {
	user, err := NewUser(name, email)
	if err != nil {
		return User{}, err
	}
	if err := ValidatePassword(password); err != nil {
		return User{}, err
	}

	passwordHash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	createdUser, err := s.repository.CreateUser(ctx, user, passwordHash)
	if errors.Is(err, ErrEmailAlreadyRegistered) {
		return User{}, ErrEmailAlreadyRegistered
	}
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return createdUser, nil
}

// Login verifies credentials and persists a hash of a newly created session token.
func (s *Service) Login(ctx context.Context, email, password string) (User, Session, error) {
	if err := ValidatePassword(password); err != nil {
		return User{}, Session{}, ErrInvalidCredentials
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		s.hashForUnknownUser(password)
		return User{}, Session{}, ErrInvalidCredentials
	}

	user, passwordHash, err := s.repository.FindUserByEmail(ctx, normalizedEmail)
	if errors.Is(err, ErrUserNotFound) {
		s.hashForUnknownUser(password)
		return User{}, Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return User{}, Session{}, fmt.Errorf("find user by email: %w", err)
	}
	if !s.passwordHasher.Verify(password, passwordHash) {
		return User{}, Session{}, ErrInvalidCredentials
	}

	token, err := newSessionToken()
	if err != nil {
		return User{}, Session{}, fmt.Errorf("create session token: %w", err)
	}
	tokenHash := sha256.Sum256([]byte(token))
	expiresAt := time.Now().UTC().Add(sessionLifetime)
	if err := s.repository.CreateSession(ctx, user.ID, tokenHash[:], expiresAt); err != nil {
		return User{}, Session{}, fmt.Errorf("save session: %w", err)
	}

	return user, Session{Token: token, ExpiresAt: expiresAt}, nil
}

// Authenticate resolves a session token without storing the raw token in PostgreSQL.
func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	decodedToken, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(decodedToken) != 32 {
		return User{}, ErrInvalidSession
	}

	tokenHash := sha256.Sum256([]byte(token))
	user, err := s.repository.FindUserBySession(ctx, tokenHash[:], time.Now().UTC())
	if errors.Is(err, ErrInvalidSession) || errors.Is(err, ErrUserNotFound) {
		return User{}, ErrInvalidSession
	}
	if err != nil {
		return User{}, fmt.Errorf("find user by session: %w", err)
	}

	return user, nil
}

func (s *Service) hashForUnknownUser(password string) {
	_, _ = s.passwordHasher.Hash(password)
}

func newSessionToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}
