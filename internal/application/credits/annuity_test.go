package credits

import (
	"testing"

	"github.com/example/banking-service/pkg/money"
)

func TestCalculateAnnuityPayment(t *testing.T) {
	principal, err := money.FromRubString("100000.00")
	if err != nil {
		t.Fatalf("unexpected money error: %v", err)
	}

	payment, err := CalculateAnnuityPayment(principal, "20.0000", 12)
	if err != nil {
		t.Fatalf("unexpected annuity error: %v", err)
	}

	if !payment.IsPositive() {
		t.Fatal("expected positive annuity payment")
	}
}

func TestCalculateAnnuityPaymentRejectsZeroTerm(t *testing.T) {
	principal, err := money.FromRubString("100000.00")
	if err != nil {
		t.Fatalf("unexpected money error: %v", err)
	}

	if _, err := CalculateAnnuityPayment(principal, "20.0000", 0); err == nil {
		t.Fatal("expected error")
	}
}
