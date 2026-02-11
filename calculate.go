package main

import (
	"fmt"
	"math"

	"github.com/leekchan/accounting"
)

type MortgageSchema struct {
	Price         float64
	DownPayment   float64
	TotalPeriod   int
	FixedInterest []float64
	FixedPeriod   []int
	FloatInterest float64
	FloatPeriod   int

	EarlyPaymentFee float64
	EarlyPayment    []float64
}

type Result struct {
	Interests []float64
	Periods   []int

	// per month
	Installment          []float64 // monthly installment
	InterestInstallment  []float64 // monthly interest installment
	PrincipalInstallment []float64 // monthly principal installment

	// per year
	YearlyRowNum               []int
	YearlyInstallment          []float64
	YearlyInterestInstallment  []float64
	YearlyPrincipalInstallment []float64

	// per given fixed tier
	// e.g. if defined "fixed berjenjang" of x%-1yr, y%-3yr
	// then each member is total of each element for x% and y% periods.
	PeriodRowNum                  []int
	PeriodMonthlyInstallment      []float64
	PeriodSumInstallment          []float64
	PeriodSumInterestInstallment  []float64
	PeriodSumPrincipalInstallment []float64

	// summary
	Principal            float64
	TotalInstallment     float64
	TotalInterests       float64
	TotalPrincipal       float64
	PrincipalBeforeFloat float64
}

// FmtResult is formatted result for display
// all fields are one to one
type FmtResult struct {
	Interests []string
	Periods   []string

	Installment          []string
	InterestInstallment  []string
	PrincipalInstallment []string

	YearlyRowNum               []int
	YearlyInstallment          []string
	YearlyInterestInstallment  []string
	YearlyPrincipalInstallment []string

	PeriodRowNum                  []int
	PeriodMonthlyInstallment      []string
	PeriodSumInstallment          []string
	PeriodSumInterestInstallment  []string
	PeriodSumPrincipalInstallment []string

	Principal            string
	TotalInstallment     string
	TotalInterests       string
	TotalPrincipal       string
	PrincipalBeforeFloat string
}

func calculateResult(schema MortgageSchema) Result {
	var finalResult Result

	principal := schema.Price * (1 - schema.DownPayment/100)
	finalResult.Principal = principal

	// calculate tiered fix period
	for i := 0; i < len(schema.FixedPeriod); i++ {
		// update pokok with remaining
		var _result Result
		principal, _result = calculate(principal, schema.FixedInterest[i], schema.FixedPeriod[i], schema.TotalPeriod)

		finalResult.Interests = append(finalResult.Interests, schema.FixedInterest[i])
		finalResult.Periods = append(finalResult.Periods, schema.FixedPeriod[i])
		finalResult.add(_result)

		// update values for next tiered fix
		schema.TotalPeriod = schema.TotalPeriod - schema.FixedPeriod[i]
	}

	// calculate remaining period as floating
	if schema.FloatPeriod > 0 {
		_, _result := calculate(principal, schema.FloatInterest, schema.FloatPeriod, schema.TotalPeriod)

		finalResult.PrincipalBeforeFloat = principal
		finalResult.Interests = append(finalResult.Interests, schema.FloatInterest)
		finalResult.Periods = append(finalResult.Periods, schema.FloatPeriod)
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

	j := len(r.YearlyRowNum)
	for i := range temp.Installment {
		if i%12 == 0 {
			j += 1
			r.YearlyRowNum = append(r.YearlyRowNum, j)
			r.YearlyInstallment = append(r.YearlyInstallment, temp.Installment[i])
			r.YearlyInterestInstallment = append(r.YearlyInterestInstallment, temp.InterestInstallment[i])
			r.YearlyPrincipalInstallment = append(r.YearlyPrincipalInstallment, temp.PrincipalInstallment[i])
		} else {
			r.YearlyInstallment[j-1] += temp.Installment[i]
			r.YearlyInterestInstallment[j-1] += temp.InterestInstallment[i]
			r.YearlyPrincipalInstallment[j-1] += temp.PrincipalInstallment[i]
		}
	}

	// summarize
	r.PeriodMonthlyInstallment = append(r.PeriodMonthlyInstallment, temp.Installment[0])

	// assume all array are growing at same rate, so index can be re-used
	idx := len(r.PeriodSumInstallment)
	r.PeriodRowNum = append(r.PeriodRowNum, idx+1)

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
	var result FmtResult

	result.Principal = acfmt.FormatMoneyFloat64(r.Principal)

	for _, v := range r.Interests {
		result.Interests = append(result.Interests, accounting.FormatNumberFloat64(v, 2, ",", "."))
	}
	for _, v := range r.Periods {
		result.Periods = append(result.Periods, fmt.Sprintf("%d", v))
	}

	for _, v := range r.Installment {
		result.Installment = append(result.Installment, acfmt.FormatMoneyFloat64(v))
	}
	for _, v := range r.InterestInstallment {
		result.InterestInstallment = append(result.InterestInstallment, acfmt.FormatMoneyFloat64(v))
	}
	for _, v := range r.PrincipalInstallment {
		result.PrincipalInstallment = append(result.PrincipalInstallment, acfmt.FormatMoneyFloat64(v))
	}

	result.YearlyRowNum = append(result.YearlyRowNum, r.YearlyRowNum...)
	for _, v := range r.YearlyInstallment {
		result.YearlyInstallment = append(result.YearlyInstallment, acfmt.FormatMoneyFloat64(v))
	}
	for _, v := range r.YearlyInterestInstallment {
		result.YearlyInterestInstallment = append(result.YearlyInterestInstallment, acfmt.FormatMoneyFloat64(v))
	}
	for _, v := range r.YearlyPrincipalInstallment {
		result.YearlyPrincipalInstallment = append(result.YearlyPrincipalInstallment, acfmt.FormatMoneyFloat64(v))
	}

	result.PeriodRowNum = append(result.PeriodRowNum, r.PeriodRowNum...)
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
