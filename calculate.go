package main

import "math"

type Result struct {
	interests []float64
	periods   []int

	// breakdown
	installment          []float64
	interestInstallment  []float64
	principalInstallment []float64

	// per period
	periodMonthlyInstallment      []float64
	periodSumInstallment          []float64
	periodSumInterestInstallment  []float64
	periodSumPrincipalInstallment []float64

	// summary
	totalInstallment     float64
	totalInterests       float64
	totalPrincipal       float64
	principalBeforeFloat float64
}

func calculateResult(price int, downPayment float64, totalPeriod int, fixedInterest []float64, fixedPeriod []int, floatInterest float64, floatPeriod int) Result {
	var finalResult Result

	principal := float64(price) * (1 - downPayment/100)

	// calculate tiered fix
	for i := 0; i < len(fixedPeriod); i++ {
		// update pokok with remaining
		var _result Result
		principal, _result = calculate(principal, fixedInterest[i], fixedPeriod[i], totalPeriod)

		finalResult.interests = append(finalResult.interests, fixedInterest[i])
		finalResult.periods = append(finalResult.periods, fixedPeriod[i])
		finalResult.add(_result)

		// update values for next tiered fix
		totalPeriod = totalPeriod - fixedPeriod[i]
	}

	if floatPeriod > 0 {
		// calculate remaining as floating
		_, _result := calculate(principal, floatInterest, floatPeriod, totalPeriod)

		finalResult.principalBeforeFloat = principal
		finalResult.interests = append(finalResult.interests, floatInterest)
		finalResult.periods = append(finalResult.periods, floatPeriod)
		finalResult.add(_result)
	}

	return finalResult
}

func calculate(principal, interestRate float64, interestPeriod, totalPeriod int) (float64, Result) {
	var (
		monthlyInterestRate float64 = float64(interestRate) / 100 / 12

		// still don't fully understand, but this makes principal is paid in certain percentage to interest & installment
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

		// TODO: sanity check
		println("[%d] cicilan: %v, bunga: %v, pokok: %v\n",
			i+1, ac.FormatMoney(monthlyInstallment), ac.FormatMoney(interestInstallment), ac.FormatMoney(principalInstallment))

		result.installment = append(result.installment, monthlyInstallment)
		result.interestInstallment = append(result.interestInstallment, interestInstallment)
		result.principalInstallment = append(result.principalInstallment, principalInstallment)

		principal -= principalInstallment
	}

	// return any remainder and breakdown
	return principal, result
}

func (r *Result) add(temp Result) {
	r.periodMonthlyInstallment = append(r.periodMonthlyInstallment, temp.periodMonthlyInstallment...)

	r.installment = append(r.installment, temp.installment...)
	r.interestInstallment = append(r.interestInstallment, temp.interestInstallment...)
	r.principalInstallment = append(r.principalInstallment, temp.principalInstallment...)

	// summarize
	r.periodMonthlyInstallment = append(r.periodMonthlyInstallment, temp.installment[0])

	// assume all array are growing at same rate, so index can be re-used
	idx := len(r.periodSumInstallment)

	r.periodSumInstallment = append(r.periodSumInstallment, 0)
	for _, v := range temp.installment {
		r.periodSumInstallment[idx] += v
	}
	r.totalInstallment += r.periodSumInstallment[idx]

	r.periodSumInterestInstallment = append(r.periodSumInterestInstallment, 0)
	for _, v := range temp.interestInstallment {
		r.periodSumInterestInstallment[idx] += v
	}
	r.totalInterests += r.periodSumInterestInstallment[idx]

	r.periodSumPrincipalInstallment = append(r.periodSumPrincipalInstallment, 0)
	for _, v := range temp.principalInstallment {
		r.periodSumPrincipalInstallment[idx] += v
	}
	r.totalPrincipal += r.periodSumPrincipalInstallment[idx]
}
