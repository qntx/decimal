package uint128

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"math"
	"math/big"
	"net"
	"testing"
)

func randUint128() Uint128 {
	randBuf := make([]byte, 16)
	rand.Read(randBuf)
	return NewFromBytes(randBuf)
}

// testNonArithmeticMethods tests non-arithmetic operations like conversions and comparisons
func TestNonArithmeticMethods(t *testing.T) {
	const iterations = 1000
	tests := []struct {
		name     string
		modifier func(Uint128) Uint128
	}{
		{"Unmodified", func(x Uint128) Uint128 { return x }},
		{"RightShift64", func(x Uint128) Uint128 { return x.Rsh(64) }},
		{"LeftShift64", func(x Uint128) Uint128 { return x.Lsh(64) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := range iterations {
				x, y := randUint128(), randUint128()
				if i%3 == 0 && tt.name == "RightShift64" {
					x = tt.modifier(x)
				} else if i%7 == 0 && tt.name == "LeftShift64" {
					x = tt.modifier(x)
				} else if tt.name == "Unmodified" {
					x = tt.modifier(x)
				}

				// Test Big/NewFromBig roundtrip
				assertBigRoundtrip(t, x)
				// Test PutBytes/NewFromBytes roundtrip
				assertBytesRoundtrip(t, x)
				// Test equality checks
				assertSelfEquality(t, x)
				// Test comparisons
				assertComparisons(t, x, y)
			}
		})
	}
}

// TestFromBigErrors tests error cases for NewFromBig
func TestFromBigErrors(t *testing.T) {
	tests := []struct {
		name      string
		input     *big.Int
		wantError error
	}{
		{
			name:      "NegativeValue",
			input:     big.NewInt(-1),
			wantError: ErrNegativeValue,
		},
		{
			name:      "ValueOverflow",
			input:     new(big.Int).Lsh(big.NewInt(1), 129),
			wantError: ErrValueOverflow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFromBigInt(tt.input)
			if err != tt.wantError {
				t.Errorf("NewFromBigInt(%v) error = %v, want %v", tt.input, err, tt.wantError)
			}
		})
	}
}

// assertBigRoundtrip verifies that Big and NewFromBig are inverses
func assertBigRoundtrip(t *testing.T, x Uint128) {
	t.Helper()
	u, err := NewFromBigInt(x.Big())
	if err != nil {
		t.Errorf("NewFromBigInt(%v) failed: %v", x, err)
		return
	}
	if u != x {
		t.Errorf("NewFromBig(%v) = %v, want %v", x, u, x)
	}
}

// assertBytesRoundtrip verifies that PutBytes and NewFromBytes are inverses
func assertBytesRoundtrip(t *testing.T, x Uint128) {
	t.Helper()
	b := make([]byte, 16)
	x.PutBytes(b)
	if got := NewFromBytes(b); got != x {
		t.Errorf("NewFromBytes(PutBytes(%v)) = %v, want %v", x, got, x)
	}
}

// assertSelfEquality verifies that a value equals itself
func assertSelfEquality(t *testing.T, x Uint128) {
	t.Helper()
	if !x.Equals(x) {
		t.Errorf("%v does not equal itself", x)
	}
	if !NewFromUint64(x.lo).Equals64(x.lo) {
		t.Errorf("%v (lo) does not equal itself", x.lo)
	}
}

// assertComparisons verifies comparison operations
func assertComparisons(t *testing.T, x, y Uint128) {
	t.Helper()
	// Compare with another Uint128
	if got := x.Cmp(y); got != x.Big().Cmp(y.Big()) {
		t.Errorf("Cmp(%v, %v) = %v, want %v", x, y, got, x.Big().Cmp(y.Big()))
	}
	if x.Cmp(x) != 0 {
		t.Errorf("%v does not equal itself in Cmp", x)
	}

	// Compare with uint64
	if got := x.Cmp64(y.lo); got != x.Big().Cmp(NewFromUint64(y.lo).Big()) {
		t.Errorf("Cmp64(%v, %v) = %v, want %v", x, y.lo, got, x.Big().Cmp(NewFromUint64(y.lo).Big()))
	}
	if NewFromUint64(x.lo).Cmp64(x.lo) != 0 {
		t.Errorf("%v (lo) does not equal itself in Cmp64", x.lo)
	}
}

func TestArithmetic(t *testing.T) {
	// compare Uint128 arithmetic methods to their math/big equivalents, using
	// random values
	randBuf := make([]byte, 17)
	randUint128 := func() Uint128 {
		rand.Read(randBuf)
		var Lo, Hi uint64
		if randBuf[16]&1 != 0 {
			Lo = binary.LittleEndian.Uint64(randBuf[:8])
		}
		if randBuf[16]&2 != 0 {
			Hi = binary.LittleEndian.Uint64(randBuf[8:])
		}
		return New(Lo, Hi)
	}
	mod128 := func(i *big.Int) *big.Int {
		// wraparound semantics
		if i.Sign() == -1 {
			i = i.Add(new(big.Int).Lsh(big.NewInt(1), 128), i)
		}
		_, rem := i.QuoRem(i, new(big.Int).Lsh(big.NewInt(1), 128), new(big.Int))
		return rem
	}
	checkBinOpX := func(x Uint128, op string, y Uint128, fn func(x, y Uint128) Uint128, fnb func(z, x, y *big.Int) *big.Int) {
		t.Helper()
		rb := fnb(new(big.Int), x.Big(), y.Big())
		defer func() {
			if r := recover(); r != nil {
				if rb.BitLen() <= 128 && rb.Sign() >= 0 {
					t.Fatalf("mismatch: %v%v%v should not panic, %v", x, op, y, rb)
				}
			} else if rb.BitLen() > 128 || rb.Sign() < 0 {
				t.Fatalf("mismatch: %v%v%v should panic, %v", x, op, y, rb)
			}
		}()
		r := fn(x, y)
		if r.Big().Cmp(rb) != 0 {
			t.Fatalf("mismatch: %v%v%v should equal %v, got %v", x, op, y, rb, r)
		}
	}
	checkBinOp := func(x Uint128, op string, y Uint128, fn func(x, y Uint128) Uint128, fnb func(z, x, y *big.Int) *big.Int) {
		t.Helper()
		r := fn(x, y)
		rb := mod128(fnb(new(big.Int), x.Big(), y.Big()))
		if r.Big().Cmp(rb) != 0 {
			t.Fatalf("mismatch: %v%v%v should equal %v, got %v", x, op, y, rb, r)
		}
	}
	checkShiftOp := func(x Uint128, op string, n uint, fn func(x Uint128, n uint) Uint128, fnb func(z, x *big.Int, n uint) *big.Int) {
		t.Helper()
		r := fn(x, n)
		rb := mod128(fnb(new(big.Int), x.Big(), n))
		if r.Big().Cmp(rb) != 0 {
			t.Fatalf("mismatch: %v%v%v should equal %v, got %v", x, op, n, rb, r)
		}
	}
	checkBinOp64X := func(x Uint128, op string, y uint64, fn func(x Uint128, y uint64) Uint128, fnb func(z, x, y *big.Int) *big.Int) {
		t.Helper()
		xb, yb := x.Big(), NewFromUint64(y).Big()
		rb := fnb(new(big.Int), xb, yb)
		defer func() {
			if r := recover(); r != nil {
				if rb.BitLen() <= 128 && rb.Sign() >= 0 {
					t.Fatalf("mismatch: %v%v%v should not panic, %v", x, op, y, rb)
				}
			} else if rb.BitLen() > 128 || rb.Sign() < 0 {
				t.Fatalf("mismatch: %v%v%v should panic, %v", x, op, y, rb)
			}
		}()
		r := fn(x, y)
		if r.Big().Cmp(rb) != 0 {
			t.Fatalf("mismatch: %v%v%v should equal %v, got %v", x, op, y, rb, r)
		}
	}
	checkBinOp64 := func(x Uint128, op string, y uint64, fn func(x Uint128, y uint64) Uint128, fnb func(z, x, y *big.Int) *big.Int) {
		t.Helper()
		xb, yb := x.Big(), NewFromUint64(y).Big()
		r := fn(x, y)
		rb := mod128(fnb(new(big.Int), xb, yb))
		if r.Big().Cmp(rb) != 0 {
			t.Fatalf("mismatch: %v%v%v should equal %v, got %v", x, op, y, rb, r)
		}
	}
	for i := 0; i < 1000; i++ {
		x, y, z := randUint128(), randUint128(), uint(randUint128().lo&0xFF)
		checkBinOpX(x, "[+]", y, Uint128.MustAdd, (*big.Int).Add)
		checkBinOpX(x, "[-]", y, Uint128.MustSub, (*big.Int).Sub)
		checkBinOpX(x, "[*]", y, Uint128.MustMul, (*big.Int).Mul)
		checkBinOp(x, "+", y, Uint128.AddWrap, (*big.Int).Add)
		checkBinOp(x, "-", y, Uint128.SubWrap, (*big.Int).Sub)
		checkBinOp(x, "*", y, Uint128.MulWrap, (*big.Int).Mul)
		if !y.IsZero() {
			checkBinOp(x, "/", y, Uint128.MustDiv, (*big.Int).Div)
			checkBinOp(x, "%", y, Uint128.MustMod, (*big.Int).Mod)
		}
		checkBinOp(x, "&", y, Uint128.And, (*big.Int).And)
		checkBinOp(x, "|", y, Uint128.Or, (*big.Int).Or)
		checkBinOp(x, "^", y, Uint128.Xor, (*big.Int).Xor)
		checkShiftOp(x, "<<", z, Uint128.Lsh, (*big.Int).Lsh)
		checkShiftOp(x, ">>", z, Uint128.Rsh, (*big.Int).Rsh)

		// check 64-bit variants
		y64 := y.lo
		checkBinOp64X(x, "[+]", y64, Uint128.MustAdd64, (*big.Int).Add)
		checkBinOp64X(x, "[-]", y64, Uint128.MustSub64, (*big.Int).Sub)
		checkBinOp64X(x, "[*]", y64, Uint128.MustMul64, (*big.Int).Mul)
		checkBinOp64(x, "+", y64, Uint128.AddWrap64, (*big.Int).Add)
		checkBinOp64(x, "-", y64, Uint128.SubWrap64, (*big.Int).Sub)
		checkBinOp64(x, "*", y64, Uint128.MulWrap64, (*big.Int).Mul)
		if y64 != 0 {
			checkBinOp64(x, "/", y64, Uint128.Div64, (*big.Int).Div)
			modfn := func(x Uint128, y uint64) Uint128 {
				return NewFromUint64(x.Mod64(y))
			}
			checkBinOp64(x, "%", y64, modfn, (*big.Int).Mod)
		}
		checkBinOp64(x, "&", y64, Uint128.And64, (*big.Int).And)
		checkBinOp64(x, "|", y64, Uint128.Or64, (*big.Int).Or)
		checkBinOp64(x, "^", y64, Uint128.Xor64, (*big.Int).Xor)
	}
}

func TestOverflowAndUnderflow(t *testing.T) {
	x := Max
	y := New(10, 10)
	z := NewFromUint64(10)

	var err error

	// Test Add overflow
	_, err = x.Add(y) // max.Add(New(10,10))
	if err != ErrOverflow {
		t.Errorf("expected ErrOverflow, got %v", err)
	}
	_, err = x.Add64(10) // max.Add64(10)
	if err != ErrOverflow {
		t.Errorf("expected ErrOverflow, got %v", err)
	}

	// Test Sub underflow
	_, err = y.Sub(x) // New(10,10).Sub(max)
	if err != ErrUnderflow {
		t.Errorf("expected ErrUnderflow, got %v", err)
	}
	_, err = z.Sub64(math.MaxUint64) // NewFrom64(10).Sub64(MaxUint64) - Note: math.MaxInt64 might not be the correct value to trigger underflow here depending on z's value.
	// Using math.MaxUint64 to ensure underflow if z.Lo is small.
	if err != ErrUnderflow {
		t.Errorf("expected ErrUnderflow, got %v", err)
	}

	// Test Mul overflow
	_, err = x.Mul(y) // max.Mul(New(10,10))
	if err != ErrOverflow {
		t.Errorf("expected ErrOverflow, got %v", err)
	}
	_, err = New(0, 10).Mul(New(0, 10)) // New(0,10).Mul(New(0,10))
	if err != ErrOverflow {
		t.Errorf("expected ErrOverflow, got %v", err)
	}
	// This specific case might not overflow if implemented correctly as 2^64 * 2^64 = 2^128 which is Max + 1, so it should wrap or error.
	// However, the original test expected an overflow for New(0, 1).Mul(New(0, 1)), which means multiplying 2^64 by 2^64.
	// A Uint128 can hold up to 2^128 - 1.  (2^64) * (2^64) = 2^128. This value is exactly one greater than Max.
	// So, this should indeed result in an overflow.
	_, err = New(0, 1).Mul(New(0, 1))
	if err != ErrOverflow {
		t.Errorf("expected ErrOverflow for New(0,1).Mul(New(0,1)), got %v", err)
	}
	_, err = x.Mul64(math.MaxUint64) // max.Mul64(MaxUint64)
	if err != ErrOverflow {
		t.Errorf("expected ErrOverflow, got %v", err)
	}
}

func TestLeadingZeros(t *testing.T) {
	tcs := []struct {
		l     Uint128
		r     Uint128
		zeros int
	}{
		{
			l:     New(0x00, 0xf000000000000000),
			r:     New(0x00, 0x8000000000000000),
			zeros: 1,
		},
		{
			l:     New(0x00, 0xf000000000000000),
			r:     New(0x00, 0xc000000000000000),
			zeros: 2,
		},
		{
			l:     New(0x00, 0xf000000000000000),
			r:     New(0x00, 0xe000000000000000),
			zeros: 3,
		},
		{
			l:     New(0x00, 0xffff000000000000),
			r:     New(0x00, 0xff00000000000000),
			zeros: 8,
		},
		{
			l:     New(0x00, 0x000000000000ffff),
			r:     New(0x00, 0x000000000000ff00),
			zeros: 56,
		},
		{
			l:     New(0xf000000000000000, 0x01),
			r:     New(0x4000000000000000, 0x00),
			zeros: 63,
		},
		{
			l:     New(0xf000000000000000, 0x00),
			r:     New(0x4000000000000000, 0x00),
			zeros: 64,
		},
		{
			l:     New(0xf000000000000000, 0x00),
			r:     New(0x8000000000000000, 0x00),
			zeros: 65,
		},
		{
			l:     New(0x00, 0x00),
			r:     New(0x00, 0x00),
			zeros: 128,
		},
		{
			l:     New(0x01, 0x00),
			r:     New(0x00, 0x00),
			zeros: 127,
		},
	}

	for _, tc := range tcs {
		zeros := tc.l.Xor(tc.r).LeadingZeros()
		if zeros != tc.zeros {
			t.Errorf("mismatch (expected: %d, got: %d)", tc.zeros, zeros)
		}
	}
}

func TestString(t *testing.T) {
	for i := 0; i < 1000; i++ {
		x := randUint128()
		if x.String() != x.Big().String() {
			t.Fatalf("mismatch:\n%v !=\n%v", x.String(), x.Big().String())
		}
		y, err := Parse(x.String())
		if err != nil {
			t.Fatal(err)
		} else if !y.Equals(x) {
			t.Fatalf("mismatch:\n%v !=\n%v", x.String(), y.String())
		}
	}
	// Test 0 string
	if Zero.String() != "0" {
		t.Fatalf(`Zero.String() should be "0", got %q`, Zero.String())
	}
	// Test Max string
	if Max.String() != "340282366920938463463374607431768211455" {
		t.Fatalf(`Max.String() should be "0", got %q`, Max.String())
	}
	// Test parsing invalid strings
	if _, err := Parse("-1"); err == nil {
		t.Fatal("expected error when parsing -1")
	}
	if _, err := Parse("340282366920938463463374607431768211456"); err == nil {
		t.Fatal("expected error when parsing max+1")
	}
}

func TestPutBytesBE(t *testing.T) {
	ipa := "2001:db8::1"
	ips := "42540766411282592856903984951653826561"

	u, err := Parse(ips)
	if err != nil {
		t.Fatal(err)
	}

	ip := net.IPv6zero
	u.PutBytesBE(ip)

	if ip.String() != ipa {
		t.Fatalf("mismatch:\n%v !=\n%v", ip, ipa)
	}
}

func TestFromBytesBE(t *testing.T) {
	ip := net.ParseIP("2001:db8::2")
	ips := "42540766411282592856903984951653826562"

	u1 := NewFromBytesBE(ip)
	u2, err := Parse(ips)
	if err != nil {
		t.Fatal(err)
	}
	if u1 != u2 {
		t.Fatalf("mismatch:\n%v !=\n%v", u1, u2)
	}
}

func TestAppendBytes(t *testing.T) {
	u := randUint128()
	v := randUint128()

	b := u.AppendBytes(nil)
	b = v.AppendBytes(b)

	if len(b) != 2*16 {
		t.Fatal("AppendBytes twice should append 32 bytes, got:", len(b))
	}
	if NewFromBytes(b) != u {
		t.Fatal("NewFromBytes is not the inverse of AppendBytes for", u)
	}
	if NewFromBytes(b[16:]) != v {
		t.Fatal("NewFromBytes is not the inverse of AppendBytes for", v)
	}
}

func TestAppendBytesBE(t *testing.T) {
	u := randUint128()
	v := randUint128()

	b := u.AppendBytesBE(nil)
	b = v.AppendBytesBE(b)

	if len(b) != 2*16 {
		t.Fatal("AppendBytesBE twice should append 32 bytes, got:", len(b))
	}
	if NewFromBytesBE(b) != u {
		t.Fatal("NewFromBytesBE is not the inverse of AppendBytesBE for", u)
	}
	if NewFromBytesBE(b[16:]) != v {
		t.Fatal("NewFromBytesBE is not the inverse of AppendBytesBE for", v)
	}
}

func TestMarshalText(t *testing.T) {
	type testStruct struct {
		Foo Uint128
		Bar *Uint128
	}

	test := testStruct{
		Foo: NewFromUint64(math.MaxUint64),
		Bar: &Max,
	}
	b, err := xml.Marshal(test)
	if err != nil {
		t.Fatal(err)
	}
	var test2 testStruct
	err = xml.Unmarshal(b, &test2)
	if err != nil {
		t.Fatal(err)
	}
	if !test2.Foo.Equals(test.Foo) || !test2.Bar.Equals(*test.Bar) {
		t.Fatalf("mismatch:\n%v !=\n%v", test2, test)
	}

	// json should also work
	b, err = json.Marshal(test)
	if err != nil {
		t.Fatal(err)
	}
	if exp := `{"Foo":"18446744073709551615","Bar":"340282366920938463463374607431768211455"}`; string(b) != exp {
		t.Fatalf("mismatch:\n%s !=\n%s", b, exp)
	}
	test2 = testStruct{}
	err = json.Unmarshal(b, &test2)
	if err != nil {
		t.Fatal(err)
	}
	if !test2.Foo.Equals(test.Foo) || !test2.Bar.Equals(*test.Bar) {
		t.Fatalf("mismatch:\n%v !=\n%v", test2, test)
	}
}
