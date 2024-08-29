package main

import "math"

func calculateResult(price int, downPayment float64, totalPeriod int, fixedInterest []float64, fixedPeriod []int, floatInterest float64, floatPeriod int) {
	principal := float64(price) * (1 - downPayment/100)

	// calculate tiered fix
	for i := 0; i < len(fixedPeriod); i++ {
		// update pokok with remaining
		principal = calculate(principal, fixedInterest[i], fixedPeriod[i], totalPeriod)

		// update values for next tiered fix
		totalPeriod = totalPeriod - fixedPeriod[i]
	}

	// calculate remaining as floating
	calculate(principal, floatInterest, floatPeriod, totalPeriod)
}

func calculate(principal, interestRate float64, interestPeriod, totalPeriod int) float64 {
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

		println("[%d] cicilan: %v, bunga: %v, pokok: %v\n",
			i+1, ac.FormatMoney(monthlyInstallment), ac.FormatMoney(interestInstallment), ac.FormatMoney(principalInstallment))

		principal -= principalInstallment
	}

	return principal // return any remainder
}
