package developer

import (
	"math"
	"testing"
)

func TestFormat(t *testing.T) {
	cases := []struct {
		raw    uint64
		ws     WordSize
		signed bool
		want   Formatted
	}{
		{0xFF, Bits8, false, Formatted{Hex: "FF", Dec: "255", Oct: "377", Bin: "11111111"}},
		{0xFF, Bits8, true, Formatted{Hex: "FF", Dec: "-1", Oct: "377", Bin: "11111111"}},
		{0x80, Bits8, true, Formatted{Hex: "80", Dec: "-128", Oct: "200", Bin: "10000000"}},
		{0x7F, Bits8, true, Formatted{Hex: "7F", Dec: "127", Oct: "177", Bin: "01111111"}},
		{0xFFFF, Bits16, false, Formatted{Hex: "FFFF", Dec: "65535", Oct: "177777", Bin: "1111111111111111"}},
		{0x1234, Bits16, true, Formatted{Hex: "1234", Dec: "4660", Oct: "011064", Bin: "0001001000110100"}},
	}
	for _, c := range cases {
		got := Format(c.raw, c.ws, c.signed)
		if got != c.want {
			t.Errorf("Format(%#x, %d, %v) = %+v, want %+v", c.raw, c.ws, c.signed, got, c.want)
		}
	}
}

func TestFormat64BitPrecision(t *testing.T) {
	unsigned := Format(^uint64(0), Bits64, false)
	if unsigned.Dec != "18446744073709551615" {
		t.Errorf("unsigned max dec = %s", unsigned.Dec)
	}
	if len(unsigned.Bin) != 64 || changedBitCount(unsigned.Bin) != 64 {
		t.Errorf("unsigned max bin = %s", unsigned.Bin)
	}

	if got := Format(^uint64(0), Bits64, true).Dec; got != "-1" {
		t.Errorf("signed -1 dec = %s", got)
	}
	if got := Format(1<<63, Bits64, true).Dec; got != "-9223372036854775808" {
		t.Errorf("signed min dec = %s", got)
	}
}

func changedBitCount(s string) int {
	n := 0
	for _, c := range s {
		if c == '1' {
			n++
		}
	}
	return n
}

func TestFromString(t *testing.T) {
	cases := []struct {
		s      string
		base   int
		ws     WordSize
		signed bool
		want   uint64
	}{
		{"FF", BaseHex, Bits8, false, 0xFF},
		{"ff", BaseHex, Bits8, false, 0xFF},
		{"255", BaseDec, Bits8, false, 0xFF},
		{"377", BaseOct, Bits8, false, 0xFF},
		{"11111111", BaseBin, Bits8, false, 0xFF},
		{"-1", BaseDec, Bits8, true, 0xFF},
		{"-128", BaseDec, Bits8, true, 0x80},
		{"100", BaseDec, Bits8, false, 0x64},
	}
	for _, c := range cases {
		got, err := FromString(c.s, c.base, c.ws, c.signed)
		if err != nil {
			t.Errorf("FromString(%q) error: %v", c.s, err)
			continue
		}
		if got != c.want {
			t.Errorf("FromString(%q) = %#x, want %#x", c.s, got, c.want)
		}
	}
}

func TestFromStringErrors(t *testing.T) {
	cases := []struct {
		s    string
		base int
	}{
		{"GG", BaseHex},
		{"2", BaseBin},
		{"8", BaseOct},
		{"", BaseDec},
		{"-5", BaseDec}, // unsigned mode rejects the sign
	}
	for _, c := range cases {
		if _, err := FromString(c.s, c.base, Bits8, false); err == nil {
			t.Errorf("FromString(%q, base %d) expected error", c.s, c.base)
		}
	}
}

func TestBinary(t *testing.T) {
	cases := []struct {
		op     string
		a, b   uint64
		ws     WordSize
		signed bool
		want   uint64
	}{
		{OpAnd, 0xF0, 0x3C, Bits8, false, 0x30},
		{OpOr, 0xF0, 0x3C, Bits8, false, 0xFC},
		{OpXor, 0xF0, 0x3C, Bits8, false, 0xCC},
		{OpNand, 0xF0, 0x3C, Bits8, false, 0xCF},
		{OpNor, 0xF0, 0x3C, Bits8, false, 0x03},
		{OpAdd, 0xFE, 0x03, Bits8, false, 0x01}, // overflow wraps
		{OpSub, 0x01, 0x02, Bits8, false, 0xFF},
		{OpMul, 0x10, 0x10, Bits8, false, 0x00}, // 256 wraps to 0 unsigned
		{OpMul, 0x10, 0x10, Bits8, true, 0x00},  // 16*16=256 wraps in 8-bit
		{OpDiv, 0x64, 0x07, Bits8, false, 0x0E}, // 100/7 = 14
		{OpDiv, 0x80, 0xFF, Bits8, true, 0x80},  // -128 / -1 wraps
		{OpDiv, 0xF9, 0x02, Bits8, true, 0xFD},  // -7 / 2 = -3
		{OpMod, 0x64, 0x07, Bits8, false, 0x02},
		{OpMod, 0xF9, 0x02, Bits8, true, 0xFF}, // -7 % 2 = -1
		{OpMod, 0x80, 0xFF, Bits8, true, 0},    // min % -1 = 0
		{OpMul, 0x80, 0x02, Bits8, true, 0},    // signed overflow wraps
		{OpShl, 0x01, 4, Bits8, false, 0x10},
		{OpShl, 0x80, 1, Bits8, false, 0x00}, // shifted out
		{OpShr, 0x80, 4, Bits8, false, 0x08},
		{OpSar, 0x80, 4, Bits8, true, 0xF8},  // arithmetic shift keeps sign
		{OpSar, 0x80, 4, Bits8, false, 0x08}, // logical shift
		{OpRol, 0x81, 1, Bits8, false, 0x03},
		{OpRor, 0x81, 1, Bits8, false, 0xC0},
		{OpAdd, 1, 1, Bits64, true, 2},
	}
	for _, c := range cases {
		got, err := Binary(c.op, c.a, c.b, c.ws, c.signed)
		if err != nil {
			t.Errorf("Binary(%s) error: %v", c.op, err)
			continue
		}
		if got != c.want {
			t.Errorf("Binary(%s, %#x, %#x, %d, %v) = %#x, want %#x", c.op, c.a, c.b, c.ws, c.signed, got, c.want)
		}
	}
}

func TestBinary64BitPrecision(t *testing.T) {
	max := ^uint64(0)
	cases := []struct {
		op     string
		a, b   uint64
		signed bool
		want   uint64
	}{
		{OpMul, max, max, false, 1}, // (2^64-1)^2 mod 2^64
		{OpMul, max, max, true, 1},  // (-1)*(-1)
		{OpMul, 1 << 63, 2, false, 0},
		{OpMul, 1 << 63, 2, true, 0},
		{OpDiv, max, 2, false, 0x7FFFFFFFFFFFFFFF},
		{OpDiv, max, 2, true, 0},                                   // -1 / 2 truncates toward zero
		{OpDiv, 0x8000000000000000, max, true, 0x8000000000000000}, // min / -1 wraps
		{OpMod, max, 10, false, 5},
		{OpMod, max, 10, true, max}, // -1 % 10 = -1
		{OpMod, 0x8000000000000000, max, true, 0},
		{OpSar, 0x8000000000000000, 4, true, 0xF800000000000000},
		{OpSar, max, 4, true, max},
		{OpSar, max, 4, false, 0x0FFFFFFFFFFFFFFF},
		{OpSar, 0x8000000000000000, 64, true, max},
		{OpSar, 0x8000000000000000, 64, false, 0},
	}
	for _, c := range cases {
		got, err := Binary(c.op, c.a, c.b, Bits64, c.signed)
		if err != nil {
			t.Errorf("Binary(%s) error: %v", c.op, err)
			continue
		}
		if got != c.want {
			t.Errorf("Binary(%s, %#x, %#x, 64, %v) = %#x, want %#x", c.op, c.a, c.b, c.signed, got, c.want)
		}
	}
}

func TestInterpretSigned64(t *testing.T) {
	if _, v := Interpret(0xFF, Bits8, true); v != -1 {
		t.Errorf("8-bit -1 = %v", v)
	}
	if _, v := Interpret(0x8000000000000000, Bits64, true); v != math.MinInt64 {
		t.Errorf("64-bit min = %v", v)
	}
	if raw, v := Interpret(0xFFFFFFFFFFFFFFFF, Bits64, true); raw != ^uint64(0) || v != -1 {
		t.Errorf("64-bit -1 = %#x, %v", raw, v)
	}
}

func TestBinaryErrors(t *testing.T) {
	if _, err := Binary(OpDiv, 1, 0, Bits8, false); err == nil {
		t.Error("div by zero: expected error")
	}
	if _, err := Binary("BOGUS", 1, 1, Bits8, false); err == nil {
		t.Error("unknown op: expected error")
	}
}

func TestNotNegate(t *testing.T) {
	if got, _ := Not(0x0F, Bits8); got != 0xF0 {
		t.Errorf("Not = %#x, want F0", got)
	}
	if got, _ := Negate(1, Bits8); got != 0xFF {
		t.Errorf("Negate(1) = %#x, want FF", got)
	}
	if got, _ := Negate(0, Bits8); got != 0 {
		t.Errorf("Negate(0) = %#x, want 0", got)
	}
}

func TestSignedArithmetic(t *testing.T) {
	// -1 + 1 = 0 in signed 8-bit.
	got, err := Binary(OpAdd, 0xFF, 0x01, Bits8, true)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Errorf("got %#x, want 0", got)
	}
	// -128 - 1 wraps to 127.
	got, err = Binary(OpSub, 0x80, 0x01, Bits8, true)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0x7F {
		t.Errorf("got %#x, want 7F", got)
	}
}

func TestBitCountAndRange(t *testing.T) {
	if BitCount(0b1011) != 3 {
		t.Error("BitCount")
	}
	if MaxValue(Bits8, false) != 255 || MaxValue(Bits8, true) != 127 {
		t.Error("MaxValue")
	}
	if MinValue(Bits8, true) != 0x80 || MinValue(Bits8, false) != 0 {
		t.Error("MinValue")
	}
	if _, err := ParseWordSize(12); err == nil {
		t.Error("ParseWordSize(12) expected error")
	}
}
