package main

import (
	"fmt"
	"math"

	"github.com/leekchan/accounting"
)

var ac accounting.Accounting

func main() {
	ac = accounting.Accounting{Symbol: "IDR ", Precision: 2}

	// hardcoded inputs
	pokok := float64(1000000000)
	bungafix := []float64{7}
	periodefix := []int{120} // len bungafix == len periodefix
	periodetotal := 132      // can be bigger than total fix, means remaining is floating. but can't be less than total fix

	bungafloat := 11
	periodefloat := periodetotal
	for i := 0; i < len(periodefix); i++ {
		periodefloat -= periodefix[i]
	}
	if periodefloat < 0 {
		panic("invalid period")
	}

	// calculate tiered fix
	for i := 0; i < len(periodefix); i++ {
		pokok = calculate(pokok, bungafix[i], periodefix[i], periodetotal) // update pokok with remaining

		// update values for next tiered fix
		periodetotal = periodetotal - periodefix[i]
	}

	// calculate remaining as floating
	calculate(pokok, float64(bungafloat), periodefloat, periodetotal)
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

		fmt.Printf("[%d] cicilan: %v, bunga: %v, pokok: %v\n",
			i+1, ac.FormatMoney(monthlyInstallment), ac.FormatMoney(interestInstallment), ac.FormatMoney(principalInstallment))

		principal -= principalInstallment
	}

	return principal // return any remainder
}
