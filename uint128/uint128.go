package uint128

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/bits"
)

var (
	ErrOverflow      = errors.New("uint128: arithmetic overflow")
	ErrUnderflow     = errors.New("uint128: arithmetic underflow")
	ErrDivideByZero  = errors.New("uint128: division by zero")
	ErrNegativeValue = errors.New("uint128: value cannot be negative")
	ErrValueOverflow = errors.New("uint128: value overflows Uint128")
	ErrInvalidBuffer = errors.New("uint128: buffer too short")
)

// Zero is a zero-valued uint128.
var Zero Uint128

// Max is the largest possible uint128 value.
var Max = New(math.MaxUint64, math.MaxUint64)

// A Uint128 is an unsigned 128-bit number.
type Uint128 struct {
	lo, hi uint64
}

// Low returns the lower 64 bits of u.
func (u Uint128) Low() uint64 {
	return u.lo
}

// High returns the higher 64 bits of u.
func (u Uint128) High() uint64 {
	return u.hi
}

// IsZero returns true if u == 0.
func (u Uint128) IsZero() bool {
	// NOTE: we do not compare against Zero, because that is a global variable
	// that could be modified.
	return u == Uint128{}
}

// Equals returns true if u == v.
//
// Uint128 values can be compared directly with ==, but use of the Equals method
// is preferred for consistency.
func (u Uint128) Equals(v Uint128) bool {
	return u == v
}

// Equals64 returns true if u == v.
func (u Uint128) Equals64(v uint64) bool {
	return u.lo == v && u.hi == 0
}

// Cmp compares u and v and returns:
//
//	-1 if u <  v
//	 0 if u == v
//	+1 if u >  v
func (u Uint128) Cmp(v Uint128) int {
	if u == v {
		return 0
	} else if u.hi < v.hi || (u.hi == v.hi && u.lo < v.lo) {
		return -1
	} else {
		return 1
	}
}

// Cmp64 compares u and v and returns:
//
//	-1 if u <  v
//	 0 if u == v
//	+1 if u >  v
func (u Uint128) Cmp64(v uint64) int {
	if u.hi == 0 && u.lo == v {
		return 0
	} else if u.hi == 0 && u.lo < v {
		return -1
	} else {
		return 1
	}
}

// And returns u&v.
func (u Uint128) And(v Uint128) Uint128 {
	return Uint128{u.lo & v.lo, u.hi & v.hi}
}

// And64 returns u&v.
func (u Uint128) And64(v uint64) Uint128 {
	return Uint128{u.lo & v, 0}
}

// Or returns u|v.
func (u Uint128) Or(v Uint128) Uint128 {
	return Uint128{u.lo | v.lo, u.hi | v.hi}
}

// Or64 returns u|v.
func (u Uint128) Or64(v uint64) Uint128 {
	return Uint128{u.lo | v, u.hi}
}

// Xor returns u^v.
func (u Uint128) Xor(v Uint128) Uint128 {
	return Uint128{u.lo ^ v.lo, u.hi ^ v.hi}
}

// Xor64 returns u^v.
func (u Uint128) Xor64(v uint64) Uint128 {
	return Uint128{u.lo ^ v, u.hi}
}

// Not returns ^u.
func (u Uint128) Not() Uint128 {
	return Uint128{^u.lo, ^u.hi}
}

// Lt returns true if u < v.
func (u Uint128) Lt(v Uint128) bool {
	return u.hi < v.hi || (u.hi == v.hi && u.lo < v.lo)
}

// Lte returns true if u <= v.
func (u Uint128) Lte(v Uint128) bool {
	return u.hi < v.hi || (u.hi == v.hi && u.lo <= v.lo)
}

// Gt returns true if u > v.
func (u Uint128) Gt(v Uint128) bool {
	return u.hi > v.hi || (u.hi == v.hi && u.lo > v.lo)
}

// Gte returns true if u >= v.
func (u Uint128) Gte(v Uint128) bool {
	return u.hi > v.hi || (u.hi == v.hi && u.lo >= v.lo)
}

// Bit returns the i-th bit of u.
func (u Uint128) Bit(i uint) uint64 {
	if i >= 128 {
		return 0
	}

	if i >= 64 {
		return (u.hi >> (i - 64)) & 1
	}

	return (u.lo >> i) & 1
}

// SetBit sets the i-th bit of u to 1 and returns the new value.
func (u Uint128) SetBit(i uint) Uint128 {
	if i >= 128 {
		return u
	}

	if i >= 64 {
		return Uint128{u.lo, u.hi | (1 << (i - 64))}
	}

	return Uint128{u.lo | (1 << i), u.hi}
}

// Add returns u+v.
func (u Uint128) Add(v Uint128) (Uint128, error) {
	lo, carry := bits.Add64(u.lo, v.lo, 0)
	hi, carry := bits.Add64(u.hi, v.hi, carry)

	if carry != 0 {
		return Uint128{}, ErrOverflow
	}

	return Uint128{lo, hi}, nil
}

// MustAdd returns u+v, panicking on overflow.
func (u Uint128) MustAdd(v Uint128) Uint128 {
	u, err := u.Add(v)
	if err != nil {
		panic(err)
	}

	return u
}

// AddWrap returns u+v with wraparound semantics; for example,
// Max.AddWrap(From64(1)) == Zero.
func (u Uint128) AddWrap(v Uint128) Uint128 {
	lo, carry := bits.Add64(u.lo, v.lo, 0)
	hi, _ := bits.Add64(u.hi, v.hi, carry)

	return Uint128{lo, hi}
}

// Add64 returns u+v.
func (u Uint128) Add64(v uint64) (Uint128, error) {
	lo, carry := bits.Add64(u.lo, v, 0)
	hi, carry := bits.Add64(u.hi, 0, carry)

	if carry != 0 {
		return Uint128{}, ErrOverflow
	}

	return Uint128{lo, hi}, nil
}

// MustAdd64 returns u+v, panicking on overflow.
func (u Uint128) MustAdd64(v uint64) Uint128 {
	u, err := u.Add64(v)
	if err != nil {
		panic(err)
	}

	return u
}

// AddWrap64 returns u+v with wraparound semantics; for example,
// Max.AddWrap64(1) == Zero.
func (u Uint128) AddWrap64(v uint64) Uint128 {
	lo, carry := bits.Add64(u.lo, v, 0)
	hi := u.hi + carry

	return Uint128{lo, hi}
}

// AddCarry returns u+v+carryIn, and the carryOut.
// carryIn and carryOut are 0 or 1.
func (u Uint128) AddCarry(v Uint128, carryIn uint64) (sum Uint128, carryOut uint64) {
	var c0 uint64
	sum.lo, c0 = bits.Add64(u.lo, v.lo, carryIn)
	sum.hi, carryOut = bits.Add64(u.hi, v.hi, c0)

	return
}

// Sub returns u-v.
func (u Uint128) Sub(v Uint128) (Uint128, error) {
	lo, borrow := bits.Sub64(u.lo, v.lo, 0)
	hi, borrow := bits.Sub64(u.hi, v.hi, borrow)

	if borrow != 0 {
		return Uint128{}, ErrUnderflow
	}

	return Uint128{lo, hi}, nil
}

// MustSub returns u-v, panicking on underflow.
func (u Uint128) MustSub(v Uint128) Uint128 {
	u, err := u.Sub(v)
	if err != nil {
		panic(err)
	}

	return u
}

// SubWrap returns u-v with wraparound semantics; for example,
// Zero.SubWrap(From64(1)) == Max.
func (u Uint128) SubWrap(v Uint128) Uint128 {
	lo, borrow := bits.Sub64(u.lo, v.lo, 0)
	hi, _ := bits.Sub64(u.hi, v.hi, borrow)

	return Uint128{lo, hi}
}

// Sub64 returns u-v.
func (u Uint128) Sub64(v uint64) (Uint128, error) {
	lo, borrow := bits.Sub64(u.lo, v, 0)
	hi, borrow := bits.Sub64(u.hi, 0, borrow)

	if borrow != 0 {
		return Uint128{}, ErrUnderflow
	}

	return Uint128{lo, hi}, nil
}

// MustSub64 returns u-v, panicking on underflow.
func (u Uint128) MustSub64(v uint64) Uint128 {
	u, err := u.Sub64(v)
	if err != nil {
		panic(err)
	}

	return u
}

// SubWrap64 returns u-v with wraparound semantics; for example,
// Zero.SubWrap64(1) == Max.
func (u Uint128) SubWrap64(v uint64) Uint128 {
	lo, borrow := bits.Sub64(u.lo, v, 0)
	hi := u.hi - borrow

	return Uint128{lo, hi}
}

// SubBorrow returns u-v-borrowIn, and the borrowOut.
// borrowIn and borrowOut are 0 or 1.
func (u Uint128) SubBorrow(v Uint128, borrowIn uint64) (diff Uint128, borrowOut uint64) {
	var b0 uint64
	diff.lo, b0 = bits.Sub64(u.lo, v.lo, borrowIn)
	diff.hi, borrowOut = bits.Sub64(u.hi, v.hi, b0)

	return
}

// Mul returns u*v, panicking on overflow.
func (u Uint128) Mul(v Uint128) (Uint128, error) {
	hi, lo := bits.Mul64(u.lo, v.lo)
	p0, p1 := bits.Mul64(u.hi, v.lo)
	p2, p3 := bits.Mul64(u.lo, v.hi)
	hi, c0 := bits.Add64(hi, p1, 0)
	hi, c1 := bits.Add64(hi, p3, c0)

	if (u.hi != 0 && v.hi != 0) || p0 != 0 || p2 != 0 || c1 != 0 {
		return Uint128{}, ErrOverflow
	}

	return Uint128{lo, hi}, nil
}

// MustMul returns u*v, panicking on overflow.
func (u Uint128) MustMul(v Uint128) Uint128 {
	u, err := u.Mul(v)
	if err != nil {
		panic(err)
	}

	return u
}

// MulWrap returns u*v with wraparound semantics; for example,
// Max.MulWrap(Max) == 1.
func (u Uint128) MulWrap(v Uint128) Uint128 {
	hi, lo := bits.Mul64(u.lo, v.lo)
	hi += u.hi*v.lo + u.lo*v.hi

	return Uint128{lo, hi}
}

// u * v = hiProduct * 2^128 + loProduct.
func (u Uint128) MulFull(v Uint128) (hiProduct, loProduct Uint128) {
	// Let u = u_h * 2^64 + u_l  (u_h = u.hi, u_l = u.lo)
	// Let v = v_h * 2^64 + v_l  (v_h = v.hi, v_l = v.lo)
	// u * v = (u_h*v_h)*2^128 + (u_h*v_l + u_l*v_h)*2^64 + (u_l*v_l)
	// Calculate u_l * v_l
	// (product_low_high, product_low_low)
	ulvl_h, ulvl_l := bits.Mul64(u.lo, v.lo) // r0 = ulvl_l, carry_to_r1_from_ulvl = ulvl_h

	// Calculate u_h * v_l
	uhvl_h, uhvl_l := bits.Mul64(u.hi, v.lo)

	// Calculate u_l * v_h
	ulvh_h, ulvh_l := bits.Mul64(u.lo, v.hi)

	// Calculate u_h * v_h
	uhvh_h, uhvh_l := bits.Mul64(u.hi, v.hi)

	// Sum middle terms: (u_h*v_l + u_l*v_h) and add carry from u_l*v_l
	// r1 is the low 64 bits of this sum.
	// carry_to_r2 is the high 64 bits of this sum.
	mid1, carry1 := bits.Add64(uhvl_l, ulvh_l, 0)
	mid2, carry2 := bits.Add64(uhvl_h, ulvh_h, carry1) // mid2 is high part of (uhvl + ulvh)

	// Add carry from u_l*v_l (ulvl_h) to the low part of middle sum (mid1)
	r1, carry_to_r2_from_mid_sum_low := bits.Add64(mid1, ulvl_h, 0)

	// Assemble loProduct
	loProduct = Uint128{lo: ulvl_l, hi: r1}

	// Calculate high part of product (r3, r2)
	// Start with high part of middle sum (mid2) and add its carry (carry_to_r2_from_mid_sum_low)
	// Then add low part of u_h*v_h (uhvh_l)
	// And finally add carry from mid2 (carry2)

	r2_part1, carry_to_r3_from_r2_part1 := bits.Add64(mid2, uhvh_l, 0)
	r2_part2, carry_to_r3_from_r2_part2 := bits.Add64(r2_part1, carry_to_r2_from_mid_sum_low, 0)
	r2_final, carry_to_r3_from_r2_final := bits.Add64(r2_part2, carry2, 0) // carry2 was from mid1's high part sum

	// r3 is high part of u_h*v_h (uhvh_h) plus all carries propagated to it
	r3 := uhvh_h + carry_to_r3_from_r2_part1 + carry_to_r3_from_r2_part2 + carry_to_r3_from_r2_final

	hiProduct = Uint128{lo: r2_final, hi: r3}

	return hiProduct, loProduct
}

// Mul64 returns u*v, panicking on overflow.
func (u Uint128) Mul64(v uint64) (Uint128, error) {
	hi, lo := bits.Mul64(u.lo, v)
	p0, p1 := bits.Mul64(u.hi, v)
	hi, c0 := bits.Add64(hi, p1, 0)

	if p0 != 0 || c0 != 0 {
		return Uint128{}, ErrOverflow
	}

	return Uint128{lo, hi}, nil
}

// MustMul64 returns u*v, panicking on overflow.
func (u Uint128) MustMul64(v uint64) Uint128 {
	u, err := u.Mul64(v)
	if err != nil {
		panic(err)
	}

	return u
}

// MulWrap64 returns u*v with wraparound semantics; for example,
// Max.MulWrap64(2) == Max.Sub64(1).
func (u Uint128) MulWrap64(v uint64) Uint128 {
	hi, lo := bits.Mul64(u.lo, v)
	hi += u.hi * v

	return Uint128{lo, hi}
}

// Div returns u/v.
func (u Uint128) Div(v Uint128) (Uint128, error) {
	q, _, err := u.QuoRem(v)

	return q, err
}

// MustDiv returns u/v, panicking on overflow.
func (u Uint128) MustDiv(v Uint128) Uint128 {
	u, err := u.Div(v)
	if err != nil {
		panic(err)
	}

	return u
}

// Div64 returns u/v.
func (u Uint128) Div64(v uint64) Uint128 {
	q, _ := u.QuoRem64(v)

	return q
}

// QuoRem returns q = u/v and r = u%v.
func (u Uint128) QuoRem(v Uint128) (q, r Uint128, err error) {
	if v.hi == 0 {
		var r64 uint64
		q, r64 = u.QuoRem64(v.lo)
		r = NewFromUint64(r64)
	} else {
		// generate a "trial quotient," guaranteed to be within 1 of the actual
		// quotient, then adjust.
		n := uint(bits.LeadingZeros64(v.hi))
		v1 := v.Lsh(n)
		u1 := u.Rsh(1)
		tq, _ := bits.Div64(u1.hi, u1.lo, v1.hi)
		tq >>= 63 - n

		if tq != 0 {
			tq--
		}

		q = NewFromUint64(tq)
		// calculate remainder using trial quotient, then adjust if remainder is
		// greater than divisor
		vq, err := v.Mul64(tq)
		if err != nil {
			return q, r, err
		}

		r, err = u.Sub(vq)
		if err != nil {
			return q, r, err
		}

		if r.Cmp(v) >= 0 {
			q, err = q.Add64(1)
			if err != nil {
				return q, r, err
			}

			r, err = r.Sub(v)
			if err != nil {
				return q, r, err
			}
		}
	}

	return q, r, err
}

// MustQuoRem returns q = u/v and r = u%v, panicking on overflow.
func (u Uint128) MustQuoRem(v Uint128) (q, r Uint128) {
	q, r, err := u.QuoRem(v)
	if err != nil {
		panic(err)
	}

	return q, r
}

// QuoRem64 returns q = u/v and r = u%v.
func (u Uint128) QuoRem64(v uint64) (q Uint128, r uint64) {
	if u.hi < v {
		q.lo, r = bits.Div64(u.hi, u.lo, v)
	} else {
		q.hi, r = bits.Div64(0, u.hi, v)
		q.lo, r = bits.Div64(r, u.lo, v)
	}

	return
}

// Mod returns r = u%v.
func (u Uint128) Mod(v Uint128) (r Uint128, err error) {
	_, r, err = u.QuoRem(v)

	return
}

// MustMod returns r = u%v, panicking on overflow.
func (u Uint128) MustMod(v Uint128) Uint128 {
	u, err := u.Mod(v)
	if err != nil {
		panic(err)
	}

	return u
}

// Mod64 returns r = u%v.
func (u Uint128) Mod64(v uint64) (r uint64) {
	_, r = u.QuoRem64(v)

	return
}

// Lsh returns u<<n.
func (u Uint128) Lsh(n uint) (s Uint128) {
	if n > 64 {
		s.lo = 0
		s.hi = u.lo << (n - 64)
	} else {
		s.lo = u.lo << n
		s.hi = u.hi<<n | u.lo>>(64-n)
	}

	return
}

// Rsh returns u>>n.
func (u Uint128) Rsh(n uint) (s Uint128) {
	if n > 64 {
		s.lo = u.hi >> (n - 64)
		s.hi = 0
	} else {
		s.lo = u.lo>>n | u.hi<<(64-n)
		s.hi = u.hi >> n
	}

	return
}

// LeadingZeros returns the number of leading zero bits in u; the result is 128
// for u == 0.
func (u Uint128) LeadingZeros() int {
	if u.hi > 0 {
		return bits.LeadingZeros64(u.hi)
	}

	return 64 + bits.LeadingZeros64(u.lo)
}

// TrailingZeros returns the number of trailing zero bits in u.
// It returns 128 if u == 0.
func (u Uint128) TrailingZeros() int {
	if u.lo > 0 {
		return bits.TrailingZeros64(u.lo)
	}

	return 64 + bits.TrailingZeros64(u.hi)
}

// OnesCount returns the number of set bits in u.
func (u Uint128) OnesCount() int {
	return bits.OnesCount64(u.lo) + bits.OnesCount64(u.hi)
}

// BitLen returns the minimum number of bits required to represent u.
// The result is 0 for u == 0.
func (u Uint128) BitLen() int {
	if u.hi != 0 {
		return 64 + bits.Len64(u.hi)
	}

	return bits.Len64(u.lo)
}

// RotateLeft returns the value of u rotated left by (k mod 128) bits.
func (u Uint128) RotateLeft(k int) Uint128 {
	const n = 128
	s := uint(k) & (n - 1)

	return u.Lsh(s).Or(u.Rsh(n - s))
}

// RotateRight returns the value of u rotated left by (k mod 128) bits.
func (u Uint128) RotateRight(k int) Uint128 {
	return u.RotateLeft(-k)
}

// Reverse returns the value of u with its bits in reversed order.
func (u Uint128) Reverse() Uint128 {
	return Uint128{bits.Reverse64(u.hi), bits.Reverse64(u.lo)}
}

// ReverseBytes returns the value of u with its bytes in reversed order.
func (u Uint128) ReverseBytes() Uint128 {
	return Uint128{bits.ReverseBytes64(u.hi), bits.ReverseBytes64(u.lo)}
}

// Len returns the minimum number of bits required to represent u; the result is
// 0 for u == 0.
func (u Uint128) Len() int {
	return 128 - u.LeadingZeros()
}

// String returns the base-10 representation of u as a string.
func (u Uint128) String() string {
	if u.IsZero() {
		return "0"
	}

	buf := []byte("0000000000000000000000000000000000000000") // log10(2^128) < 40

	for i := len(buf); ; i -= 19 {
		q, r := u.QuoRem64(1e19) // largest power of 10 that fits in a uint64

		var n int

		for ; r != 0; r /= 10 {
			n++
			buf[i-n] += byte(r % 10)
		}

		if q.IsZero() {
			return string(buf[i-n:])
		}

		u = q
	}
}

// PutBytes stores u in b in little-endian order. It panics if len(b) < 16.
func (u Uint128) PutBytes(b []byte) {
	binary.LittleEndian.PutUint64(b[:8], u.lo)
	binary.LittleEndian.PutUint64(b[8:], u.hi)
}

// PutBytesBE stores u in b in big-endian order. It panics if len(ip) < 16.
func (u Uint128) PutBytesBE(b []byte) {
	binary.BigEndian.PutUint64(b[:8], u.hi)
	binary.BigEndian.PutUint64(b[8:], u.lo)
}

// AppendBytes appends u to b in little-endian order and returns the extended buffer.
func (u Uint128) AppendBytes(b []byte) []byte {
	b = binary.LittleEndian.AppendUint64(b, u.lo)
	b = binary.LittleEndian.AppendUint64(b, u.hi)

	return b
}

// AppendBytesBE appends u to b in big-endian order and returns the extended buffer.
func (u Uint128) AppendBytesBE(b []byte) []byte {
	b = binary.BigEndian.AppendUint64(b, u.hi)
	b = binary.BigEndian.AppendUint64(b, u.lo)

	return b
}

// Big returns u as a *big.Int.
func (u Uint128) Big() *big.Int {
	i := new(big.Int).SetUint64(u.hi)
	i = i.Lsh(i, 64)
	i = i.Xor(i, new(big.Int).SetUint64(u.lo))

	return i
}

// Scan implements fmt.Scanner.
func (u *Uint128) Scan(s fmt.ScanState, ch rune) error {
	i := new(big.Int)
	if err := i.Scan(s, ch); err != nil {
		return err
	} else if i.Sign() < 0 {
		return ErrNegativeValue
	} else if i.BitLen() > 128 {
		return ErrValueOverflow
	}

	u.lo = i.Uint64()
	u.hi = i.Rsh(i, 64).Uint64()

	return nil
}

// New returns the Uint128 value (lo,hi).
func New(lo, hi uint64) Uint128 {
	return Uint128{lo, hi}
}

// NewFromUint64 converts v to a Uint128 value.
func NewFromUint64(v uint64) Uint128 {
	return New(v, 0)
}

// NewFromBytes converts b to a Uint128 value.
func NewFromBytes(b []byte) Uint128 {
	return New(
		binary.LittleEndian.Uint64(b[:8]),
		binary.LittleEndian.Uint64(b[8:]),
	)
}

// NewFromBytesBE converts big-endian b to a Uint128 value.
func NewFromBytesBE(b []byte) Uint128 {
	return New(
		binary.BigEndian.Uint64(b[8:]),
		binary.BigEndian.Uint64(b[:8]),
	)
}

// NewFromBigInt converts i to a Uint128 value. It panics if i is negative or
// overflows 128 bits.
func NewFromBigInt(i *big.Int) (Uint128, error) {
	if i.Sign() < 0 {
		return Uint128{}, ErrNegativeValue
	} else if i.BitLen() > 128 {
		return Uint128{}, ErrValueOverflow
	}

	return Uint128{
		lo: i.Uint64(),
		hi: i.Rsh(i, 64).Uint64(),
	}, nil
}

// Parse parses s as a Uint128 value.
func Parse(s string) (Uint128, error) {
	u := Uint128{}
	_, err := fmt.Sscan(s, &u)

	return u, err
}

// MarshalText implements encoding.TextMarshaler.
func (u Uint128) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (u *Uint128) UnmarshalText(b []byte) error {
	_, err := fmt.Sscan(string(b), u)

	return err
}
