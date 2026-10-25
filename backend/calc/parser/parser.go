// Package parser implements a small Pratt-style expression parser used by the
// scientific engine. It supports numbers, identifiers (constants and function
// calls), the four basic arithmetic operators, ^ (right-associative), unary
// plus/minus, postfix factorial (!), parentheses and implicit multiplication
// such as "2(3+4)", "2pi" or "3sin(0.5)".
package parser

import (
	"fmt"
	"math"
)

// node is an AST node.
type node struct {
	kind     nodeKind
	val      float64 // for a number node
	name     string  // for an ident node (function or constant)
	args     []*node // for a function call
	lhs, rhs *node   // for unary/binary operators
	pos      int     // byte offset in the source, for error messages
}

type nodeKind int

const (
	nNumber nodeKind = iota
	nIdent
	nUnary  // lhs holds the operand; name is the operator ("-")
	nBinary // lhs, rhs; name is the operator
	nCall   // name is the function; args holds arguments
)

// binary operator precedence.
const (
	precAdd  = 1
	precMul  = 2
	precPow  = 3
	precBang = 4
)

type parser struct {
	toks []token
	pos  int
	// funcs returns the arity of each supported function; nil means unknown.
	funcs map[string]int
}

// Parse builds an AST for expr. funcs maps function names to their arity;
// anything not in funcs is treated as a constant (validated later during eval).
func Parse(expr string, funcs map[string]int) (*node, error) {
	toks, err := lex(expr)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks, funcs: funcs}
	n, err := p.parseExpr(precAdd)
	if err != nil {
		return nil, err
	}
	if p.peek().typ != tokEOF {
		return nil, fmt.Errorf("unexpected %s at position %d", p.peek(), p.peek().pos)
	}
	return n, nil
}

func (p *parser) peek() token { return p.toks[p.pos] }

func (p *parser) next() token {
	t := p.toks[p.pos]
	if t.typ != tokEOF {
		p.pos++
	}
	return t
}

// parseExpr parses operators at or above minPrec.
func (p *parser) parseExpr(minPrec int) (*node, error) {
	lhs, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		op := p.peek()
		if op.typ == tokBang {
			if precBang < minPrec {
				return lhs, nil
			}
			p.next()
			lhs = &node{kind: nCall, name: "fact", args: []*node{lhs}, pos: op.pos}
			continue
		}
		if op.typ == tokOp {
			prec, ok := opPrec(op.val)
			if !ok || prec < minPrec {
				return lhs, nil
			}
			p.next()
			// ^ is right-associative, everything else is left-associative.
			nextMin := prec + 1
			if op.val == "^" {
				nextMin = prec
			}
			rhs, err := p.parseExpr(nextMin)
			if err != nil {
				return nil, err
			}
			lhs = &node{kind: nBinary, name: op.val, lhs: lhs, rhs: rhs, pos: op.pos}
			continue
		}
		// Implicit multiplication: a primary directly after a value,
		// e.g. "2(3+4)", "2pi", "2√9". Treat with multiplication precedence.
		if op.typ == tokNumber || op.typ == tokIdent || op.typ == tokLparen {
			if precMul < minPrec {
				return lhs, nil
			}
			rhs, err := p.parseExpr(precMul + 1)
			if err != nil {
				return nil, err
			}
			lhs = &node{kind: nBinary, name: "*", lhs: lhs, rhs: rhs, pos: op.pos}
			continue
		}
		return lhs, nil
	}
}

func opPrec(op string) (int, bool) {
	switch op {
	case "+", "-":
		return precAdd, true
	case "*", "/", "%":
		return precMul, true
	case "^":
		return precPow, true
	}
	return 0, false
}

func (p *parser) parseUnary() (*node, error) {
	t := p.peek()
	if t.typ == tokOp && (t.val == "-" || t.val == "+") {
		p.next()
		// The operand parses at power precedence so that -2^2 == -(2^2)
		// while (-2)^2 is still expressible with explicit parens.
		operand, err := p.parseExpr(precPow)
		if err != nil {
			return nil, err
		}
		if t.val == "-" {
			return &node{kind: nUnary, name: "-", lhs: operand, pos: t.pos}, nil
		}
		return operand, nil
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (*node, error) {
	t := p.next()
	switch t.typ {
	case tokNumber:
		v, err := parseNumber(t.val)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q at position %d", t.val, t.pos)
		}
		return &node{kind: nNumber, val: v, pos: t.pos}, nil
	case tokIdent:
		if arity, ok := p.funcs[t.val]; ok {
			// Function call with or without parentheses.
			if p.peek().typ == tokLparen {
				p.next()
				args, err := p.parseArgs()
				if err != nil {
					return nil, err
				}
				if len(args) != arity {
					return nil, fmt.Errorf("%s expects %d argument(s) but got %d at position %d", t.val, arity, len(args), t.pos)
				}
				return &node{kind: nCall, name: t.val, args: args, pos: t.pos}, nil
			}
			if arity == 0 {
				return &node{kind: nCall, name: t.val, pos: t.pos}, nil
			}
			// Function without parens applies to the next unary expression:
			// sin 0.5, sqrt 9.
			arg, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			return &node{kind: nCall, name: t.val, args: []*node{arg}, pos: t.pos}, nil
		}
		return &node{kind: nIdent, name: t.val, pos: t.pos}, nil
	case tokLparen:
		n, err := p.parseExpr(precAdd)
		if err != nil {
			return nil, err
		}
		if p.next().typ != tokRparen {
			return nil, fmt.Errorf("missing closing parenthesis at position %d", t.pos)
		}
		return n, nil
	default:
		return nil, fmt.Errorf("unexpected %s at position %d", t, t.pos)
	}
}

func (p *parser) parseArgs() ([]*node, error) {
	var args []*node
	if p.peek().typ == tokRparen {
		p.next()
		return args, nil
	}
	for {
		n, err := p.parseExpr(precAdd)
		if err != nil {
			return nil, err
		}
		args = append(args, n)
		t := p.next()
		if t.typ == tokRparen {
			return args, nil
		}
		if t.typ != tokComma {
			return nil, fmt.Errorf("expected ',' or ')' at position %d", t.pos)
		}
	}
}

func parseNumber(s string) (float64, error) {
	// Reject "1.2.3" style strings.
	dots := 0
	for _, c := range s {
		if c == '.' {
			dots++
		}
	}
	if dots > 1 {
		return 0, fmt.Errorf("too many decimal points")
	}
	var v float64
	if _, err := fmt.Sscanf(s, "%g", &v); err != nil {
		return 0, err
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("not a finite number")
	}
	return v, nil
}
