package main

import (
	"bytes"
	"errors"
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
	a.jqPeriodInput.On(jquery.CHANGE, a.updateFloatingPeriod)
	a.jqFixedPeriodInputs.On(jquery.CHANGE, a.updateFloatingPeriod)
	a.jqCalculateButton.On(jquery.CLICK, a.calculateResult)
}

func (a *App) updateFloatingPeriod(e jquery.Event) {
	_period, err := strconv.ParseInt(a.jqPeriodInput.Val(), 10, 64)
	if err != nil {
		// TODO: report error
		println("ERR fail to parse period: ", err.Error())
		return
	}
	// need to work with int instead of int64, bc gopherjs will translate int64 to object instead
	period := int(_period)

	fixedPeriod := 0
	a.jqFixedPeriodInputs.Each(func(i int, input interface{}) {
		jqin := jQuery(input)
		p, err := strconv.ParseInt(jqin.Val(), 10, 64)
		if err != nil {
			// TODO: report error
			println("ERR fail to parse fixed period: ", err.Error())
			return
		}
		fixedPeriod += int(p)
	})

	floatPeriod := period - fixedPeriod
	if floatPeriod < 0 {
		floatPeriod = 0
	}

	a.jqFloatPeriodInput.SetVal(floatPeriod)
}

func (a *App) calculateResult(e jquery.Event) {
	var (
		finalerr error
	)
	defer func() {
		if finalerr != nil {
			println("ERR " + finalerr.Error())
			e.PreventDefault()
		}
	}()

	var (
		period int
		price  int
		dp     float64
	)
	_price, err := strconv.ParseInt(a.jqPriceInput.Val(), 10, 64)
	if err != nil {
		finalerr = errors.New("fail to parse price " + err.Error())
		return
	}
	price = int(_price)

	dp, err = strconv.ParseFloat(a.jqDownPaymentInput.Val(), 64)
	if err != nil {
		finalerr = errors.New("fail to parse down payment " + err.Error())
		return
	}

	_period, err := strconv.ParseInt(a.jqPeriodInput.Val(), 10, 64)
	if err != nil {
		finalerr = errors.New("fail to parse period " + err.Error())
		return
	}
	period = int(_period)

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
		if err != nil {
			finalerr = errors.New("fail to parse fixed interest " + err.Error())
			return
		}
		fixedInterests = append(fixedInterests, it)
	})

	a.jqFixedPeriodInputs.Each(func(i int, input interface{}) {
		jqin := jQuery(input)
		p, err := strconv.ParseInt(jqin.Val(), 10, 64)
		if err != nil {
			finalerr = errors.New("fail to parse fixed period " + err.Error())
			return
		}
		sumFixedPeriod += int(p)
		fixedPeriods = append(fixedPeriods, int(p))
	})

	if len(fixedInterests) != len(fixedPeriods) {
		finalerr = errors.New("mismatched len of fixed interest and fixed period")
		return
	}

	floatPeriod = period - sumFixedPeriod
	if floatPeriod < 0 {
		finalerr = errors.New("fail to calculate floating period: doesn't add up")
		return
	}

	// only validates if valid
	if floatPeriod > 0 {
		floatInterest, err = strconv.ParseFloat(a.jqFloatInterestInput.Val(), 64)
		if err != nil {
			finalerr = errors.New("fail to parse float interest " + err.Error())
			return
		}
	}

	result := calculateResult(price, dp, period, fixedInterests, fixedPeriods, floatInterest, floatPeriod)
	a.renderResult(result)
}

func (a *App) renderResult(result Result) {
	fmtResult := result.format(a.acfmt)

	var b bytes.Buffer
	a.resultTemplate.Execute(&b, fmtResult)

	content := b.String()
	a.jqResult.SetHtml(content)
}
