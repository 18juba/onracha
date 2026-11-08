package auth

import (
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	// MinimumPasswordLength is the minimum number of Unicode characters in a password.
	MinimumPasswordLength = 15
	// MaximumPasswordLength limits password input while allowing long passphrases.
	MaximumPasswordLength = 128
)

var (
	// ErrInvalidName indicates that a user name is empty after trimming whitespace.
	ErrInvalidName = errors.New("user name is required")
	// ErrInvalidEmail indicates that a user email is not a valid address.
	ErrInvalidEmail = errors.New("user email is invalid")
	// ErrPasswordTooShort indicates that a password does not meet the minimum length.
	ErrPasswordTooShort = errors.New("password is too short")
	// ErrPasswordTooLong indicates that a password exceeds the maximum length.
	ErrPasswordTooLong = errors.New("password is too long")
	// ErrEmailAlreadyRegistered indicates that an account already uses the email.
	ErrEmailAlreadyRegistered = errors.New("email is already registered")
	// ErrUserNotFound indicates that no user exists for the requested email.
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidCredentials is returned for an unknown email or incorrect password.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrInvalidSession indicates that a session token is invalid or expired.
	ErrInvalidSession = errors.New("invalid or expired session")
)

// User contains the public identity fields for an ONRACHA account.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// NewUser trims the name and normalizes the email before validating them.
func NewUser(name, email string) (User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return User{}, ErrInvalidName
	}

	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return User{}, ErrInvalidEmail
	}

	return User{Name: name, Email: normalizedEmail}, nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || len(email) > 254 {
		return "", ErrInvalidEmail
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", ErrInvalidEmail
	}

	return email, nil
}

// ValidatePassword checks the password length without trimming or restricting characters.
func ValidatePassword(password string) error {
	if !utf8.ValidString(password) {
		return ErrPasswordTooLong
	}

	length := utf8.RuneCountInString(password)
	if length < MinimumPasswordLength {
		return ErrPasswordTooShort
	}
	if length > MaximumPasswordLength {
		return ErrPasswordTooLong
	}

	return nil
}
