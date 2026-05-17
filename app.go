package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/leekchan/accounting"
)

var jQuery = jquery.NewJQuery //for convenience

type App struct {
	acfmt accounting.Accounting

	resultTemplate    *template.Template
	breakdownTemplate *template.Template

	jqResult    jquery.JQuery
	jqBreakdown jquery.JQuery

	jqPriceInput         jquery.JQuery
	jqBankAppraisal      jquery.JQuery
	jqCredit             jquery.JQuery
	jqDeltaPriceToCredit jquery.JQuery
	jqDownPaymentInput   jquery.JQuery
	jqDownPaymentAmount  jquery.JQuery
	jqDownPaymentButtons jquery.JQuery
	jqPeriodInput        jquery.JQuery
	jqPeriodButtons      jquery.JQuery

	jqFixedInterestInputs []jquery.JQuery
	jqFixedPeriodInputs   []jquery.JQuery
	jqFloatInterestInput  jquery.JQuery
	jqFloatPeriodInput    jquery.JQuery
	jqCalculateButton     jquery.JQuery

	jqCopyResultButton    jquery.JQuery
	jqCopyBreakdownButton jquery.JQuery

	// internal usage
	jqSeed       jquery.JQuery
	jqPresetLoan jquery.JQuery
}

func NewApp() *App {
	form := jQuery("form")
	resultHtml := jQuery("#result-template").Html()
	breakdownHtml := jQuery("#breakdown-template").Html()

	var jqFixedInterestInputs []jquery.JQuery
	form.Find("#fixedInterest").Find("input.interest").Each(func(i int, input interface{}) {
		jqFixedInterestInputs = append(jqFixedInterestInputs, jQuery(input))
	})
	var jqFixedPeriodInputs []jquery.JQuery
	form.Find("#fixedInterest").Find("input.period").Each(func(i int, input interface{}) {
		jqFixedPeriodInputs = append(jqFixedPeriodInputs, jQuery(input))
	})

	return &App{
		acfmt: accounting.Accounting{},

		resultTemplate:    template.Must(template.New("result").Parse(resultHtml)),
		breakdownTemplate: template.Must(template.New("breakdown").Parse(breakdownHtml)),

		jqResult:    jQuery("#result"),
		jqBreakdown: jQuery("#breakdown"),

		jqPriceInput:          form.Find("#price"),
		jqBankAppraisal:       form.Find("#bankAppraisal"),
		jqCredit:              form.Find("#credit"),
		jqDeltaPriceToCredit:  form.Find("#deltaPriceCredit"),
		jqDownPaymentInput:    form.Find("#downPayment"),
		jqDownPaymentButtons:  form.Find("#easyInputDownPayment"),
		jqDownPaymentAmount:   form.Find("#downPaymentAmount"),
		jqPeriodInput:         form.Find("#totalPeriod"),
		jqPeriodButtons:       form.Find("#easyInputPeriod"),
		jqFixedInterestInputs: jqFixedInterestInputs,
		jqFixedPeriodInputs:   jqFixedPeriodInputs,
		jqFloatInterestInput:  form.Find("#floatInterest"),
		jqFloatPeriodInput:    form.Find("#floatInterestPeriod"),
		jqCalculateButton:     form.Find("#calculate"),

		jqCopyResultButton:    jQuery("#copyResult"),
		jqCopyBreakdownButton: jQuery("#copyBreakdown"),

		jqSeed:       jQuery("#seed"),
		jqPresetLoan: form.Find("#presetLoan"),
	}
}

func (a *App) BindEvents() {
	println("App BindEvents. Result Btn Len:", a.jqCopyResultButton.Length) // Debug log

	a.jqPriceInput.On("input", a.onPriceInput)
	a.jqBankAppraisal.On(jquery.KEYUP, a.onBankAppraisalKeyup)
	a.jqDownPaymentInput.On(jquery.KEYUP, a.onDownPaymentKeyup)
	a.jqDownPaymentButtons.On(jquery.CLICK, a.onDownPaymentClick)
	a.jqPeriodInput.On(jquery.CHANGE, a.onPeriodChange)
	a.jqPeriodButtons.On(jquery.CLICK, a.onPeriodClick)

	for i := range a.jqFixedPeriodInputs {
		a.jqFixedPeriodInputs[i].On(jquery.CHANGE, a.onPeriodChange)
	}

	a.jqCalculateButton.On(jquery.CLICK, a.onCalculate)
	a.jqCopyResultButton.On(jquery.CLICK, func(e jquery.Event) {
		a.copyToClipboard(a.jqResult)
	})
	a.jqCopyBreakdownButton.On(jquery.CLICK, func(e jquery.Event) {
		a.copyToClipboard(a.jqBreakdown)
	})

	a.jqSeed.On(jquery.CLICK, a.seed)
	a.jqPresetLoan.On(jquery.CHANGE, a.onPresetLoanChange)
}

func (a *App) copyToClipboard(jq jquery.JQuery) {
	doc := js.Global.Get("document")
	win := js.Global.Get("window")

	rangeObj := doc.Call("createRange")
	rangeObj.Call("selectNode", jq.Get(0))

	selection := win.Call("getSelection")
	selection.Call("removeAllRanges")
	selection.Call("addRange", rangeObj)

	success := doc.Call("execCommand", "copy").Bool()

	if success {
		toastEl := doc.Call("getElementById", "copyToast")
		bootstrap := js.Global.Get("bootstrap")
		if toastEl != nil && bootstrap != nil && bootstrap != js.Undefined {
			toast := bootstrap.Get("Toast").Call("getOrCreateInstance", toastEl)
			toast.Call("show")
		}
	}

	selection.Call("removeAllRanges")
}

func (a *App) Render() {
	a.updatePriceFormatted(a.jqPriceInput)
	a.updateBankAppraisalAmount(a.jqBankAppraisal)
	a.updateDownPaymentAmount(a.jqDownPaymentInput)
	a.updatePeriodInMonth(a.jqPeriodInput)
	for i := range a.jqFixedPeriodInputs {
		a.updatePeriodInMonth(a.jqFixedPeriodInputs[i])
	}
	a.updateFloatingPeriod()
}

// Event handler
func (a *App) onPriceInput(e jquery.Event) {
	// format text with comma for easier read
	a.updatePriceFormatted(jQuery(e.Target))

	// update related values
	a.updateBankAppraisalAmount(a.jqBankAppraisal)
	a.updateDownPaymentAmount(a.jqDownPaymentInput)
}

func (a *App) onBankAppraisalKeyup(e jquery.Event) {
	el := jQuery(e.Target)
	a.updateBankAppraisalAmount(el)
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

type loanPreset struct {
	fixedInterests []float64
	fixedPeriods   []int
	floatInterest  float64
}

var loanPresets = map[string]loanPreset{
	"tiered3": {
		fixedInterests: []float64{3.99, 7.99, 9.99},
		fixedPeriods:   []int{3, 3, 4},
		floatInterest:  11,
	},
	"tiered4": {
		fixedInterests: []float64{2.9, 5.99, 7.99, 9.99},
		fixedPeriods:   []int{1, 2, 3, 4},
		floatInterest:  11,
	},
}

func (a *App) onPresetLoanChange(e jquery.Event) {
	el := jQuery(e.Target)
	presetKey := el.Val()

	preset, ok := loanPresets[presetKey]
	if !ok {
		return
	}

	// Clear all fixed interest/period inputs
	for i := range a.jqFixedInterestInputs {
		a.jqFixedInterestInputs[i].SetVal("")
		a.jqFixedPeriodInputs[i].SetVal("")
		a.jqFixedInterestInputs[i].RemoveClass("is-invalid")
		a.jqFixedPeriodInputs[i].RemoveClass("is-invalid")
		a.updatePeriodInMonth(a.jqFixedPeriodInputs[i])
	}

	// Fill in preset values
	for i, interest := range preset.fixedInterests {
		if i >= len(a.jqFixedInterestInputs) {
			break
		}
		a.jqFixedInterestInputs[i].SetVal(fmt.Sprintf("%g", interest))
		a.jqFixedPeriodInputs[i].SetVal(preset.fixedPeriods[i])
		a.updatePeriodInMonth(a.jqFixedPeriodInputs[i])
	}

	// Fill float interest if specified by preset
	if preset.floatInterest > 0 {
		a.jqFloatInterestInput.SetVal(fmt.Sprintf("%g", preset.floatInterest))
	}

	a.updateFloatingPeriod()
}

// DOM logic
func (a *App) updatePriceFormatted(el jquery.JQuery) {
	// Strip all non-digit characters
	currentVal := el.Val()
	rawVal := ""
	for _, c := range currentVal {
		if c >= '0' && c <= '9' {
			rawVal += string(c)
		}
	}

	if rawVal == "" {
		return
	}

	price, _ := strconv.ParseFloat(rawVal, 64)
	formatted := accounting.FormatNumberFloat64(price, 0, ",", ".")
	el.SetVal(formatted)
}

func (a *App) updateBankAppraisalAmount(el jquery.JQuery) {
	bankAppraisal, _ := strconv.ParseFloat(el.Val(), 64)

	price, _ := strconv.ParseFloat(strings.ReplaceAll(a.jqPriceInput.Val(), ",", ""), 64)
	credit := price * bankAppraisal / 100 // if any err, credit is 0
	a.jqCredit.SetVal(a.acfmt.FormatMoneyFloat64(credit))
	a.jqDeltaPriceToCredit.SetVal(a.acfmt.FormatMoneyFloat64(price - credit))
}

func (a *App) updateDownPaymentAmount(el jquery.JQuery) {
	dp, _ := strconv.ParseFloat(el.Val(), 64)

	credit, _ := strconv.ParseFloat(strings.ReplaceAll(a.jqCredit.Val(), ",", ""), 64)
	principal := credit * dp / 100 // if any err, principal is 0
	a.jqDownPaymentAmount.SetVal(a.acfmt.FormatMoneyFloat64((principal)))
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
	price, err := strconv.ParseFloat(strings.ReplaceAll(a.jqPriceInput.Val(), ",", ""), 64)
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

	yearPeriod, err := strconv.ParseInt(a.jqPeriodInput.Val(), 10, 64)
	if err != nil {
		finalerr = errors.New("fail to parse period " + err.Error())
		a.jqPeriodInput.AddClass("is-invalid")
		return finalerr
	}
	a.jqPeriodInput.RemoveClass("is-invalid")
	period = int(yearPeriod) * 12

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

		fixedPeriod, err := strconv.ParseInt(a.jqFixedPeriodInputs[i].Val(), 10, 64)
		if err != nil && len(a.jqFixedPeriodInputs[i].Val()) != 0 {
			finalerr = errors.New("fail to parse fixed interest " + err.Error())
			a.jqFixedPeriodInputs[i].AddClass("is-invalid")
			return finalerr
		}

		if interest == 0 && fixedPeriod == 0 {
			continue
		} else if interest != 0 && fixedPeriod == 0 {
			a.jqFixedPeriodInputs[i].AddClass("is-invalid")
			continue
		} else if interest == 0 && fixedPeriod != 0 {
			a.jqFixedInterestInputs[i].AddClass("is-invalid")
			continue
		}

		a.jqFixedInterestInputs[i].RemoveClass("is-invalid")
		a.jqFixedPeriodInputs[i].RemoveClass("is-invalid")

		fixedInterests = append(fixedInterests, interest)

		fixedPeriod = fixedPeriod * 12
		sumFixedPeriod += int(fixedPeriod)
		fixedPeriods = append(fixedPeriods, int(fixedPeriod))
	}

	floatPeriod = period - sumFixedPeriod
	if floatPeriod < 0 {
		finalerr = errors.New("fail to calculate floating period: doesn't add up")

		// highlight last non-empty input in fixed period
		for i := len(a.jqFixedPeriodInputs) - 1; i >= 0; i-- {
			fixedPeriod, _ := strconv.ParseInt(a.jqFixedPeriodInputs[i].Val(), 10, 64)
			if fixedPeriod > 0 {
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

	result := calculateResult(MortgageSchema{
		Price:         price,
		DownPayment:   dp,
		TotalPeriod:   period,
		FixedInterest: fixedInterests,
		FixedPeriod:   fixedPeriods,
		FloatInterest: floatInterest,
		FloatPeriod:   floatPeriod,
	})
	a.renderResult(result)
	a.renderBreakdown(result)

	return nil
}

func (a *App) renderResult(result Result) {
	fmtResult := result.format(a.acfmt)

	var b bytes.Buffer
	a.resultTemplate.Execute(&b, fmtResult)
	content := b.String()
	a.jqResult.SetHtml(content)
}

func (a *App) renderBreakdown(result Result) {
	fmtResult := result.format(a.acfmt)

	var b bytes.Buffer
	a.breakdownTemplate.Execute(&b, fmtResult)
	content := b.String()
	a.jqBreakdown.SetHtml(content)
}

func (a *App) seed() {
	// mock values
	if a.jqPriceInput.Val() == "" {
		a.jqPriceInput.SetVal("1500000000")
		a.updatePriceFormatted(a.jqPriceInput)
	}
	if a.jqBankAppraisal.Val() == "" {
		a.jqBankAppraisal.SetVal("80")
		a.updateBankAppraisalAmount(a.jqBankAppraisal)
	}
	if a.jqDownPaymentInput.Val() == "" {
		a.jqDownPaymentInput.SetVal("20")
		a.updateDownPaymentAmount(a.jqDownPaymentInput)
	}
	if a.jqPeriodInput.Val() == "" {
		a.jqPeriodInput.SetVal("15")
		a.updatePeriodInMonth(a.jqPeriodInput)
	}
	// default fixed interest
	if len(a.jqFixedInterestInputs) > 0 && a.jqFixedInterestInputs[0].Val() == "" {
		a.jqFixedInterestInputs[0].SetVal("8.75")
		a.jqFixedPeriodInputs[0].SetVal("10")
		a.updatePeriodInMonth(a.jqFixedPeriodInputs[0])
	}
	// update float period after setting mock values
	a.updateFloatingPeriod()

	// default float interest
	if a.jqFloatInterestInput.Val() == "" {
		a.jqFloatInterestInput.SetVal("11.0")
	}
}
