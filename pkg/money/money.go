package money

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidMoneyAmount = errors.New("invalid money amount")
	ErrNegativeMoney      = errors.New("money amount cannot be negative")
)

type Money struct {
	minorUnits int64
}

func Zero() Money {
	return Money{minorUnits: 0}
}

func FromMinorUnits(minorUnits int64) (Money, error) {
	if minorUnits < 0 {
		return Money{}, ErrNegativeMoney
	}

	return Money{minorUnits: minorUnits}, nil
}

func FromRubString(value string) (Money, error) {
	normalized := strings.TrimSpace(value)

	if normalized == "" {
		return Money{}, ErrInvalidMoneyAmount
	}

	if strings.HasPrefix(normalized, "-") {
		return Money{}, ErrNegativeMoney
	}

	parts := strings.Split(normalized, ".")

	if len(parts) > 2 {
		return Money{}, ErrInvalidMoneyAmount
	}

	rubles, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Money{}, ErrInvalidMoneyAmount
	}

	kopecks := int64(0)

	if len(parts) == 2 {
		fraction := parts[1]

		if len(fraction) > 2 {
			return Money{}, ErrInvalidMoneyAmount
		}

		if len(fraction) == 1 {
			fraction += "0"
		}

		if len(fraction) == 0 {
			fraction = "00"
		}

		kopecks, err = strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return Money{}, ErrInvalidMoneyAmount
		}
	}

	return FromMinorUnits(rubles*100 + kopecks)
}

func (m Money) MinorUnits() int64 {
	return m.minorUnits
}

func (m Money) IsZero() bool {
	return m.minorUnits == 0
}

func (m Money) IsPositive() bool {
	return m.minorUnits > 0
}

func (m Money) Add(other Money) Money {
	return Money{minorUnits: m.minorUnits + other.minorUnits}
}

func (m Money) Sub(other Money) (Money, error) {
	result := m.minorUnits - other.minorUnits

	if result < 0 {
		return Money{}, ErrNegativeMoney
	}

	return Money{minorUnits: result}, nil
}

func (m Money) String() string {
	return fmt.Sprintf("%d.%02d", m.minorUnits/100, m.minorUnits%100)
}
