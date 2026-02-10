package main

import (
	"reflect"
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

func TestResult_add(t *testing.T) {
	tests := []struct {
		name  string
		final Result
		new   Result
		want  Result
	}{
		{
			name:  "empty_final,new_12_months",
			final: Result{},
			new: Result{
				Interests: []float64{1},
				Periods:   []int{1},

				Installment:          []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				InterestInstallment:  []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1, 1.2},
				PrincipalInstallment: []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.07, 0.08, 0.09, 0.10, 0.11, 0.12},

				YearlyRowNum:               []int{12},
				YearlyInstallment:          []float64{78},
				YearlyInterestInstallment:  []float64{7.8},
				YearlyPrincipalInstallment: []float64{0.78},

				PeriodRowNum:                  []int{1},
				PeriodMonthlyInstallment:      []float64{1},
				PeriodSumInstallment:          []float64{1},
				PeriodSumInterestInstallment:  []float64{1},
				PeriodSumPrincipalInstallment: []float64{1},

				Principal:            1,
				TotalInstallment:     1,
				TotalInterests:       1,
				TotalPrincipal:       1,
				PrincipalBeforeFloat: 1,
			},
			want: Result{
				Interests: []float64{1},
				Periods:   []int{1},

				Installment:          []float64{1},
				InterestInstallment:  []float64{1},
				PrincipalInstallment: []float64{1},

				YearlyRowNum:               []int{1},
				YearlyInstallment:          []float64{1},
				YearlyInterestInstallment:  []float64{1},
				YearlyPrincipalInstallment: []float64{1},

				PeriodRowNum:                  []int{1},
				PeriodMonthlyInstallment:      []float64{1},
				PeriodSumInstallment:          []float64{1},
				PeriodSumInterestInstallment:  []float64{1},
				PeriodSumPrincipalInstallment: []float64{1},

				Principal:            1,
				TotalInstallment:     1,
				TotalInterests:       1,
				TotalPrincipal:       1,
				PrincipalBeforeFloat: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.final.add(tt.new)

			if !reflect.DeepEqual(tt.final, tt.want) {
				t.Errorf("Result.add() = %v, want %v", tt.final, tt.want)
			}
		})
	}
}
