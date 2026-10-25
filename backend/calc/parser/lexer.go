package parser

import (
	"fmt"
	"strings"
)

// tokenType classifies a lexical token.
type tokenType int

const (
	tokNumber tokenType = iota
	tokIdent
	tokOp // + - * / ^ %
	tokLparen
	tokRparen
	tokComma
	tokBang // postfix factorial
	tokEOF
)

type token struct {
	typ tokenType
	val string
	pos int // byte offset in the source expression
}

func (t token) String() string {
	switch t.typ {
	case tokEOF:
		return "end of expression"
	default:
		return fmt.Sprintf("%q", t.val)
	}
}

func isASCIILetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isIdentByte(c byte) bool {
	return isASCIILetter(c) || c == '_' || c >= '0' && c <= '9'
}

// symbols maps multi-byte input glyphs to their ASCII replacements so
// scientific expressions can use native typography: × ÷ − √ π τ 𝑒 ℯ.
var symbols = []struct{ glyph, replacement string }{
	{"×", "*"},
	{"÷", "/"},
	{"−", "-"},
	{"√", "sqrt"},
	{"π", "pi"},
	{"τ", "tau"},
	{"𝑒", "e"},
	{"ℯ", "e"},
}

// lex splits an expression into tokens. Identifiers are ASCII-only; Unicode
// math glyphs are mapped to their ASCII equivalents.
func lex(src string) ([]token, error) {
	var toks []token
	i := 0
	for i < len(src) {
		c := src[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c >= '0' && c <= '9' || (c == '.' && i+1 < len(src) && src[i+1] >= '0' && src[i+1] <= '9'):
			start := i
			i++
			for i < len(src) && (src[i] >= '0' && src[i] <= '9' || src[i] == '.') {
				i++
			}
			toks = append(toks, token{tokNumber, src[start:i], start})
		case isASCIILetter(c) || c == '_':
			start := i
			for i < len(src) && isIdentByte(src[i]) {
				i++
			}
			toks = append(toks, token{tokIdent, strings.ToLower(src[start:i]), start})
		case c == '(':
			toks = append(toks, token{tokLparen, "(", i})
			i++
		case c == ')':
			toks = append(toks, token{tokRparen, ")", i})
			i++
		case c == ',':
			toks = append(toks, token{tokComma, ",", i})
			i++
		case c == '!':
			toks = append(toks, token{tokBang, "!", i})
			i++
		case c == '+' || c == '-' || c == '*' || c == '/' || c == '^' || c == '%':
			if c == '*' && i+1 < len(src) && src[i+1] == '*' {
				toks = append(toks, token{tokOp, "^", i})
				i += 2
				continue
			}
			toks = append(toks, token{tokOp, string(c), i})
			i++
		default:
			matched := false
			for _, sym := range symbols {
				if strings.HasPrefix(src[i:], sym.glyph) {
					val := sym.replacement
					if val == "*" || val == "/" || val == "-" {
						toks = append(toks, token{tokOp, val, i})
					} else {
						toks = append(toks, token{tokIdent, val, i})
					}
					i += len(sym.glyph)
					matched = true
					break
				}
			}
			if !matched {
				return nil, fmt.Errorf("unexpected character %q at position %d", string(rune(c)), i)
			}
		}
	}
	toks = append(toks, token{tokEOF, "", i})
	return toks, nil
}
