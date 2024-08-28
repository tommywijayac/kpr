package main

import (
	"github.com/tommywijayac/kpr/internal"
)

func main() {
	// hardcoded inputs
	pokok := float64(1296000000)
	bungafix := []float64{2.79, 5.79, 8.1, 10.1}
	periodefix := []int{12, 24, 36, 48} // len bungafix == len periodefix
	periodetotal := 120                 // can be bigger than total fix, means remaining is floating. but can't be less than total fix

	bungafloat := 11
	periodefloat := periodetotal
	for i := 0; i < len(periodefix); i++ {
		periodefloat -= periodefix[i]
	}
	if periodefloat < 0 {
		panic("invalid period")
	}

	internal := internal.New()

	// calculate tiered fix
	for i := 0; i < len(periodefix); i++ {
		pokok = internal.Calculate(pokok, bungafix[i], periodefix[i], periodetotal) // update pokok with remaining

		// update values for next tiered fix
		periodetotal = periodetotal - periodefix[i]
	}

	// calculate remaining as floating
	internal.Calculate(pokok, float64(bungafloat), periodefloat, periodetotal)
}
