// Package services exposes the calculator engines to the frontend as Wails
// v3 services. Method names and signatures become the TypeScript binding
// surface; every method is also callable from Go tests.
package services

import (
	"fmt"
	"strconv"
	"strings"

	"calc-app/backend/calc/developer"
	"calc-app/backend/calc/finance"
	"calc-app/backend/calc/scientific"
)

// CalcService fronts the scientific, developer and finance engines.
type CalcService struct{}

// ScientificEvaluate evaluates a scientific expression in the given angle
// mode ("deg", "rad" or "grad").
func (s *CalcService) ScientificEvaluate(expr, angleMode string) (float64, error) {
	return scientific.Evaluate(expr, scientific.AngleMode(angleMode))
}

// DevResult is the result of a developer-mode operation: the raw word plus
// its rendering in all four bases. Raw is a decimal string so 64-bit values
// survive the JSON boundary without losing precision in JavaScript.
type DevResult struct {
	Raw       string              `json:"raw"`
	Value     float64             `json:"value"` // signed or unsigned interpretation
	Formatted developer.Formatted `json:"formatted"`
}

// DevParse parses a value in the given base (2, 8, 10, 16) and returns all
// representations of the word.
func (s *CalcService) DevParse(text string, base int, bits int, signed bool) (DevResult, error) {
	ws, err := developer.ParseWordSize(bits)
	if err != nil {
		return DevResult{}, err
	}
	raw, err := developer.FromString(text, base, ws, signed)
	if err != nil {
		return DevResult{}, err
	}
	return s.devResult(raw, ws, signed), nil
}

// DevBinary applies a binary operator (AND, OR, XOR, NAND, NOR, ADD, SUB,
// MUL, DIV, MOD, SHL, SHR, SAR, ROL, ROR) to two words given as decimal
// strings.
func (s *CalcService) DevBinary(op string, a, b string, bits int, signed bool) (DevResult, error) {
	ws, err := developer.ParseWordSize(bits)
	if err != nil {
		return DevResult{}, err
	}
	araw, err := parseRaw(a)
	if err != nil {
		return DevResult{}, err
	}
	braw, err := parseRaw(b)
	if err != nil {
		return DevResult{}, err
	}
	raw, err := developer.Binary(op, araw, braw, ws, signed)
	if err != nil {
		return DevResult{}, err
	}
	return s.devResult(raw, ws, signed), nil
}

// DevUnary applies a unary operator ("NOT" or "NEG") to a word given as a
// decimal string.
func (s *CalcService) DevUnary(op string, a string, bits int, signed bool) (DevResult, error) {
	ws, err := developer.ParseWordSize(bits)
	if err != nil {
		return DevResult{}, err
	}
	araw, err := parseRaw(a)
	if err != nil {
		return DevResult{}, err
	}
	var raw uint64
	switch op {
	case "NOT":
		raw, err = developer.Not(araw, ws)
	case "NEG":
		raw, err = developer.Negate(araw, ws)
	default:
		return DevResult{}, fmt.Errorf("unknown unary operator %q", op)
	}
	if err != nil {
		return DevResult{}, err
	}
	return s.devResult(raw, ws, signed), nil
}

// DevFormat renders a word given as a decimal string in all four bases.
func (s *CalcService) DevFormat(raw string, bits int, signed bool) (DevResult, error) {
	ws, err := developer.ParseWordSize(bits)
	if err != nil {
		return DevResult{}, err
	}
	v, err := parseRaw(raw)
	if err != nil {
		return DevResult{}, err
	}
	return s.devResult(v, ws, signed), nil
}

func parseRaw(s string) (uint64, error) {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid raw value %q", s)
	}
	return v, nil
}

func (s *CalcService) devResult(raw uint64, ws developer.WordSize, signed bool) DevResult {
	_, value := developer.Interpret(raw, ws, signed)
	return DevResult{
		Raw:       strconv.FormatUint(raw, 10),
		Value:     value,
		Formatted: developer.Format(raw, ws, signed),
	}
}

// LoanSummary is re-exported so the binding generator emits it once.
type LoanSummary = finance.LoanSummary
type Payment = finance.Payment

// CalcLoan computes the periodic payment and amortization schedule of a
// fixed-rate loan.
func (s *CalcService) CalcLoan(principal, annualRate float64, years, paymentsPerYear int) (LoanSummary, error) {
	return finance.LoanPayment(principal, annualRate, years, paymentsPerYear)
}

// CalcSavings projects the future value of monthly deposits.
func (s *CalcService) CalcSavings(monthlyDeposit, annualRate float64, years int) (SavingsResult, error) {
	return finance.SavingsFutureValue(monthlyDeposit, annualRate, years)
}

// SavingsResult mirrors finance.SavingsResult for the binding surface.
type SavingsResult = finance.SavingsResult

// CalcROI returns the percentage return on investment.
func (s *CalcService) CalcROI(initial, final float64) (float64, error) {
	return finance.ROI(initial, final)
}

// CalcCAGR returns the compound annual growth rate in percent.
func (s *CalcService) CalcCAGR(initial, final float64, years int) (float64, error) {
	return finance.CAGR(initial, final, years)
}

// CalcPercentChange returns the percentage change between two values.
func (s *CalcService) CalcPercentChange(from, to float64) (float64, error) {
	return finance.PercentChange(from, to)
}
