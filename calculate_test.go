package main

import (
	"testing"
)

func TestCalculate(t *testing.T) {
	// calling calculate(principal, interestRate, interestPeriod, totalPeriod)
	// Example: 120,000,000 principal, 5% interest, 2 years (24 months), 2 years total
	principal := 120000000.0
	interestRate := 5.0
	interestPeriod := 24
	totalPeriod := 24

	_, result := calculate(principal, interestRate, interestPeriod, totalPeriod)

	if len(result.YearlyInstallment) == 0 {
		t.Errorf("Expected YearlyInstallment to be populated, but got empty")
	} else {
		t.Logf("YearlyInstallment: %v", result.YearlyInstallment)
	}
}
