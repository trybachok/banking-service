package crypto

import "testing"

func TestPasswordHasherHashAndCompare(t *testing.T) {
	hasher := NewPasswordHasher()

	hash, err := hasher.Hash("strong-password")
	if err != nil {
		t.Fatalf("unexpected hash error: %v", err)
	}

	if hash == "strong-password" {
		t.Fatal("password hash must not equal plain password")
	}

	if err := hasher.Compare("strong-password", hash); err != nil {
		t.Fatalf("expected password to match: %v", err)
	}
}

func TestPasswordHasherRejectsWrongPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	hash, err := hasher.Hash("strong-password")
	if err != nil {
		t.Fatalf("unexpected hash error: %v", err)
	}

	if err := hasher.Compare("wrong-password", hash); err == nil {
		t.Fatal("expected wrong password to be rejected")
	}
}
