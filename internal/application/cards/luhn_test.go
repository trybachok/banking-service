package cards

import "testing"

func TestGeneratePANProducesValidLuhnNumber(t *testing.T) {
	pan, err := GeneratePAN()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pan) != 16 {
		t.Fatalf("expected 16 digits, got %d", len(pan))
	}

	if !ValidateLuhn(pan) {
		t.Fatalf("expected valid Luhn PAN, got %s", pan)
	}
}

func TestValidateLuhnRejectsInvalidNumber(t *testing.T) {
	if ValidateLuhn("1234567890123456") {
		t.Fatal("expected invalid number")
	}
}

func TestMaskPAN(t *testing.T) {
	result := MaskPAN("2200123412345678")

	if result != "2200 **** **** 5678" {
		t.Fatalf("unexpected mask: %s", result)
	}
}
