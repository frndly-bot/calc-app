// Package scientific evaluates scientific expressions: trigonometry with
// configurable angle units, logarithms, powers, roots, factorial and the
// common constants. It is built on the shared parser package.
package scientific

import (
	"fmt"
	"math"
	"math/big"

	"calc-app/backend/calc/parser"
)

// AngleMode selects how trigonometric function arguments and results are
// interpreted.
type AngleMode string

const (
	Degrees AngleMode = "deg"
	Radians AngleMode = "rad"
	Grads   AngleMode = "grad"
)

const (
	degToRad  = math.Pi / 180
	gradToRad = math.Pi / 200
)

// funcs maps every supported function name to its declared arity.
var funcs = map[string]int{
	"sin": 1, "cos": 1, "tan": 1,
	"asin": 1, "acos": 1, "atan": 1, "atan2": 2,
	"sinh": 1, "cosh": 1, "tanh": 1, "asinh": 1, "acosh": 1, "atanh": 1,
	"ln": 1, "log": 1, "log2": 1, "exp": 1, "exp10": 1,
	"sqrt": 1, "cbrt": 1, "root": 2, "abs": 1, "sign": 1,
	"mod": 2, "pow": 2, "hypot": 2, "fact": 1,
	"round": 1, "floor": 1, "ceil": 1, "trunc": 1,
	"min": 2, "max": 2,
}

// Evaluate parses and evaluates a scientific expression. The angle mode only
// affects trigonometric functions and their inverses.
func Evaluate(expr string, mode AngleMode) (float64, error) {
	switch mode {
	case Degrees, Radians, Grads:
	default:
		return 0, fmt.Errorf("unknown angle mode %q", mode)
	}
	env := parser.Env{
		LookupFunc:  lookupFunc(mode),
		LookupConst: lookupConst,
	}
	result, err := parser.EvalString(expr, funcs, env)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(result) {
		return 0, fmt.Errorf("result is not a number")
	}
	if math.IsInf(result, 0) {
		return 0, fmt.Errorf("result overflows")
	}
	return result, nil
}

func lookupConst(name string) (float64, bool) {
	switch name {
	case "pi":
		return math.Pi, true
	case "tau":
		return 2 * math.Pi, true
	case "e":
		return math.E, true
	case "phi":
		return math.Phi, true
	}
	return 0, false
}

// lookupFunc builds the function registry for the given angle mode and
// returns an arity-aware resolver for it.
func lookupFunc(mode AngleMode) func(name string, arity int) (parser.Func, bool) {
	toRad := func(x float64) float64 { return x * degToRad }
	fromRad := func(r float64) float64 { return r / degToRad }
	switch mode {
	case Grads:
		toRad = func(x float64) float64 { return x * gradToRad }
		fromRad = func(r float64) float64 { return r / gradToRad }
	case Radians:
		toRad = func(x float64) float64 { return x }
		fromRad = func(r float64) float64 { return r }
	}

	check1 := func(name string, f func(float64) float64) parser.Func {
		return func(args []float64) (float64, error) {
			if len(args) != 1 {
				return 0, fmt.Errorf("%s expects 1 argument", name)
			}
			return f(args[0]), nil
		}
	}
	check2 := func(name string, f func(a, b float64) float64) parser.Func {
		return func(args []float64) (float64, error) {
			if len(args) != 2 {
				return 0, fmt.Errorf("%s expects 2 arguments", name)
			}
			return f(args[0], args[1]), nil
		}
	}

	registry := map[string]parser.Func{
		"sin": check1("sin", func(x float64) float64 { return math.Sin(toRad(x)) }),
		"cos": check1("cos", func(x float64) float64 { return math.Cos(toRad(x)) }),
		"tan": check1("tan", func(x float64) float64 { return math.Tan(toRad(x)) }),
		"asin": check1("asin", func(x float64) float64 {
			if x < -1 || x > 1 {
				return math.NaN()
			}
			return fromRad(math.Asin(x))
		}),
		"acos": check1("acos", func(x float64) float64 {
			if x < -1 || x > 1 {
				return math.NaN()
			}
			return fromRad(math.Acos(x))
		}),
		"atan":  check1("atan", func(x float64) float64 { return fromRad(math.Atan(x)) }),
		"atan2": check2("atan2", func(a, b float64) float64 { return fromRad(math.Atan2(a, b)) }),
		"sinh":  check1("sinh", math.Sinh),
		"cosh":  check1("cosh", math.Cosh),
		"tanh":  check1("tanh", math.Tanh),
		"asinh": check1("asinh", math.Asinh),
		"acosh": check1("acosh", func(x float64) float64 {
			if x < 1 {
				return math.NaN()
			}
			return math.Acosh(x)
		}),
		"atanh": check1("atanh", func(x float64) float64 {
			if x <= -1 || x >= 1 {
				return math.NaN()
			}
			return math.Atanh(x)
		}),
		"ln": check1("ln", func(x float64) float64 {
			if x <= 0 {
				return math.NaN()
			}
			return math.Log(x)
		}),
		"log": check1("log", func(x float64) float64 {
			if x <= 0 {
				return math.NaN()
			}
			return math.Log10(x)
		}),
		"log2": check1("log2", func(x float64) float64 {
			if x <= 0 {
				return math.NaN()
			}
			return math.Log2(x)
		}),
		"exp":   check1("exp", math.Exp),
		"exp10": check1("exp10", func(x float64) float64 { return math.Pow(10, x) }),
		"sqrt": check1("sqrt", func(x float64) float64 {
			if x < 0 {
				return math.NaN()
			}
			return math.Sqrt(x)
		}),
		"cbrt": check1("cbrt", math.Cbrt),
		"root": check2("root", func(x, n float64) float64 {
			if n == 0 {
				return math.NaN()
			}
			if x < 0 && math.Mod(n, 2) == 1 {
				return -math.Pow(-x, 1/n)
			}
			return math.Pow(x, 1/n)
		}),
		"abs": check1("abs", math.Abs),
		"sign": check1("sign", func(x float64) float64 {
			switch {
			case x > 0:
				return 1
			case x < 0:
				return -1
			}
			return 0
		}),
		"mod":   check2("mod", math.Mod),
		"pow":   check2("pow", math.Pow),
		"hypot": check2("hypot", math.Hypot),
		"fact":  check1("fact", Factorial),
		"round": check1("round", math.Round),
		"floor": check1("floor", math.Floor),
		"ceil":  check1("ceil", math.Ceil),
		"trunc": check1("trunc", math.Trunc),
		"min":   check2("min", math.Min),
		"max":   check2("max", math.Max),
	}

	return func(name string, arity int) (parser.Func, bool) {
		fn, ok := registry[name]
		if !ok {
			return nil, false
		}
		// Validate arity here as well, so paren-less calls are checked too.
		expected := funcs[name]
		if arity != expected {
			return func([]float64) (float64, error) {
				return 0, fmt.Errorf("%s expects %d argument(s) but got %d", name, expected, arity)
			}, true
		}
		return fn, true
	}
}

// Factorial computes n! for integer values 0..1000 using big.Int and returns
// the result as float64. An error-free NaN signals an out-of-domain input.
func Factorial(x float64) float64 {
	if x < 0 || x != math.Trunc(x) || x > 1000 {
		return math.NaN()
	}
	n := int64(x)
	r := new(big.Int).MulRange(1, n)
	f := new(big.Float).SetInt(r)
	v, _ := f.Float64()
	return v
}
