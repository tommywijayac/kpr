package main

import (
	"fmt"
	"math"

	"github.com/leekchan/accounting"
)

type Result struct {
	Interests []float64
	Periods   []int

	// breakdown
	Installment          []float64
	InterestInstallment  []float64
	PrincipalInstallment []float64

	// per period
	PeriodMonthlyInstallment      []float64
	PeriodSumInstallment          []float64
	PeriodSumInterestInstallment  []float64
	PeriodSumPrincipalInstallment []float64

	// summary
	TotalInstallment     float64
	TotalInterests       float64
	TotalPrincipal       float64
	PrincipalBeforeFloat float64
}

type FmtResult struct {
	Interests []string
	Periods   []string

	Installment          []string
	InterestInstallment  []string
	PrincipalInstallment []string

	PeriodMonthlyInstallment      []string
	PeriodSumInstallment          []string
	PeriodSumInterestInstallment  []string
	PeriodSumPrincipalInstallment []string

	TotalInstallment     string
	TotalInterests       string
	TotalPrincipal       string
	PrincipalBeforeFloat string
}

func calculateResult(price int, downPayment float64, totalPeriod int, fixedInterest []float64, fixedPeriod []int, floatInterest float64, floatPeriod int) Result {
	var finalResult Result

	principal := float64(price) * (1 - downPayment/100)

	// calculate tiered fix
	for i := 0; i < len(fixedPeriod); i++ {
		// update pokok with remaining
		var _result Result
		principal, _result = calculate(principal, fixedInterest[i], fixedPeriod[i], totalPeriod)

		finalResult.Interests = append(finalResult.Interests, fixedInterest[i])
		finalResult.Periods = append(finalResult.Periods, fixedPeriod[i])
		finalResult.add(_result)

		// update values for next tiered fix
		totalPeriod = totalPeriod - fixedPeriod[i]
	}

	if floatPeriod > 0 {
		// calculate remaining as floating
		_, _result := calculate(principal, floatInterest, floatPeriod, totalPeriod)

		finalResult.PrincipalBeforeFloat = principal
		finalResult.Interests = append(finalResult.Interests, floatInterest)
		finalResult.Periods = append(finalResult.Periods, floatPeriod)
		finalResult.add(_result)
	}

	return finalResult
}

func calculate(principal, interestRate float64, interestPeriod, totalPeriod int) (float64, Result) {
	var (
		monthlyInterestRate float64 = float64(interestRate) / 100 / 12

		// still don't fully understand, but this makes principal is paid in certain percentage to interest & Installment
		loan        float64 = 1 + monthlyInterestRate
		totalLoan   float64 = math.Pow(loan, float64(totalPeriod))
		installment float64 = 1 - 1/totalLoan

		interestInstallment float64 = principal * monthlyInterestRate   // angsuran bunga
		monthlyInstallment  float64 = interestInstallment / installment // by dividing with (1 - 1/totalLoan), installment will contain (interestInstallment + principalInstallment)

		result Result
	)

	for i := 0; i < interestPeriod; i++ {
		interestInstallment = principal * monthlyInterestRate            // angsuran bunga
		principalInstallment := monthlyInstallment - interestInstallment // angsuran pokok

		result.Installment = append(result.Installment, monthlyInstallment)
		result.InterestInstallment = append(result.InterestInstallment, interestInstallment)
		result.PrincipalInstallment = append(result.PrincipalInstallment, principalInstallment)

		principal -= principalInstallment
	}

	// return any remainder and breakdown
	return principal, result
}

func (r *Result) add(temp Result) {
	r.PeriodMonthlyInstallment = append(r.PeriodMonthlyInstallment, temp.PeriodMonthlyInstallment...)

	r.Installment = append(r.Installment, temp.Installment...)
	r.InterestInstallment = append(r.InterestInstallment, temp.InterestInstallment...)
	r.PrincipalInstallment = append(r.PrincipalInstallment, temp.PrincipalInstallment...)

	// summarize
	r.PeriodMonthlyInstallment = append(r.PeriodMonthlyInstallment, temp.Installment[0])

	// assume all array are growing at same rate, so index can be re-used
	idx := len(r.PeriodSumInstallment)

	r.PeriodSumInstallment = append(r.PeriodSumInstallment, 0)
	for _, v := range temp.Installment {
		r.PeriodSumInstallment[idx] += v
	}
	r.TotalInstallment += r.PeriodSumInstallment[idx]

	r.PeriodSumInterestInstallment = append(r.PeriodSumInterestInstallment, 0)
	for _, v := range temp.InterestInstallment {
		r.PeriodSumInterestInstallment[idx] += v
	}
	r.TotalInterests += r.PeriodSumInterestInstallment[idx]

	r.PeriodSumPrincipalInstallment = append(r.PeriodSumPrincipalInstallment, 0)
	for _, v := range temp.PrincipalInstallment {
		r.PeriodSumPrincipalInstallment[idx] += v
	}
	r.TotalPrincipal += r.PeriodSumPrincipalInstallment[idx]
}

func (r *Result) format(acfmt accounting.Accounting) FmtResult {
	var (
		result FmtResult
	)

	for _, v := range r.Interests {
		result.Interests = append(result.Interests, accounting.FormatNumberFloat64(v, 2, ",", "."))
	}
	for _, v := range r.Periods {
		result.Periods = append(result.Periods, fmt.Sprintf("%d", v))
	}

	for _, v := range r.PeriodMonthlyInstallment {
		result.PeriodMonthlyInstallment = append(result.PeriodMonthlyInstallment, acfmt.FormatMoneyFloat64(v))
	}
	for _, v := range r.PeriodSumInstallment {
		result.PeriodSumInstallment = append(result.PeriodSumInstallment, acfmt.FormatMoneyFloat64(v))
	}
	for _, v := range r.PeriodSumInterestInstallment {
		result.PeriodSumInterestInstallment = append(result.PeriodSumInterestInstallment, acfmt.FormatMoneyFloat64(v))
	}
	for _, v := range r.PeriodSumPrincipalInstallment {
		result.PeriodSumPrincipalInstallment = append(result.PeriodSumPrincipalInstallment, acfmt.FormatMoneyFloat64(v))
	}

	result.TotalInstallment = acfmt.FormatMoneyFloat64(r.TotalInstallment)
	result.TotalInterests = acfmt.FormatMoneyFloat64(r.TotalInterests)
	result.TotalPrincipal = acfmt.FormatMoneyFloat64(r.TotalPrincipal)

	return result
}
