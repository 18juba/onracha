package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestNewUserTrimsNameAndNormalizesEmail(t *testing.T) {
	user, err := NewUser("  Lucas Oliveira  ", " LUCAS@Example.COM ")
	if err != nil {
		t.Fatalf("NewUser() error = %v, want nil", err)
	}

	if user.Name != "Lucas Oliveira" {
		t.Errorf("Name = %q, want %q", user.Name, "Lucas Oliveira")
	}
	if user.Email != "lucas@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "lucas@example.com")
	}
}

func TestNewUserRejectsBlankName(t *testing.T) {
	_, err := NewUser(" \t ", "lucas@example.com")
	if !errors.Is(err, ErrInvalidName) {
		t.Fatalf("NewUser() error = %v, want %v", err, ErrInvalidName)
	}
}

func TestNewUserRejectsInvalidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{name: "empty", email: ""},
		{name: "missing domain", email: "lucas"},
		{name: "display name", email: "Lucas <lucas@example.com>"},
		{name: "multiple at signs", email: "lucas@@example.com"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewUser("Lucas Oliveira", test.email)
			if !errors.Is(err, ErrInvalidEmail) {
				t.Fatalf("NewUser() error = %v, want %v", err, ErrInvalidEmail)
			}
		})
	}
}

func TestValidatePasswordRequiresAtLeastMinimumLength(t *testing.T) {
	password := strings.Repeat("a", MinimumPasswordLength-1)
	if err := ValidatePassword(password); !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("ValidatePassword() error = %v, want %v", err, ErrPasswordTooShort)
	}
}

func TestValidatePasswordAcceptsEightCharacters(t *testing.T) {
	if err := ValidatePassword("12345678"); err != nil {
		t.Fatalf("ValidatePassword() error = %v, want nil for an 8-character password", err)
	}
}
func TestValidatePasswordRejectsPasswordAboveMaximumLength(t *testing.T) {
	password := strings.Repeat("a", MaximumPasswordLength+1)
	if err := ValidatePassword(password); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("ValidatePassword() error = %v, want %v", err, ErrPasswordTooLong)
	}
}

func TestValidatePasswordAllowsUnicodePassphraseAndWhitespace(t *testing.T) {
	if err := ValidatePassword("bola society é para todos"); err != nil {
		t.Fatalf("ValidatePassword() error = %v, want nil", err)
	}
}
