package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Time       uint32 = 2
	argon2Memory     uint32 = 19 * 1024
	argon2Threads    uint8  = 1
	argon2KeyLength  uint32 = 32
	argon2SaltLength        = 16
)

// Argon2idHasher hashes passwords with Argon2id using the project's fixed parameters.
type Argon2idHasher struct{}

// NewArgon2idHasher creates the password hasher used by the authentication service.
func NewArgon2idHasher() *Argon2idHasher {
	return &Argon2idHasher{}
}

// Hash returns a salted Argon2id hash in a self-describing encoded format.
func (h *Argon2idHasher) Hash(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	derivedKey := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLength)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory,
		argon2Time,
		argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(derivedKey),
	), nil
}

// Verify compares a password with an encoded Argon2id hash in constant time.
func (h *Argon2idHasher) Verify(password, encodedHash string) bool {
	if ValidatePassword(password) != nil || len(encodedHash) > 256 {
		return false
	}

	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=19456,t=2,p=1" {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != argon2SaltLength {
		return false
	}

	expectedKey, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expectedKey) != int(argon2KeyLength) {
		return false
	}

	actualKey := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLength)
	return subtle.ConstantTimeCompare(actualKey, expectedKey) == 1
}
