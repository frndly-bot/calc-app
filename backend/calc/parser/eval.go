package parser

import (
	"fmt"
	"math"
)

// Func is a callable function in an expression. It receives already-evaluated
// arguments in call order.
type Func func(args []float64) (float64, error)

// Env supplies constants and function implementations for evaluation.
type Env struct {
	// LookupFunc returns the implementation for a function called with the
	// given arity, or false when unknown.
	LookupFunc func(name string, arity int) (Func, bool)
	// LookupConst returns the value of a named constant, or false.
	LookupConst func(name string) (float64, bool)
}

// Eval evaluates a parsed AST using env.
func Eval(n *node, env Env) (float64, error) {
	return evalNode(n, env)
}

// EvalString parses expr (funcs maps function names to their declared arity)
// and evaluates it in one step.
func EvalString(expr string, funcs map[string]int, env Env) (float64, error) {
	ast, err := Parse(expr, funcs)
	if err != nil {
		return 0, err
	}
	return evalNode(ast, env)
}

func evalNode(n *node, env Env) (float64, error) {
	switch n.kind {
	case nNumber:
		return n.val, nil
	case nIdent:
		if v, ok := env.LookupConst(n.name); ok {
			return v, nil
		}
		return 0, fmt.Errorf("unknown identifier %q at position %d", n.name, n.pos)
	case nUnary:
		v, err := evalNode(n.lhs, env)
		if err != nil {
			return 0, err
		}
		return -v, nil
	case nBinary:
		lhs, err := evalNode(n.lhs, env)
		if err != nil {
			return 0, err
		}
		rhs, err := evalNode(n.rhs, env)
		if err != nil {
			return 0, err
		}
		return applyBinary(n.name, lhs, rhs, n.pos)
	case nCall:
		args := make([]float64, 0, len(n.args))
		for _, a := range n.args {
			v, err := evalNode(a, env)
			if err != nil {
				return 0, err
			}
			args = append(args, v)
		}
		fn, ok := env.LookupFunc(n.name, len(args))
		if !ok {
			return 0, fmt.Errorf("unknown function %q with %d argument(s) at position %d", n.name, len(args), n.pos)
		}
		return fn(args)
	}
	return 0, fmt.Errorf("internal error: unknown node kind")
}

func applyBinary(op string, a, b float64, pos int) (float64, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, fmt.Errorf("division by zero at position %d", pos)
		}
		return a / b, nil
	case "%":
		if b == 0 {
			return 0, fmt.Errorf("modulo by zero at position %d", pos)
		}
		return math.Mod(a, b), nil
	case "^":
		r := math.Pow(a, b)
		if math.IsNaN(r) {
			return 0, fmt.Errorf("invalid power operation at position %d", pos)
		}
		return r, nil
	}
	return 0, fmt.Errorf("unknown operator %q at position %d", op, pos)
}
