package finance

import (
	"math"
	"testing"
)

func TestLoanPaymentZeroRate(t *testing.T) {
	s, err := LoanPayment(1200, 0, 1, 12)
	if err != nil {
		t.Fatal(err)
	}
	if s.PaymentPerPeriod != 100 {
		t.Errorf("payment = %v, want 100", s.PaymentPerPeriod)
	}
	if s.TotalInterest != 0 {
		t.Errorf("interest = %v, want 0", s.TotalInterest)
	}
	if len(s.Schedule) != 12 {
		t.Errorf("schedule rows = %d, want 12", len(s.Schedule))
	}
}

func TestLoanPaymentStandard(t *testing.T) {
	// 100k, 12% annual, 1 year, monthly payments.
	s, err := LoanPayment(100000, 12, 1, 12)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(s.PaymentPerPeriod-8884.88) > 0.01 {
		t.Errorf("payment = %v, want ~8884.88", s.PaymentPerPeriod)
	}
	if math.Abs(s.TotalInterest-6618.55) > 0.1 {
		t.Errorf("total interest = %v, want ~6618.55", s.TotalInterest)
	}
	// Schedule must amortize to zero.
	last := s.Schedule[len(s.Schedule)-1]
	if last.Balance != 0 {
		t.Errorf("final balance = %v, want 0", last.Balance)
	}
	// Principal sums to the loan amount.
	var principal float64
	for _, p := range s.Schedule {
		principal += p.Principal
	}
	if math.Abs(principal-100000) > 0.01 {
		t.Errorf("principal sum = %v, want 100000", principal)
	}
}

func TestLoanPaymentErrors(t *testing.T) {
	if _, err := LoanPayment(0, 5, 1, 12); err == nil {
		t.Error("zero principal: expected error")
	}
	if _, err := LoanPayment(1000, -1, 1, 12); err == nil {
		t.Error("negative rate: expected error")
	}
	if _, err := LoanPayment(1000, 5, 0, 12); err == nil {
		t.Error("zero years: expected error")
	}
}

func TestSavingsFutureValue(t *testing.T) {
	r, err := SavingsFutureValue(100, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if r.FutureValue != 2400 || r.TotalInterest != 0 {
		t.Errorf("zero rate: %+v", r)
	}

	// $100/month for 1 year at 12% annual (1% monthly): FV = 1268.25
	r, err = SavingsFutureValue(100, 12, 1)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(r.FutureValue-1268.25) > 0.01 {
		t.Errorf("FV = %v, want ~1268.25", r.FutureValue)
	}
	if math.Abs(r.TotalDeposited-1200) > 0.001 {
		t.Errorf("deposited = %v, want 1200", r.TotalDeposited)
	}
}

func TestROIAndCAGR(t *testing.T) {
	roi, err := ROI(1000, 1500)
	if err != nil {
		t.Fatal(err)
	}
	if roi != 50 {
		t.Errorf("ROI = %v, want 50", roi)
	}

	cagr, err := CAGR(1000, 2000, 5)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(cagr-14.87) > 0.01 {
		t.Errorf("CAGR = %v, want ~14.87", cagr)
	}

	if _, err := ROI(0, 10); err == nil {
		t.Error("ROI(0,...) expected error")
	}
	if _, err := CAGR(-1, 10, 2); err == nil {
		t.Error("CAGR negative initial expected error")
	}
}

func TestPercentChange(t *testing.T) {
	got, err := PercentChange(100, 125)
	if err != nil {
		t.Fatal(err)
	}
	if got != 25 {
		t.Errorf("got %v, want 25", got)
	}
	if _, err := PercentChange(0, 1); err == nil {
		t.Error("expected error for zero base")
	}
}
