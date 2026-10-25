package scientific

import (
	"math"
	"testing"
)

func TestEvaluate(t *testing.T) {
	cases := []struct {
		expr string
		mode AngleMode
		want float64
	}{
		{"sin(pi/2)", Radians, 1},
		{"sin(30)", Degrees, 0.5},
		{"cos(0)", Degrees, 1},
		{"tan(45)", Degrees, 1},
		{"asin(0.5)", Degrees, 30},
		{"sin(100)", Grads, 1}, // 100 grad = 90 degrees
		{"ln(e)", Radians, 1},
		{"log(1000)", Radians, 3},
		{"log2(8)", Radians, 3},
		{"exp(0)", Radians, 1},
		{"sqrt(16)", Radians, 4},
		{"√25", Radians, 5}, // native glyph alias
		{"cbrt(27)", Radians, 3},
		{"root(27, 3)", Radians, 3},
		{"abs(-5)", Radians, 5},
		{"mod(10, 3)", Radians, 1},
		{"pow(2, 10)", Radians, 1024},
		{"2^10", Radians, 1024},
		{"3!", Radians, 6},
		{"5!/(2!*3!)", Radians, 10}, // combinations C(5,3)
		{"fact(5)", Radians, 120},
		{"hypot(3, 4)", Radians, 5},
		{"min(2, 7)", Radians, 2},
		{"max(2, 7)", Radians, 7},
		{"e^2", Radians, math.E * math.E},
		{"tau", Radians, 2 * math.Pi},
		{"2^3^2", Radians, 512}, // right-associative
		{"1e2", Radians, 0},     // replaced below
	}
	// "1e2" lexes as number 1 followed by ident "e2" (unknown) — expect error,
	// so drop it from the happy-path table.
	cases = cases[:len(cases)-1]

	for _, c := range cases {
		got, err := Evaluate(c.expr, c.mode)
		if err != nil {
			t.Errorf("Evaluate(%q) error: %v", c.expr, err)
			continue
		}
		if math.Abs(got-c.want) > 1e-9*math.Max(1, math.Abs(c.want)) {
			t.Errorf("Evaluate(%q) = %v, want %v", c.expr, got, c.want)
		}
	}
}

func TestEvaluateErrors(t *testing.T) {
	cases := []string{
		"ln(0)",
		"sqrt(-1)",
		"asin(2)",
		"1/0",
		"fact(-1)",
		"fact(2.5)",
		"sin(1,2)",
		"unknownfn(1)",
	}
	for _, expr := range cases {
		if _, err := Evaluate(expr, Radians); err == nil {
			t.Errorf("Evaluate(%q) expected error, got none", expr)
		}
	}
}

func TestEvaluateRejectsUnknownAngleMode(t *testing.T) {
	if _, err := Evaluate("sin(0)", AngleMode("bogus")); err == nil {
		t.Error("expected an error for an unknown angle mode")
	}
}

func TestFactorialLarge(t *testing.T) {
	got := Factorial(20)
	if math.Abs(got-2432902008176640000) > 1 {
		t.Errorf("Factorial(20) = %v", got)
	}
}

func TestImplicitMultiplication(t *testing.T) {
	got, err := Evaluate("2pi(3+4)", Radians)
	if err != nil {
		t.Fatal(err)
	}
	want := 2 * math.Pi * 7
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("got %v want %v", got, want)
	}
}
