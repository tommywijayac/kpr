package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"text/template"

	"github.com/gopherjs/jquery"
	"github.com/leekchan/accounting"
)

var jQuery = jquery.NewJQuery //for convenience

type App struct {
	acfmt accounting.Accounting

	resultTemplate *template.Template
	jqResult       jquery.JQuery

	jqPriceInput         jquery.JQuery
	jqDownPaymentInput   jquery.JQuery
	jqDownPaymentButtons jquery.JQuery
	jqPeriodInput        jquery.JQuery
	jqPeriodButtons      jquery.JQuery

	jqFixedInterestInputs []jquery.JQuery
	jqFixedPeriodInputs   []jquery.JQuery
	jqFloatInterestInput  jquery.JQuery
	jqFloatPeriodInput    jquery.JQuery
	jqCalculateButton     jquery.JQuery
}

func NewApp() *App {
	form := jQuery("form")
	resultHtml := jQuery("#result-template").Html()

	var jqFixedInterestInputs []jquery.JQuery
	form.Find("#fixedInterest").Find("input.interest").Each(func(i int, input interface{}) {
		jqFixedInterestInputs = append(jqFixedInterestInputs, jQuery(input))
	})
	var jqFixedPeriodInputs []jquery.JQuery
	form.Find("#fixedInterest").Find("input.period").Each(func(i int, input interface{}) {
		jqFixedPeriodInputs = append(jqFixedPeriodInputs, jQuery(input))
	})

	return &App{
		acfmt: accounting.Accounting{Symbol: "IDR ", Precision: 2},

		resultTemplate: template.Must(template.New("result").Parse(resultHtml)),
		jqResult:       jQuery("#result"),

		jqPriceInput:          form.Find("#price"),
		jqDownPaymentInput:    form.Find("#downPayment"),
		jqDownPaymentButtons:  form.Find("#easyInputDownPayment"),
		jqPeriodInput:         form.Find("#totalPeriod"),
		jqPeriodButtons:       form.Find("#easyInputPeriod"),
		jqFixedInterestInputs: jqFixedInterestInputs,
		jqFixedPeriodInputs:   jqFixedPeriodInputs,
		jqFloatInterestInput:  form.Find("#floatInterest"),
		jqFloatPeriodInput:    form.Find("#floatInterestPeriod"),
		jqCalculateButton:     form.Find("#calculate"),
	}
}

func (a *App) BindEvents() {
	a.jqPriceInput.On(jquery.KEYUP, a.onPriceKeyup)
	a.jqDownPaymentInput.On(jquery.KEYUP, a.onDownPaymentKeyup)
	a.jqDownPaymentButtons.On(jquery.CLICK, a.onDownPaymentClick)
	a.jqPeriodInput.On(jquery.CHANGE, a.onPeriodChange)
	a.jqPeriodButtons.On(jquery.CLICK, a.onPeriodClick)

	for i := range a.jqFixedPeriodInputs {
		a.jqFixedPeriodInputs[i].On(jquery.CHANGE, a.onPeriodChange)
	}

	a.jqCalculateButton.On(jquery.CLICK, a.onCalculate)
}

func (a *App) Render() {
	a.updatePriceFormatted(a.jqPriceInput)
	a.updateDownPaymentAmount(a.jqDownPaymentInput)
	a.updatePeriodInMonth(a.jqPeriodInput)
	for i := range a.jqFixedPeriodInputs {
		a.updatePeriodInMonth(a.jqFixedPeriodInputs[i])
	}
	a.updateFloatingPeriod()
}

// Event handler
func (a *App) onPriceKeyup(e jquery.Event) {
	el := jQuery(e.Target)
	a.updatePriceFormatted(el)
	a.updateDownPaymentAmount(a.jqDownPaymentInput)
}

func (a *App) onDownPaymentKeyup(e jquery.Event) {
	el := jQuery(e.Target)
	a.updateDownPaymentAmount(el)
}

func (a *App) onDownPaymentClick(e jquery.Event) {
	el := jQuery(e.Target)
	a.jqDownPaymentInput.SetVal(el.Val())
	a.updateDownPaymentAmount(a.jqDownPaymentInput)
}

func (a *App) onPeriodChange(e jquery.Event) {
	el := jQuery(e.Target)
	a.updatePeriodInMonth(el)
	a.updateFloatingPeriod()
}

func (a *App) onPeriodClick(e jquery.Event) {
	el := jQuery(e.Target)
	a.jqPeriodInput.SetVal(el.Val())
	a.updatePeriodInMonth(a.jqPeriodInput)
	a.updateFloatingPeriod()
}

func (a *App) onCalculate(e jquery.Event) {
	if err := a.calculateResult(); err != nil {
		e.PreventDefault()
		return
	}
}

// DOM logic
func (a *App) updatePriceFormatted(el jquery.JQuery) {
	price, _ := strconv.ParseFloat(el.Val(), 64) // if err, price is 0
	el.Parent().Next().Find("span").SetText(a.acfmt.FormatMoneyFloat64((price)))
}

func (a *App) updateDownPaymentAmount(el jquery.JQuery) {
	dp, _ := strconv.ParseFloat(el.Val(), 64)
	price, _ := strconv.ParseFloat(a.jqPriceInput.Val(), 64)
	principal := price * dp / 100 // if any err, principal is 0
	el.Parent().Next().Find("span").SetText(a.acfmt.FormatMoneyFloat64((principal)))
}

func (a *App) updatePeriodInMonth(el jquery.JQuery) {
	text := ""
	year, err := strconv.ParseInt(el.Val(), 10, 64)
	if err == nil {
		text = fmt.Sprintf("%d bulan", year*12)
	}
	el.Parent().Next().Find("span").SetText(text)
}

func (a *App) updateFloatingPeriod() {
	var (
		finalerr error
	)
	defer func() {
		if finalerr != nil {
			println("ERR " + finalerr.Error())
			return
		}
	}()

	_period, err := strconv.ParseInt(a.jqPeriodInput.Val(), 10, 64)
	if err != nil && len(a.jqPeriodInput.Val()) != 0 {
		finalerr = errors.New("fail to parse period " + err.Error())
		return
	}
	// need to work with int instead of int64, bc gopherjs will translate int64 to object instead
	period := int(_period)

	fixedPeriod := 0
	for i := range a.jqFixedPeriodInputs {
		p, err := strconv.ParseInt(a.jqFixedPeriodInputs[i].Val(), 10, 64)
		if err != nil && len(a.jqFixedPeriodInputs[i].Val()) != 0 {
			finalerr = errors.New("fail to parse fixed period " + err.Error())
			return
		}
		fixedPeriod += int(p)
	}

	floatPeriod := period - fixedPeriod

	a.jqFloatPeriodInput.SetVal(floatPeriod)
	a.jqFloatPeriodInput.Parent().Next().Find("span").SetText(fmt.Sprintf("%d bulan", floatPeriod*12))
}

func (a *App) calculateResult() error {
	var (
		finalerr error
	)
	defer func() {
		if finalerr != nil {
			println("ERR " + finalerr.Error())
		}
	}()

	var (
		period int
		price  float64
		dp     float64
	)

	price, err := strconv.ParseFloat(a.jqPriceInput.Val(), 64)
	if err != nil {
		finalerr = errors.New("fail to parse price " + err.Error())
		a.jqPriceInput.AddClass("is-invalid")
		return finalerr
	}
	a.jqPriceInput.RemoveClass("is-invalid")

	dp, err = strconv.ParseFloat(a.jqDownPaymentInput.Val(), 64)
	if err != nil {
		finalerr = errors.New("fail to parse down payment " + err.Error())
		a.jqDownPaymentInput.AddClass("is-invalid")
		return finalerr
	}
	a.jqDownPaymentInput.RemoveClass("is-invalid")

	_period, err := strconv.ParseInt(a.jqPeriodInput.Val(), 10, 64)
	if err != nil {
		finalerr = errors.New("fail to parse period " + err.Error())
		a.jqPeriodInput.AddClass("is-invalid")
		return finalerr
	}
	a.jqPeriodInput.RemoveClass("is-invalid")
	period = int(_period) * 12

	var (
		fixedInterests []float64
		fixedPeriods   []int
		sumFixedPeriod int
		floatInterest  float64
		floatPeriod    int
	)

	// check both interest and period, must form a pair
	// use interest to iterate
	for i := range a.jqFixedInterestInputs {
		interest, err := strconv.ParseFloat(a.jqFixedInterestInputs[i].Val(), 64)
		if err != nil && len(a.jqFixedInterestInputs[i].Val()) != 0 {
			finalerr = errors.New("fail to parse fixed interest " + err.Error())
			a.jqFixedInterestInputs[i].AddClass("is-invalid")
			return finalerr
		}

		period, err := strconv.ParseInt(a.jqFixedPeriodInputs[i].Val(), 10, 64)
		if err != nil && len(a.jqFixedPeriodInputs[i].Val()) != 0 {
			finalerr = errors.New("fail to parse fixed interest " + err.Error())
			a.jqFixedPeriodInputs[i].AddClass("is-invalid")
			return finalerr
		}

		if interest == 0 && period == 0 {
			continue
		} else if interest != 0 && period == 0 {
			a.jqFixedPeriodInputs[i].AddClass("is-invalid")
			continue
		} else if interest == 0 && period != 0 {
			a.jqFixedInterestInputs[i].AddClass("is-invalid")
			continue
		}

		a.jqFixedInterestInputs[i].RemoveClass("is-invalid")
		a.jqFixedPeriodInputs[i].RemoveClass("is-invalid")

		fixedInterests = append(fixedInterests, interest)

		period = period * 12
		sumFixedPeriod += int(period)
		fixedPeriods = append(fixedPeriods, int(period))
	}

	floatPeriod = period - sumFixedPeriod
	if floatPeriod < 0 {
		finalerr = errors.New("fail to calculate floating period: doesn't add up")

		// highlight last non-empty input in fixed period
		for i := len(a.jqFixedPeriodInputs) - 1; i >= 0; i-- {
			period, _ := strconv.ParseInt(a.jqFixedPeriodInputs[i].Val(), 10, 64)
			if period > 0 {
				a.jqFixedPeriodInputs[i].AddClass("is-invalid")
			}
		}

		return finalerr
	}

	// only validate interest if period is valid
	if floatPeriod > 0 {
		floatInterest, err = strconv.ParseFloat(a.jqFloatInterestInput.Val(), 64)
		if err != nil {
			finalerr = errors.New("fail to parse float interest " + err.Error())
			a.jqFloatInterestInput.AddClass("is-invalid")
			return finalerr
		}
		if floatInterest == 0 {
			finalerr = errors.New("float interest can't be 0")
			a.jqFloatInterestInput.AddClass("is-invalid")
			return finalerr
		}
	}
	a.jqFloatInterestInput.RemoveClass("is-invalid")

	result := calculateResult(price, dp, period, fixedInterests, fixedPeriods, floatInterest, floatPeriod)
	a.renderResult(result)

	return nil
}

func (a *App) renderResult(result Result) {
	fmtResult := result.format(a.acfmt)

	var b bytes.Buffer
	a.resultTemplate.Execute(&b, fmtResult)

	content := b.String()
	a.jqResult.SetHtml(content)
}
