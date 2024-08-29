package main

import (
	"github.com/leekchan/accounting"
)

var ac accounting.Accounting

func main() {
	app := NewApp()
	app.BindEvents()

	// ac = accounting.Accounting{Symbol: "IDR ", Precision: 2}

	// hardcoded inputs
	// price := 1400000000
	// dp := float64(10)
	// bungafix := []float64{2.79, 5.79, 8.1, 10.1}
	// periodefix := []int{12, 24, 36, 48} // len bungafix == len periodefix
	// periodetotal := 120                 // can be bigger than total fix, means remaining is floating. but can't be less than total fix

	// bungafloat := float64(11)
	// periodefloat := periodetotal
	// for i := 0; i < len(periodefix); i++ {
	// 	periodefloat -= periodefix[i]
	// }
	// if periodefloat < 0 {
	// 	panic("invalid period")
	// }

	// calculateResult(price, dp, periodetotal, bungafix, periodefix, bungafloat, periodefloat)
}
