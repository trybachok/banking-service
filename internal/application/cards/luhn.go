package cards

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
)

func GeneratePAN() (string, error) {
	body := "2200"

	for len(body) < 15 {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generate PAN digit: %w", err)
		}

		body += digit.String()
	}

	checkDigit := CalculateLuhnCheckDigit(body)

	return body + strconv.Itoa(checkDigit), nil
}

func CalculateLuhnCheckDigit(body string) int {
	sum := 0
	double := true

	for i := len(body) - 1; i >= 0; i-- {
		digit := int(body[i] - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return (10 - (sum % 10)) % 10
}

func ValidateLuhn(number string) bool {
	if len(number) == 0 {
		return false
	}

	sum := 0
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		if number[i] < '0' || number[i] > '9' {
			return false
		}

		digit := int(number[i] - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}

func MaskPAN(pan string) string {
	if len(pan) < 4 {
		return "****"
	}

	return pan[:4] + " **** **** " + pan[len(pan)-4:]
}

func Last4(pan string) string {
	if len(pan) < 4 {
		return pan
	}

	return pan[len(pan)-4:]
}
