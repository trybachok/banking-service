package credits

import (
	"math"
	"strconv"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/pkg/money"
)

func CalculateAnnuityPayment(principal money.Money, annualRatePercent string, termMonths int) (money.Money, error) {
	if !principal.IsPositive() {
		return money.Money{}, domain.NewDomainError(domain.ErrorCodeValidation, "principal amount must be positive", domain.ErrValidation)
	}

	if termMonths <= 0 {
		return money.Money{}, domain.NewDomainError(domain.ErrorCodeValidation, "term months must be greater than zero", domain.ErrValidation)
	}

	rate, err := strconv.ParseFloat(annualRatePercent, 64)
	if err != nil {
		return money.Money{}, domain.NewDomainError(domain.ErrorCodeValidation, "annual rate must be numeric", err)
	}

	monthlyRate := rate / 100 / 12
	principalMinor := float64(principal.MinorUnits())

	if monthlyRate == 0 {
		return money.FromMinorUnits(int64(math.Ceil(principalMinor / float64(termMonths))))
	}

	pow := math.Pow(1+monthlyRate, float64(termMonths))
	paymentMinor := principalMinor * monthlyRate * pow / (pow - 1)

	return money.FromMinorUnits(int64(math.Ceil(paymentMinor)))
}
