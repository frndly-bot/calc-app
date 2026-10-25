package parser

import (
	"math"
	"testing"
)

func basicEnv() Env {
	return Env{
		LookupFunc: func(name string, arity int) (Func, bool) {
			switch name {
			case "sqrt":
				if arity == 1 {
					return func(a []float64) (float64, error) { return math.Sqrt(a[0]), nil }, true
				}
			case "add":
				if arity == 2 {
					return func(a []float64) (float64, error) { return a[0] + a[1], nil }, true
				}
			case "fact":
				if arity == 1 {
					return func(a []float64) (float64, error) {
						f := 1.0
						for i := 2; i <= int(a[0]); i++ {
							f *= float64(i)
						}
						return f, nil
					}, true
				}
			}
			return nil, false
		},
		LookupConst: func(name string) (float64, bool) {
			switch name {
			case "pi":
				return math.Pi, true
			case "e":
				return math.E, true
			}
			return 0, false
		},
	}
}

func funcs() map[string]int {
	return map[string]int{"sqrt": 1, "add": 2, "fact": 1}
}

func TestEvalString(t *testing.T) {
	env := basicEnv()
	cases := []struct {
		expr string
		want float64
	}{
		{"1+2", 3},
		{"2*3+4", 10},
		{"2+3*4", 14},
		{"(2+3)*4", 20},
		{"10/4", 2.5},
		{"2^10", 1024},
		{"-2^2", -4}, // unary minus binds looser than ^
		{"(-2)^2", 4},
		{"2^-2", 0.25},
		{"7 % 3", 1},
		{"3!", 6},
		{"(2+1)!", 6},
		{"3!^2", 36},  // postfix binds tighter than ^
		{"sqrt 9", 3}, // paren-less call
		{"sqrt(16)+1", 5},
		{"add(2,3)", 5},
		{"2(3+4)", 14}, // implicit multiplication
		{"2pi", 2 * math.Pi},
		{"(1+2)(4)", 12}, // implicit mult between parens
		{"pi/2", math.Pi / 2},
		{"1.5*2", 3},
	}

	for _, c := range cases {
		got, err := EvalString(c.expr, funcs(), env)
		if err != nil {
			t.Errorf("EvalString(%q) error: %v", c.expr, err)
			continue
		}
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("EvalString(%q) = %v, want %v", c.expr, got, c.want)
		}
	}
}

func TestEvalErrors(t *testing.T) {
	env := basicEnv()
	cases := []string{
		"1+",     // trailing operator
		"(1+2",   // missing close paren
		"1/0",    // division by zero
		"1%0",    // modulo by zero
		"foo(1)", // unknown function
		"bar",    // unknown identifier
		"add(1)", // wrong arity
		"1 2 +",  // malformed
		"2**3**", // trailing operator
	}
	for _, expr := range cases {
		if _, err := EvalString(expr, funcs(), env); err == nil {
			t.Errorf("EvalString(%q) expected error, got none", expr)
		}
	}
}

func TestPrecedence(t *testing.T) {
	env := basicEnv()
	got, err := EvalString("2+3*4^2", funcs(), env)
	if err != nil {
		t.Fatal(err)
	}
	want := 2 + 3*math.Pow(4, 2)
	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}
