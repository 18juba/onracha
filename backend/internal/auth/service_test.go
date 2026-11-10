package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"
)

func TestRegisterCreatesUserWithHashedPassword(t *testing.T) {
	repository := &fakeRepository{
		createdUser: User{ID: 42, Name: "Lucas Oliveira", Email: "lucas@example.com"},
	}
	hasher := &fakePasswordHasher{hash: "encoded-password-hash"}
	service := NewService(repository, hasher)

	user, err := service.Register(context.Background(), " Lucas Oliveira ", "LUCAS@example.com", "uma senha muito segura 2026")
	if err != nil {
		t.Fatalf("Register() error = %v, want nil", err)
	}
	if user.ID != 42 || user.Email != "lucas@example.com" {
		t.Errorf("Register() user = %+v, want persisted normalized user", user)
	}
	if repository.savedPasswordHash != "encoded-password-hash" {
		t.Errorf("saved password hash = %q, want encoded hash", repository.savedPasswordHash)
	}
	if repository.savedUser.Email != "lucas@example.com" {
		t.Errorf("saved email = %q, want normalized email", repository.savedUser.Email)
	}
	if hasher.hashCalls != 1 {
		t.Errorf("Hash() calls = %d, want 1", hasher.hashCalls)
	}
}

func TestRegisterRejectsInvalidInputBeforeRepository(t *testing.T) {
	repository := &fakeRepository{}
	hasher := &fakePasswordHasher{}
	service := NewService(repository, hasher)

	_, err := service.Register(context.Background(), " ", "lucas@example.com", "uma senha muito segura 2026")
	if !errors.Is(err, ErrInvalidName) {
		t.Fatalf("Register() error = %v, want %v", err, ErrInvalidName)
	}
	if repository.createCalls != 0 || hasher.hashCalls != 0 {
		t.Fatal("invalid input must not reach the repository or password hasher")
	}
}

func TestRegisterRejectsShortPasswordBeforeHashing(t *testing.T) {
	repository := &fakeRepository{}
	hasher := &fakePasswordHasher{}
	service := NewService(repository, hasher)

	_, err := service.Register(context.Background(), "Lucas", "lucas@example.com", "curta")
	if !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("Register() error = %v, want %v", err, ErrPasswordTooShort)
	}
	if repository.createCalls != 0 || hasher.hashCalls != 0 {
		t.Fatal("invalid password must not reach the repository or password hasher")
	}
}

func TestRegisterReturnsDuplicateEmailError(t *testing.T) {
	repository := &fakeRepository{createErr: ErrEmailAlreadyRegistered}
	service := NewService(repository, &fakePasswordHasher{hash: "hash"})

	_, err := service.Register(context.Background(), "Lucas", "lucas@example.com", "uma senha muito segura 2026")
	if !errors.Is(err, ErrEmailAlreadyRegistered) {
		t.Fatalf("Register() error = %v, want %v", err, ErrEmailAlreadyRegistered)
	}
}

func TestLoginCreatesHashedSessionForValidCredentials(t *testing.T) {
	user := User{ID: 42, Name: "Lucas Oliveira", Email: "lucas@example.com"}
	repository := &fakeRepository{foundUser: user, foundPasswordHash: "stored-hash"}
	hasher := &fakePasswordHasher{verifyOK: true}
	service := NewService(repository, hasher)

	loggedInUser, session, err := service.Login(context.Background(), " LUCAS@example.com ", "uma senha muito segura 2026")
	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}
	if loggedInUser.ID != user.ID {
		t.Errorf("Login() user ID = %d, want %d", loggedInUser.ID, user.ID)
	}
	if session.Token == "" {
		t.Fatal("Login() returned an empty session token")
	}
	if !session.ExpiresAt.After(time.Now()) {
		t.Errorf("session expiry %v should be in the future", session.ExpiresAt)
	}
	if len(repository.sessions) != 1 {
		t.Fatalf("created sessions = %d, want 1", len(repository.sessions))
	}
	expectedHash := sha256.Sum256([]byte(session.Token))
	if string(repository.sessions[0].tokenHash) != string(expectedHash[:]) {
		t.Error("repository should receive only the hash of the session token")
	}
}

func TestLoginUsesGenericErrorForUnknownEmail(t *testing.T) {
	repository := &fakeRepository{findErr: ErrUserNotFound}
	hasher := &fakePasswordHasher{hash: "dummy-hash"}
	service := NewService(repository, hasher)

	_, _, err := service.Login(context.Background(), "missing@example.com", "uma senha muito segura 2026")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
	if hasher.hashCalls != 1 {
		t.Errorf("Hash() calls = %d, want 1 to keep unknown-user checks expensive", hasher.hashCalls)
	}
}

func TestLoginUsesGenericErrorForWrongPassword(t *testing.T) {
	repository := &fakeRepository{
		foundUser:         User{ID: 42, Email: "lucas@example.com"},
		foundPasswordHash: "stored-hash",
	}
	service := NewService(repository, &fakePasswordHasher{verifyOK: false})

	_, _, err := service.Login(context.Background(), "lucas@example.com", "senha incorreta e comprida")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
	if len(repository.sessions) != 0 {
		t.Errorf("created sessions = %d, want 0", len(repository.sessions))
	}
}

func TestAuthenticateLoadsUserFromActiveSession(t *testing.T) {
	user := User{ID: 42, Email: "lucas@example.com"}
	token := base64.RawURLEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	tokenHash := sha256.Sum256([]byte(token))
	repository := &fakeRepository{sessionUser: user, sessionTokenHash: tokenHash[:]}
	service := NewService(repository, &fakePasswordHasher{})

	got, err := service.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("Authenticate() error = %v, want nil", err)
	}
	if got.ID != user.ID {
		t.Errorf("Authenticate() user ID = %d, want %d", got.ID, user.ID)
	}
}

func TestAuthenticateReturnsInvalidSessionWhenNotFound(t *testing.T) {
	repository := &fakeRepository{sessionErr: ErrInvalidSession}
	service := NewService(repository, &fakePasswordHasher{})

	token := base64.RawURLEncoding.EncodeToString([]byte("fedcba9876543210fedcba9876543210"))
	_, err := service.Authenticate(context.Background(), token)
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("Authenticate() error = %v, want %v", err, ErrInvalidSession)
	}
}

type fakeRepository struct {
	createdUser       User
	createErr         error
	createCalls       int
	savedUser         User
	savedPasswordHash string
	foundUser         User
	foundPasswordHash string
	findErr           error
	sessions          []fakeSession
	sessionUser       User
	sessionTokenHash  []byte
	sessionErr        error
}

type fakeSession struct {
	userID    int64
	tokenHash []byte
	expiresAt time.Time
}

func (r *fakeRepository) CreateUser(_ context.Context, user User, passwordHash string) (User, error) {
	r.createCalls++
	r.savedUser = user
	r.savedPasswordHash = passwordHash
	if r.createErr != nil {
		return User{}, r.createErr
	}
	if r.createdUser.ID != 0 {
		return r.createdUser, nil
	}
	user.ID = 1
	return user, nil
}

func (r *fakeRepository) FindUserByEmail(_ context.Context, _ string) (User, string, error) {
	if r.findErr != nil {
		return User{}, "", r.findErr
	}
	return r.foundUser, r.foundPasswordHash, nil
}

func (r *fakeRepository) CreateSession(_ context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	r.sessions = append(r.sessions, fakeSession{
		userID:    userID,
		tokenHash: append([]byte(nil), tokenHash...),
		expiresAt: expiresAt,
	})
	return nil
}

func (r *fakeRepository) FindUserBySession(_ context.Context, tokenHash []byte, _ time.Time) (User, error) {
	if r.sessionErr != nil {
		return User{}, r.sessionErr
	}
	if string(tokenHash) != string(r.sessionTokenHash) {
		return User{}, ErrInvalidSession
	}
	return r.sessionUser, nil
}

func (r *fakeRepository) DeleteSession(_ context.Context, tokenHash []byte) error {
	activeSessions := r.sessions[:0]
	for _, session := range r.sessions {
		if string(session.tokenHash) != string(tokenHash) {
			activeSessions = append(activeSessions, session)
		}
	}
	r.sessions = activeSessions
	return nil
}

type fakePasswordHasher struct {
	hash      string
	hashCalls int
	verifyOK  bool
}

func (h *fakePasswordHasher) Hash(_ string) (string, error) {
	h.hashCalls++
	return h.hash, nil
}

func (h *fakePasswordHasher) Verify(_, _ string) bool {
	return h.verifyOK
}
