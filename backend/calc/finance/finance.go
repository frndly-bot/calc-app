// Package finance implements the financial calculations behind the Finance
// mode: loan payments with amortization schedules, savings projections,
// ROI and CAGR.
package finance

import (
	"fmt"
	"math"
)

// LoanSummary is the result of a loan calculation.
type LoanSummary struct {
	PaymentPerPeriod float64   `json:"paymentPerPeriod"`
	NumPeriods       int       `json:"numPeriods"`
	TotalInterest    float64   `json:"totalInterest"`
	TotalPaid        float64   `json:"totalPaid"`
	Schedule         []Payment `json:"schedule"`
}

// Payment is one row of an amortization schedule.
type Payment struct {
	Number    int     `json:"number"`
	Payment   float64 `json:"payment"`
	Principal float64 `json:"principal"`
	Interest  float64 `json:"interest"`
	Balance   float64 `json:"balance"`
}

// LoanPayment computes the periodic payment and full amortization schedule
// for a fixed-rate loan. principal > 0, annualRate >= 0 (e.g. 6.5 for 6.5%),
// years > 0, paymentsPerYear > 0.
func LoanPayment(principal, annualRate float64, years, paymentsPerYear int) (LoanSummary, error) {
	if principal <= 0 {
		return LoanSummary{}, fmt.Errorf("principal must be positive")
	}
	if years <= 0 || paymentsPerYear <= 0 {
		return LoanSummary{}, fmt.Errorf("term and payments per year must be positive")
	}
	if annualRate < 0 {
		return LoanSummary{}, fmt.Errorf("interest rate cannot be negative")
	}
	n := years * paymentsPerYear
	r := annualRate / 100 / float64(paymentsPerYear)

	var payment float64
	if r == 0 {
		payment = principal / float64(n)
	} else {
		factor := math.Pow(1+r, float64(n))
		payment = principal * r * factor / (factor - 1)
	}

	schedule := make([]Payment, 0, n)
	balance := principal
	totalInterest := 0.0
	for i := 1; i <= n; i++ {
		interest := balance * r
		principalPart := payment - interest
		if i == n || principalPart > balance {
			// Final payment: settle exactly.
			principalPart = balance
			payment = principalPart + interest
		}
		balance -= principalPart
		totalInterest += interest
		schedule = append(schedule, Payment{
			Number:    i,
			Payment:   round2(payment),
			Principal: round2(principalPart),
			Interest:  round2(interest),
			Balance:   round2(math.Max(balance, 0)),
		})
	}

	return LoanSummary{
		PaymentPerPeriod: round2(payment),
		NumPeriods:       n,
		TotalInterest:    round2(totalInterest),
		TotalPaid:        round2(principal + totalInterest),
		Schedule:         schedule,
	}, nil
}

// SavingsResult is the result of a compound-savings projection.
type SavingsResult struct {
	FutureValue    float64 `json:"futureValue"`
	TotalDeposited float64 `json:"totalDeposited"`
	TotalInterest  float64 `json:"totalInterest"`
}

// SavingsFutureValue projects the balance of regular deposits earning
// compound interest. monthlyDeposit is deposited at the end of each month;
// annualRate is expressed in percent (e.g. 4.5).
func SavingsFutureValue(monthlyDeposit, annualRate float64, years int) (SavingsResult, error) {
	if monthlyDeposit <= 0 {
		return SavingsResult{}, fmt.Errorf("monthly deposit must be positive")
	}
	if years <= 0 {
		return SavingsResult{}, fmt.Errorf("years must be positive")
	}
	if annualRate < 0 {
		return SavingsResult{}, fmt.Errorf("interest rate cannot be negative")
	}
	r := annualRate / 100 / 12
	n := years * 12
	var fv float64
	if r == 0 {
		fv = monthlyDeposit * float64(n)
	} else {
		fv = monthlyDeposit * (math.Pow(1+r, float64(n)) - 1) / r
	}
	deposited := monthlyDeposit * float64(n)
	return SavingsResult{
		FutureValue:    round2(fv),
		TotalDeposited: round2(deposited),
		TotalInterest:  round2(fv - deposited),
	}, nil
}

// ROI returns the return on investment in percent: (final-initial)/initial*100.
func ROI(initial, final float64) (float64, error) {
	if initial == 0 {
		return 0, fmt.Errorf("initial value cannot be zero")
	}
	return round2((final - initial) / math.Abs(initial) * 100), nil
}

// CAGR returns the compound annual growth rate in percent between two values
// over a number of years.
func CAGR(initial, final float64, years int) (float64, error) {
	if initial <= 0 || final <= 0 {
		return 0, fmt.Errorf("initial and final values must be positive")
	}
	if years <= 0 {
		return 0, fmt.Errorf("years must be positive")
	}
	r := math.Pow(final/initial, 1/float64(years)) - 1
	return round2(r * 100), nil
}

// PercentChange returns the percentage change from a to b.
func PercentChange(from, to float64) (float64, error) {
	if from == 0 {
		return 0, fmt.Errorf("base value cannot be zero")
	}
	return round2((to - from) / math.Abs(from) * 100), nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
