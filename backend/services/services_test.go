package services

import (
	"math"
	"testing"

	"calc-app/backend/config"
)

func newTestServices(t *testing.T) (*HistoryService, *SettingsService) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	h := &HistoryService{}
	if err := h.Load(); err != nil {
		t.Fatal(err)
	}
	st := &SettingsService{}
	if err := st.Load(); err != nil {
		t.Fatal(err)
	}
	return h, st
}

func TestHistoryRoundtrip(t *testing.T) {
	h, _ := newTestServices(t)
	if _, err := h.Add("normal", "1+2", "3"); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Add("scientific", "sin(30)", "0.5"); err != nil {
		t.Fatal(err)
	}
	if got := len(h.List()); got != 2 {
		t.Fatalf("list length = %d, want 2", got)
	}
	// Newest first.
	if h.List()[0].Expr != "sin(30)" {
		t.Errorf("newest first failed: %+v", h.List()[0])
	}

	// Reload from disk into a fresh service.
	h2 := &HistoryService{}
	if err := h2.Load(); err != nil {
		t.Fatal(err)
	}
	if got := len(h2.List()); got != 2 {
		t.Fatalf("persisted length = %d, want 2", got)
	}
	if h2.List()[0].ID == h2.List()[1].ID {
		t.Error("ids not unique")
	}
	// nextID continues beyond the max persisted id.
	e, err := h2.Add("normal", "2+2", "4")
	if err != nil {
		t.Fatal(err)
	}
	if e.ID <= h2.List()[1].ID {
		t.Errorf("nextID not monotonic: %d", e.ID)
	}
}

func TestHistoryDeleteAndClear(t *testing.T) {
	h, _ := newTestServices(t)
	e1, _ := h.Add("normal", "1", "1")
	e2, _ := h.Add("normal", "2", "2")
	if err := h.Delete(e1.ID); err != nil {
		t.Fatal(err)
	}
	list := h.List()
	if len(list) != 1 || list[0].ID != e2.ID {
		t.Fatalf("delete failed: %+v", list)
	}
	if err := h.Clear(); err != nil {
		t.Fatal(err)
	}
	if len(h.List()) != 0 {
		t.Error("clear failed")
	}
}

func TestSettingsMergeAndPersist(t *testing.T) {
	_, st := newTestServices(t)
	got := st.Get()
	if got != DefaultSettings() {
		t.Fatalf("fresh settings = %+v", got)
	}

	next := DefaultSettings()
	next.Theme = "light"
	next.Mode = "finance"
	next.AngleUnit = "rad"
	if err := st.Save(next); err != nil {
		t.Fatal(err)
	}

	st2 := &SettingsService{}
	if err := st2.Load(); err != nil {
		t.Fatal(err)
	}
	if st2.Get() != next {
		t.Errorf("roundtrip = %+v, want %+v", st2.Get(), next)
	}
}

func TestCalcServiceScientific(t *testing.T) {
	s := &CalcService{}
	v, err := s.ScientificEvaluate("sin(30)", "deg")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(v-0.5) > 1e-9 {
		t.Errorf("sin(30) = %v", v)
	}
}

func TestCalcServiceDeveloper(t *testing.T) {
	s := &CalcService{}
	r, err := s.DevBinary("AND", "240", "60", 8, false)
	if err != nil {
		t.Fatal(err)
	}
	if r.Raw != "48" || r.Formatted.Hex != "30" || r.Formatted.Bin != "00110000" {
		t.Errorf("DevBinary = %+v", r)
	}

	r, err = s.DevParse("-1", 10, 8, true)
	if err != nil {
		t.Fatal(err)
	}
	if r.Raw != "255" || r.Formatted.Dec != "-1" || r.Formatted.Hex != "FF" {
		t.Errorf("DevParse(-1) = %+v", r)
	}

	r, err = s.DevUnary("NOT", "15", 8, false)
	if err != nil {
		t.Fatal(err)
	}
	if r.Raw != "240" {
		t.Errorf("NOT = %s", r.Raw)
	}
}

func TestCalcServiceDeveloperKeeps64BitPrecision(t *testing.T) {
	s := &CalcService{}
	const max = "18446744073709551615"
	r, err := s.DevFormat(max, 64, false)
	if err != nil {
		t.Fatal(err)
	}
	if r.Raw != max {
		t.Errorf("raw = %s, want %s", r.Raw, max)
	}
	if r.Formatted.Hex != "FFFFFFFFFFFFFFFF" || r.Formatted.Dec != max {
		t.Errorf("format = %+v", r.Formatted)
	}

	// Flip the top bit through the raw string without losing precision.
	r, err = s.DevBinary("XOR", max, "9223372036854775808", 64, false)
	if err != nil {
		t.Fatal(err)
	}
	if r.Formatted.Hex != "7FFFFFFFFFFFFFFF" {
		t.Errorf("xor result = %s", r.Formatted.Hex)
	}
}

func TestCalcServiceFinance(t *testing.T) {
	s := &CalcService{}
	sum, err := s.CalcLoan(100000, 12, 1, 12)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(sum.PaymentPerPeriod-8884.88) > 0.01 {
		t.Errorf("payment = %v", sum.PaymentPerPeriod)
	}
	roi, err := s.CalcROI(1000, 1500)
	if err != nil || roi != 50 {
		t.Errorf("roi = %v err = %v", roi, err)
	}
}

func TestConfigFileIsolation(t *testing.T) {
	// Services must respect XDG_CONFIG_HOME (set per-test above), so two
	// parallel test services never share state.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, err := config.Filepath("x.json")
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Error("empty path")
	}
}
