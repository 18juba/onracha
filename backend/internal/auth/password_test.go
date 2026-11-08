package auth

import (
	"strings"
	"testing"
)

func TestArgon2idHasherHashesAndVerifiesPassword(t *testing.T) {
	hasher := NewArgon2idHasher()
	password := "uma senha muito segura 2026"

	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v, want nil", err)
	}
	if hash == password || !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("Hash() = %q, want an Argon2id encoded hash", hash)
	}
	if !hasher.Verify(password, hash) {
		t.Error("Verify() = false for the password used to create the hash")
	}
	if hasher.Verify("senha diferente e comprida", hash) {
		t.Error("Verify() = true for a different password")
	}

	secondHash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("second Hash() error = %v, want nil", err)
	}
	if hash == secondHash {
		t.Error("Hash() returned the same value twice; each password hash needs a unique salt")
	}
}

func TestArgon2idHasherRejectsMalformedHash(t *testing.T) {
	if NewArgon2idHasher().Verify("uma senha muito segura 2026", "not-a-hash") {
		t.Error("Verify() = true for a malformed stored hash")
	}
}
