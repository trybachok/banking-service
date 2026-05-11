package jwt

import "testing"

func TestManagerIssueAndParse(t *testing.T) {
	manager := NewManager("test-secret", 24)

	token, expiresAt, err := manager.Issue("user-1", "customer")
	if err != nil {
		t.Fatalf("unexpected issue error: %v", err)
	}

	if token == "" {
		t.Fatal("expected token")
	}

	if expiresAt.IsZero() {
		t.Fatal("expected expires at")
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if claims.UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", claims.UserID)
	}

	if claims.Role != "customer" {
		t.Fatalf("expected customer, got %q", claims.Role)
	}
}

func TestManagerRejectsInvalidToken(t *testing.T) {
	manager := NewManager("test-secret", 24)

	if _, err := manager.Parse("invalid-token"); err == nil {
		t.Fatal("expected invalid token error")
	}
}
