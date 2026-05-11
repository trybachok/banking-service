package money

import "testing"

func TestFromRubString(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected int64
	}{
		{name: "rubles only", input: "100", expected: 10000},
		{name: "rubles and kopecks", input: "100.25", expected: 10025},
		{name: "one fractional digit", input: "100.5", expected: 10050},
		{name: "zero", input: "0.00", expected: 0},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			value, err := FromRubString(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if value.MinorUnits() != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, value.MinorUnits())
			}
		})
	}
}

func TestFromRubStringRejectsNegativeAmount(t *testing.T) {
	_, err := FromRubString("-1.00")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMoneySubRejectsNegativeResult(t *testing.T) {
	left, err := FromRubString("10.00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	right, err := FromRubString("11.00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = left.Sub(right)
	if err == nil {
		t.Fatal("expected error")
	}
}
