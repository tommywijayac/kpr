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

	jqPriceInput          jquery.JQuery
	jqDownPaymentInput    jquery.JQuery
	jqPeriodInput         jquery.JQuery
	jqFixedInterestInputs jquery.JQuery
	jqFixedPeriodInputs   jquery.JQuery
	jqFloatInterestInput  jquery.JQuery
	jqFloatPeriodInput    jquery.JQuery
	jqCalculateButton     jquery.JQuery
}

func NewApp() *App {
	form := jQuery("form")
	resultHtml := jQuery("#result-template").Html()

	return &App{
		acfmt: accounting.Accounting{Symbol: "IDR ", Precision: 2},

		resultTemplate: template.Must(template.New("result").Parse(resultHtml)),
		jqResult:       jQuery("#result"),

		jqPriceInput:          form.Find("#price"),
		jqDownPaymentInput:    form.Find("#downPayment"),
		jqPeriodInput:         form.Find("#totalPeriod"),
		jqFixedInterestInputs: form.Find("#fixedInterest").Find("input.interest"),
		jqFixedPeriodInputs:   form.Find("#fixedInterest").Find("input.period"),
		jqFloatInterestInput:  form.Find("#floatInterest"),
		jqFloatPeriodInput:    form.Find("#floatInterestPeriod"),
		jqCalculateButton:     form.Find("#calculate"),
	}
}

func (a *App) BindEvents() {
	a.jqPriceInput.On(jquery.KEYUP, a.onPriceKeyup)
	a.jqDownPaymentInput.On(jquery.KEYUP, a.onDownPaymentKeyup)

	a.jqPeriodInput.On(jquery.CHANGE, a.onPeriodChange)
	a.jqFixedPeriodInputs.On(jquery.CHANGE, a.onPeriodChange)

	a.jqCalculateButton.On(jquery.CLICK, a.onCalculate)
}

func (a *App) Render() {
	a.updatePriceFormatted(a.jqPriceInput)
	a.updateDownPaymentAmount(a.jqDownPaymentInput)
	a.updatePeriodInMonth(a.jqPeriodInput)
	a.jqFixedPeriodInputs.Each(func(i int, input interface{}) {
		a.updatePeriodInMonth(jQuery(input))
	})
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

func (a *App) onPeriodChange(e jquery.Event) {
	el := jQuery(e.Target)
	a.updatePeriodInMonth(el)
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
	a.jqFixedPeriodInputs.Each(func(i int, input interface{}) {
		jqin := jQuery(input)
		p, err := strconv.ParseInt(jqin.Val(), 10, 64)
		if err != nil && len(jqin.Val()) != 0 {
			finalerr = errors.New("fail to parse fixed period " + err.Error())
			return
		}
		fixedPeriod += int(p)
	})

	floatPeriod := period - fixedPeriod
	if floatPeriod < 0 {
		floatPeriod = 0
	}

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
		return finalerr
	}

	dp, err = strconv.ParseFloat(a.jqDownPaymentInput.Val(), 64)
	if err != nil {
		finalerr = errors.New("fail to parse down payment " + err.Error())
		return finalerr
	}

	_period, err := strconv.ParseInt(a.jqPeriodInput.Val(), 10, 64)
	if err != nil {
		finalerr = errors.New("fail to parse period " + err.Error())
		return finalerr
	}
	period = int(_period) * 12

	var (
		fixedInterests []float64
		fixedPeriods   []int
		sumFixedPeriod int
		floatInterest  float64
		floatPeriod    int
	)
	a.jqFixedInterestInputs.Each(func(i int, input interface{}) {
		jqin := jQuery(input)
		it, err := strconv.ParseFloat(jqin.Val(), 64)
		if err != nil && len(jqin.Val()) != 0 {
			finalerr = errors.New("fail to parse fixed interest " + err.Error())
			return
		}

		if it == 0 {
			return
		}

		fixedInterests = append(fixedInterests, it)
	})

	a.jqFixedPeriodInputs.Each(func(i int, input interface{}) {
		jqin := jQuery(input)
		p, err := strconv.ParseInt(jqin.Val(), 10, 64)
		if err != nil && len(jqin.Val()) != 0 {
			finalerr = errors.New("fail to parse fixed period " + err.Error())
			return
		}

		if p == 0 {
			return
		}

		p = p * 12
		sumFixedPeriod += int(p)
		fixedPeriods = append(fixedPeriods, int(p))
	})

	if len(fixedInterests) != len(fixedPeriods) {
		finalerr = errors.New("mismatched len of fixed interest and fixed period")
		return finalerr
	}

	floatPeriod = period - sumFixedPeriod
	if floatPeriod < 0 {
		finalerr = errors.New("fail to calculate floating period: doesn't add up")
		return finalerr
	}

	// only validates if valid
	if floatPeriod > 0 {
		floatInterest, err = strconv.ParseFloat(a.jqFloatInterestInput.Val(), 64)
		if err != nil {
			finalerr = errors.New("fail to parse float interest " + err.Error())
			return finalerr
		}
	}

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
