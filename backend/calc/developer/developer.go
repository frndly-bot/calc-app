// Package developer implements programmer-calculator arithmetic: fixed word
// sizes, signed/unsigned interpretation, bitwise operations, shifts, rotates
// and HEX/DEC/OCT/BIN formatting.
package developer

import (
	"fmt"
	"strconv"
	"strings"
)

// WordSize is the operand width in bits.
type WordSize int

const (
	Bits8  WordSize = 8
	Bits16 WordSize = 16
	Bits32 WordSize = 32
	Bits64 WordSize = 64
)

// Valid reports whether w is one of the supported word sizes.
func (w WordSize) Valid() bool {
	return w == 8 || w == 16 || w == 32 || w == 64
}

// Base names used in formatted output.
const (
	BaseHex = 16
	BaseDec = 10
	BaseOct = 8
	BaseBin = 2
)

func mask(ws WordSize) uint64 {
	if ws == 64 {
		return ^uint64(0)
	}
	return (uint64(1) << uint(ws)) - 1
}

func signBit(ws WordSize) uint64 {
	return uint64(1) << uint(ws-1)
}

// signedInt returns the two's-complement interpretation of a masked word,
// sign-extending narrower word sizes.
func signedInt(raw uint64, ws WordSize) int64 {
	raw &= mask(ws)
	if raw&signBit(ws) == 0 {
		return int64(raw)
	}
	return int64(raw | ^mask(ws))
}

// Interpret converts a masked word to the value it represents under the given
// signedness. The first result holds the masked bit pattern.
func Interpret(raw uint64, ws WordSize, signed bool) (uint64, float64) {
	raw &= mask(ws)
	if !signed {
		return raw, float64(raw)
	}
	return raw, float64(signedInt(raw, ws))
}

// FromString parses s in the given base into a word. When signed, a leading
// '-' parses as the two's complement bit pattern.
func FromString(s string, base int, ws WordSize, signed bool) (uint64, error) {
	if !ws.Valid() {
		return 0, fmt.Errorf("unsupported word size %d", int(ws))
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty value")
	}
	neg := false
	if signed && strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	if s == "" {
		return 0, fmt.Errorf("empty value")
	}
	var v uint64
	var err error
	switch base {
	case BaseHex:
		v, err = strconv.ParseUint(s, 16, 64)
	case BaseDec:
		v, err = strconv.ParseUint(s, 10, 64)
	case BaseOct:
		v, err = strconv.ParseUint(s, 8, 64)
	case BaseBin:
		v, err = strconv.ParseUint(s, 2, 64)
	default:
		return 0, fmt.Errorf("unsupported base %d", base)
	}
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q", baseName(base), s)
	}
	v &= mask(ws)
	if neg {
		if v == 0 {
			return 0, nil
		}
		v = (^v + 1) & mask(ws) // two's complement negate
	}
	return v, nil
}

func baseName(base int) string {
	switch base {
	case BaseHex:
		return "hexadecimal"
	case BaseDec:
		return "decimal"
	case BaseOct:
		return "octal"
	case BaseBin:
		return "binary"
	}
	return fmt.Sprintf("base-%d", base)
}

// Formatted holds one value rendered in all four bases.
type Formatted struct {
	Hex string
	Dec string
	Oct string
	Bin string
}

// Format renders raw (already masked) in all four bases. Signed DEC output
// uses two's-complement interpretation; HEX/OCT/BIN always show raw bits.
func Format(raw uint64, ws WordSize, signed bool) Formatted {
	raw &= mask(ws)
	hexDigits := int(ws) / 4
	dec := strconv.FormatUint(raw, 10)
	if signed {
		value := int64(raw)
		if ws < 64 && raw&signBit(ws) != 0 {
			value -= int64(1) << uint(ws)
		}
		dec = strconv.FormatInt(value, 10)
	}
	return Formatted{
		Hex: strings.ToUpper(fmt.Sprintf("%0*X", hexDigits, raw)),
		Dec: dec,
		Oct: fmt.Sprintf("%0*o", (int(ws)+2)/3, raw),
		Bin: fmt.Sprintf("%0*b", int(ws), raw),
	}
}

// Binary ops.
const (
	OpAnd  = "AND"
	OpOr   = "OR"
	OpXor  = "XOR"
	OpNand = "NAND"
	OpNor  = "NOR"
	OpAdd  = "ADD"
	OpSub  = "SUB"
	OpMul  = "MUL"
	OpDiv  = "DIV"
	OpMod  = "MOD"
	OpShl  = "SHL"
	OpShr  = "SHR"
	OpSar  = "SAR"
	OpRol  = "ROL"
	OpRor  = "ROR"
	OpNot  = "NOT"
	OpTwos = "NEG"
)

// Binary applies a binary operator. a and b must already be masked words.
// b is interpreted as unsigned for shifts/rotates (shift amounts are never
// negative) and per `signed` for arithmetic.
func Binary(op string, a, b uint64, ws WordSize, signed bool) (uint64, error) {
	if !ws.Valid() {
		return 0, fmt.Errorf("unsupported word size %d", int(ws))
	}
	m := mask(ws)
	a &= m
	b &= m
	switch op {
	case OpAnd:
		return a & b, nil
	case OpOr:
		return a | b, nil
	case OpXor:
		return a ^ b, nil
	case OpNand:
		return ^(a & b) & m, nil
	case OpNor:
		return ^(a | b) & m, nil
	case OpAdd:
		return (a + b) & m, nil
	case OpSub:
		return (a - b) & m, nil
	case OpMul:
		if signed {
			return uint64(signedInt(a, ws)*signedInt(b, ws)) & m, nil
		}
		return (a * b) & m, nil
	case OpDiv:
		if signed {
			bi := signedInt(b, ws)
			if bi == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return uint64(signedInt(a, ws)/bi) & m, nil
		}
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return a / b, nil
	case OpMod:
		if signed {
			bi := signedInt(b, ws)
			if bi == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			return uint64(signedInt(a, ws)%bi) & m, nil
		}
		if b == 0 {
			return 0, fmt.Errorf("modulo by zero")
		}
		return a % b, nil
	case OpShl:
		if b >= uint64(ws) {
			return 0, nil
		}
		return (a << uint(b)) & m, nil
	case OpShr:
		if b >= uint64(ws) {
			return 0, nil
		}
		return a >> uint(b), nil
	case OpSar:
		if b >= uint64(ws) {
			if signed && signedInt(a, ws) < 0 {
				return m, nil
			}
			return 0, nil
		}
		if signed {
			return uint64(signedInt(a, ws)>>uint(b)) & m, nil
		}
		return a >> uint(b), nil
	case OpRol:
		return rotate(a, b, ws, true), nil
	case OpRor:
		return rotate(a, b, ws, false), nil
	}
	return 0, fmt.Errorf("unknown operator %q", op)
}

func rotate(a, n uint64, ws WordSize, left bool) uint64 {
	w := uint(ws)
	m := mask(ws)
	a &= m
	if w == 64 {
		n %= 64
	} else {
		n %= uint64(w)
	}
	if n == 0 {
		return a
	}
	if left {
		return ((a << uint(n)) | (a >> (w - uint(n)))) & m
	}
	return ((a >> uint(n)) | (a << (w - uint(n)))) & m
}

// Not returns the bitwise complement of a within the word size.
func Not(a uint64, ws WordSize) (uint64, error) {
	if !ws.Valid() {
		return 0, fmt.Errorf("unsupported word size %d", int(ws))
	}
	return ^a & mask(ws), nil
}

// Negate returns the two's complement of a within the word size.
func Negate(a uint64, ws WordSize) (uint64, error) {
	if !ws.Valid() {
		return 0, fmt.Errorf("unsupported word size %d", int(ws))
	}
	if a&mask(ws) == 0 {
		return 0, nil
	}
	return (^a + 1) & mask(ws), nil
}

// ParseWordSize converts 8/16/32/64 to a WordSize.
func ParseWordSize(bits int) (WordSize, error) {
	ws := WordSize(bits)
	if !ws.Valid() {
		return 0, fmt.Errorf("word size must be 8, 16, 32 or 64, got %d", bits)
	}
	return ws, nil
}

// BitCount returns the number of set bits.
func BitCount(a uint64) int {
	return bitsOnes(a)
}

func bitsOnes(x uint64) int {
	c := 0
	for x != 0 {
		x &= x - 1
		c++
	}
	return c
}

// MaxValue returns the largest unsigned value for the word size, or the
// largest signed value when signed is true.
func MaxValue(ws WordSize, signed bool) uint64 {
	if signed {
		return mask(ws) >> 1
	}
	return mask(ws)
}

// MinValue returns the most negative signed value for the word size
// (0 when unsigned).
func MinValue(ws WordSize, signed bool) uint64 {
	if signed {
		return signBit(ws)
	}
	return 0
}
