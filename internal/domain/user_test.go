package domain

import "testing"

func TestValidateEmail(t *testing.T) {
	if err := ValidateEmail("user@example.com"); err != nil {
		t.Fatalf("expected valid email, got error: %v", err)
	}
}

func TestValidateEmailRejectsInvalidValue(t *testing.T) {
	if err := ValidateEmail("invalid-email"); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateUsername(t *testing.T) {
	if err := ValidateUsername("valid_user-1"); err != nil {
		t.Fatalf("expected valid username, got error: %v", err)
	}
}

func TestValidateUsernameRejectsInvalidValue(t *testing.T) {
	if err := ValidateUsername("!"); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("strong-password"); err != nil {
		t.Fatalf("expected valid password, got error: %v", err)
	}
}

func TestValidatePasswordRejectsShortValue(t *testing.T) {
	if err := ValidatePassword("short"); err == nil {
		t.Fatal("expected validation error")
	}
}
