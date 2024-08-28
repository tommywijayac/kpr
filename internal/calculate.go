package internal

import (
	"fmt"
	"math"

	"github.com/leekchan/accounting"
)

type Usecase struct {
	ac accounting.Accounting
}

func New() *Usecase {
	return &Usecase{
		ac: accounting.Accounting{Symbol: "IDR ", Precision: 2},
	}
}

func (m *Usecase) Calculate(principal, interestRate float64, interestPeriod, totalPeriod int) float64 {
	var (
		monthlyInterestRate float64 = float64(interestRate) / 100 / 12

		// still don't fully understand, but this makes principal is paid in certain percentage to interest & installment
		loan        float64 = 1 + monthlyInterestRate
		totalLoan   float64 = math.Pow(loan, float64(totalPeriod))
		installment float64 = 1 - 1/totalLoan

		interestInstallment float64 = principal * monthlyInterestRate   // angsuran bunga
		monthlyInstallment  float64 = interestInstallment / installment // by dividing with (1 - 1/totalLoan), installment will contain (interestInstallment + principalInstallment)
	)

	for i := 0; i < interestPeriod; i++ {
		interestInstallment = principal * monthlyInterestRate            // angsuran bunga
		principalInstallment := monthlyInstallment - interestInstallment // angsuran pokok

		fmt.Printf("[%d] cicilan: %v, bunga: %v, pokok: %v\n",
			i+1, m.ac.FormatMoney(monthlyInstallment), m.ac.FormatMoney(interestInstallment), m.ac.FormatMoney(principalInstallment))

		principal -= principalInstallment
	}

	return principal // return any remainder
}
