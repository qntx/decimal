package uint256

import (
	"errors"
	"math/big"

	"github.com/qntx/decimal/uint128"
)

var (
	ErrOverflow      = errors.New("uint256: arithmetic overflow")
	ErrUnderflow     = errors.New("uint256: arithmetic underflow")
	ErrDivideByZero  = errors.New("uint256: division by zero")
	ErrNegativeValue = errors.New("uint256: value cannot be negative")
	ErrValueOverflow = errors.New("uint256: value overflows Uint256")
	ErrInvalidBuffer = errors.New("uint256: buffer too short")
)

var (
	Zero = Uint256{}
	One  = Uint256{lo: uint128.NewFromUint64(1)}
	Max  = Uint256{lo: uint128.Max, hi: uint128.Max}
)

type Uint256 struct {
	lo, hi uint128.Uint128
}

// Low returns the lower 128 bits of u.
func (u Uint256) Low() uint128.Uint128 {
	return u.lo
}

// High returns the higher 128 bits of u.
func (u Uint256) High() uint128.Uint128 {
	return u.hi
}

// IsZero returns true if u == 0.
func (u Uint256) IsZero() bool {
	return u.lo.IsZero() && u.hi.IsZero()
}

// Equals returns true if u == v.
func (u Uint256) Equals(v Uint256) bool {
	return u.lo.Equals(v.lo) && u.hi.Equals(v.hi)
}

// Equals128 returns true if u == v.
func (u Uint256) Equals128(v uint128.Uint128) bool {
	return u.hi.IsZero() && u.lo.Equals(v)
}

// Cmp compares u and v and returns:
//
//	-1 if u < v
//	 0 if u == v
//	+1 if u > v
func (u Uint256) Cmp(v Uint256) int {
	if h := u.hi.Cmp(v.hi); h != 0 {
		return h
	}

	return u.lo.Cmp(v.lo)
}

// Cmp128 compares u with v (a Uint128 value).
// It returns:
//
//	+1 if u > v
//	 0 if u == v
//	-1 if u < v
func (u Uint256) Cmp128(v uint128.Uint128) int {
	if !u.hi.IsZero() {
		// If the high part of u is non-zero, u is definitely greater than any Uint128.
		return 1
	}
	// If the high part of u is zero, compare the low part of u with v.
	return u.lo.Cmp(v)
}

// And returns u & v.
func (u Uint256) And(v Uint256) Uint256 {
	return Uint256{u.lo.And(v.lo), u.hi.And(v.hi)}
}

// Or returns u | v.
func (u Uint256) Or(v Uint256) Uint256 {
	return Uint256{u.lo.Or(v.lo), u.hi.Or(v.hi)}
}

// Xor returns u ^ v.
func (u Uint256) Xor(v Uint256) Uint256 {
	return Uint256{u.lo.Xor(v.lo), u.hi.Xor(v.hi)}
}

// Not returns ^u.
func (u Uint256) Not() Uint256 {
	return Uint256{u.lo.Not(), u.hi.Not()}
}

// Lt returns true if u < v.
func (u Uint256) Lt(v Uint256) bool {
	return u.hi.Lt(v.hi) || (u.hi.Equals(v.hi) && u.lo.Lt(v.lo))
}

// Lte returns true if u <= v.
func (u Uint256) Lte(v Uint256) bool {
	return u.hi.Lt(v.hi) || (u.hi.Equals(v.hi) && u.lo.Lte(v.lo))
}

// Gt returns true if u > v.
func (u Uint256) Gt(v Uint256) bool {
	return u.hi.Gt(v.hi) || (u.hi.Equals(v.hi) && u.lo.Gt(v.lo))
}

// Gte returns true if u >= v.
func (u Uint256) Gte(v Uint256) bool {
	return u.hi.Gt(v.hi) || (u.hi.Equals(v.hi) && u.lo.Gte(v.lo))
}

// Bit returns the i-th bit of u.
func (u Uint256) Bit(i uint) uint64 {
	if i >= 256 {
		return 0
	}

	if i >= 128 {
		return u.hi.Bit(i - 128)
	}

	return u.lo.Bit(i)
}

// SetBit sets the i-th bit of u to 1 and returns the new value.
func (u Uint256) SetBit(i uint) Uint256 {
	if i >= 256 {
		return u
	}

	if i >= 128 {
		return Uint256{u.lo, u.hi.SetBit(i - 128)}
	}

	return Uint256{u.lo.SetBit(i), u.hi}
}

// Add returns u + v.
func (u Uint256) Add(v Uint256) (Uint256, error) {
	lo, carryLo := u.lo.AddCarry(v.lo, 0)

	hi, carryHi := u.hi.AddCarry(v.hi, carryLo)
	if carryHi != 0 {
		return Uint256{}, ErrOverflow
	}

	return Uint256{lo, hi}, nil
}

// MustAdd returns u + v, panics on overflow.
func (u Uint256) MustAdd(v Uint256) Uint256 {
	res, err := u.Add(v)
	if err != nil {
		panic(err)
	}

	return res
}

// AddWrap returns u + v, wraps on overflow.
func (u Uint256) AddWrap(v Uint256) Uint256 {
	lo, carryLo := u.lo.AddCarry(v.lo, 0)
	hi, _ := u.hi.AddCarry(v.hi, carryLo)

	return Uint256{lo, hi}
}

// Sub returns u - v.
func (u Uint256) Sub(v Uint256) (Uint256, error) {
	lo, borrowLo := u.lo.SubBorrow(v.lo, 0)

	hi, borrowHi := u.hi.SubBorrow(v.hi, borrowLo)
	if borrowHi != 0 {
		return Uint256{}, ErrUnderflow
	}

	return Uint256{lo, hi}, nil
}

// MustSub returns u - v, panics on underflow.
func (u Uint256) MustSub(v Uint256) Uint256 {
	res, err := u.Sub(v)
	if err != nil {
		panic(err)
	}

	return res
}

// SubWrap returns u - v, wraps on underflow.
func (u Uint256) SubWrap(v Uint256) Uint256 {
	lo, borrowLo := u.lo.SubBorrow(v.lo, 0)
	hi, _ := u.hi.SubBorrow(v.hi, borrowLo)

	return Uint256{lo, hi}
}

// Mul returns u * v.
func (u Uint256) Mul(v Uint256) (Uint256, error) {
	//   u = u_h * 2^128 + u_l
	//   v = v_h * 2^128 + v_l
	// u*v = (u_h*v_h)*2^256 + (u_h*v_l)*2^128 + (u_l*v_h)*2^128 + (u_l*v_l)
	// 1. Calculate u_l * v_l
	// This product can be up to 256 bits.
	// prodHiCarry is the high 128 bits of (u.lo * v.lo)
	// prodLo is the low 128 bits of (u.lo * v.lo)
	prodHiCarry, prodLo := u.lo.MulFull(v.lo)

	// 2. Check for overflow from u_h * v_h term
	// If both u.hi and v.hi are non-zero, (u.hi * v.hi) * 2^256 will surely overflow.
	if !u.hi.IsZero() && !v.hi.IsZero() {
		return Uint256{}, ErrOverflow
	}

	// 3. Calculate cross terms: u_l * v_h and u_h * v_l
	// These terms contribute to the high part of the 256-bit result.
	// Each must fit within a Uint128, otherwise (term * 2^128) would overflow Uint256.

	// termLoHi = u.lo * v.hi
	termLoHi, err := u.lo.Mul(v.hi)
	if err != nil { // Indicates u.lo * v.hi >= 2^128
		return Uint256{}, ErrOverflow
	}

	// termHiLo = u.hi * v.lo
	termHiLo, err := u.hi.Mul(v.lo)
	if err != nil { // Indicates u.hi * v.lo >= 2^128
		return Uint256{}, ErrOverflow
	}

	// 4. Sum parts for the high 128 bits of the result
	// resHi = prodHiCarry + termLoHi + termHiLo
	var resHi uint128.Uint128

	var c1, c2 uint64

	resHi, c1 = prodHiCarry.AddCarry(termLoHi, 0)
	resHi, c2 = resHi.AddCarry(termHiLo, c1)

	// If c2 (final carry) is not 0, the sum of high parts overflowed 128 bits.
	if c2 != 0 {
		return Uint256{}, ErrOverflow
	}

	return Uint256{lo: prodLo, hi: resHi}, nil
}

// MustMul returns u * v, panics on overflow.
func (u Uint256) MustMul(v Uint256) Uint256 {
	res, err := u.Mul(v)
	if err != nil {
		panic(err)
	}

	return res
}

// MulWrap returns u * v, wraps on overflow.
func (u Uint256) MulWrap(v Uint256) Uint256 {
	//   u = u_h * 2^128 + u_l
	//   v = v_h * 2^128 + v_l
	// u*v = (u_h*v_h)*2^256 + (u_h*v_l)*2^128 + (u_l*v_h)*2^128 + (u_l*v_l)
	// For wrapping arithmetic, we are interested in (u*v) mod 2^256.
	// The (u_h*v_h)*2^256 term is ignored in wrapping arithmetic as it's >= 2^256.
	// 1. Calculate u_l * v_l
	// prodHiCarry is the high 128 bits of (u.lo * v.lo)
	// prodLo is the low 128 bits of (u.lo * v.lo)
	prodHiCarry, prodLo := u.lo.MulFull(v.lo) // prodLo is the final low part of the result

	// 2. Calculate cross terms (their low 128 bits)
	// termLoHi = (u.lo * v.hi) mod 2^128
	termLoHi := u.lo.MulWrap(v.hi)

	// termHiLo = (u.hi * v.lo) mod 2^128
	termHiLo := u.hi.MulWrap(v.lo)

	// 3. Sum parts for the high 128 bits of the result, with wrapping
	// resHi = (prodHiCarry + termLoHi + termHiLo) mod 2^128

	// Add first two parts: prodHiCarry + termLoHi
	resHi := prodHiCarry.AddWrap(termLoHi)
	// Add the third part: (prodHiCarry + termLoHi) + termHiLo
	resHi = resHi.AddWrap(termHiLo)

	return Uint256{lo: prodLo, hi: resHi}
}

// Mul128 multiplies u by v (a Uint128 value) and returns the 256-bit product.
// It returns an error if the multiplication overflows.
func (u Uint256) Mul128(v uint128.Uint128) (Uint256, error) {
	// Convert v to a Uint256 with hi part as zero
	vAsUint256 := Uint256{lo: v, hi: uint128.Zero} // hi is uint128.Zero

	return u.Mul(vAsUint256)
}

// quoRemCore implements the restoring division algorithm.
// It is not the most efficient but is simple to implement correctly.
func quoRemCore(u, v Uint256) (q, r Uint256) {
	if u.Lt(v) {
		return Uint256{}, u
	}

	q = Uint256{}

	r = Uint256{}
	for i := 255; i >= 0; i-- {
		r = r.Lsh(1)
		if u.Bit(uint(i)) != 0 {
			r = r.SetBit(0)
		}

		if r.Gte(v) {
			r, _ = r.Sub(v)
			q = q.SetBit(uint(i))
		}
	}

	return
}

// Div returns u / v, panics on divide by zero.
func (u Uint256) Div(v Uint256) (Uint256, error) {
	if v.IsZero() {
		return Zero, ErrDivideByZero
	}

	q, _, err := u.QuoRem(v)

	return q, err
}

// QuoRem returns q = u/v and r = u%v, panics on divide by zero.
func (u Uint256) QuoRem(v Uint256) (q, r Uint256, err error) {
	if v.IsZero() {
		return Zero, Zero, ErrDivideByZero
	}

	q, r = quoRemCore(u, v)

	return q, r, nil
}

// QuoRem128 returns q = u/v and r = u%v, where v is a 128-bit value.
func (u Uint256) QuoRem128(v uint128.Uint128) (q Uint256, r uint128.Uint128, err error) {
	if v.IsZero() {
		return Zero, uint128.Zero, ErrDivideByZero
	}

	v256 := Uint256{lo: v}
	quotient, remainder := quoRemCore(u, v256)
	// The remainder must fit in a Uint128 because the divisor is a Uint128.
	return quotient, remainder.lo, nil
}

// Mod returns u % v.
func (u Uint256) Mod(v Uint256) (Uint256, error) {
	if v.IsZero() {
		return Zero, ErrDivideByZero
	}

	_, r, err := u.QuoRem(v)

	return r, err
}

// Lsh returns u << n.
func (u Uint256) Lsh(n uint) Uint256 {
	if n >= 256 {
		return Zero
	}

	if n >= 128 {
		return Uint256{hi: u.lo.Lsh(n - 128)}
	}

	hi := u.hi.Lsh(n).Or(u.lo.Rsh(128 - n))
	lo := u.lo.Lsh(n)

	return Uint256{lo, hi}
}

// Rsh returns u >> n.
func (u Uint256) Rsh(n uint) Uint256 {
	if n >= 256 {
		return Zero
	}

	if n >= 128 {
		return Uint256{lo: u.hi.Rsh(n - 128)}
	}

	lo := u.lo.Rsh(n).Or(u.hi.Lsh(128 - n))
	hi := u.hi.Rsh(n)

	return Uint256{lo, hi}
}

// LeadingZeros returns the number of leading zeros.
func (u Uint256) LeadingZeros() int {
	if !u.hi.IsZero() {
		return u.hi.LeadingZeros()
	}

	return 128 + u.lo.LeadingZeros()
}

// TrailingZeros returns the number of trailing zeros.
func (u Uint256) TrailingZeros() int {
	if !u.lo.IsZero() {
		return u.lo.TrailingZeros()
	}

	return 128 + u.hi.TrailingZeros()
}

// OnesCount returns the number of 1 bits.
func (u Uint256) OnesCount() int {
	return u.lo.OnesCount() + u.hi.OnesCount()
}

// BitLen returns the minimum number of bits required to represent u.
func (u Uint256) BitLen() int {
	if !u.hi.IsZero() {
		return 128 + u.hi.BitLen()
	}

	return u.lo.BitLen()
}

// String returns the decimal string representation of u.
func (u Uint256) String() string {
	return u.BigInt().String()
}

// PutBytes stores u in little-endian byte slice b.
func (u Uint256) PutBytes(b []byte) {
	if len(b) < 32 {
		panic(ErrInvalidBuffer)
	}

	u.lo.PutBytes(b[:16])
	u.hi.PutBytes(b[16:])
}

// BigInt returns *big.Int representation.
func (u Uint256) BigInt() *big.Int {
	i := u.hi.BigInt()
	i.Lsh(i, 128)
	i.Or(i, u.lo.BigInt())

	return i
}

// New creates a new Uint256.
func New(lo, hi uint128.Uint128) Uint256 {
	return Uint256{lo, hi}
}

// NewFromUint64 converts uint64 to Uint256.
func NewFromUint64(v uint64) Uint256 {
	return Uint256{lo: uint128.NewFromUint64(v)}
}

// NewFromUint128 converts uint128 to Uint256.
func NewFromUint128(v uint128.Uint128) Uint256 {
	return Uint256{lo: v}
}

// NewFromBigInt converts *big.Int to Uint256.
func NewFromBigInt(i *big.Int) (Uint256, error) {
	if i.Sign() < 0 {
		return Zero, ErrNegativeValue
	}

	if i.BitLen() > 256 {
		return Zero, ErrValueOverflow
	}
	// For the low part: create a mask for the lower 128 bits: (1 << 128) - 1
	mask128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))

	// Extract lower 128 bits of i and convert to uint128.Uint128
	iLo := new(big.Int).And(i, mask128)

	lo, errLo := uint128.NewFromBigInt(iLo) // iLo is guaranteed to be <= 128 bits
	if errLo != nil {
		// This should not happen if iLo is correctly formed and uint128.NewFromBigInt is robust.
		return Zero, errLo
	}

	// For the high part: right shift i by 128 bits and convert.
	iHi := new(big.Int).Rsh(i, 128)

	hi, errHi := uint128.NewFromBigInt(iHi) // iHi is what's left, should be <= 128 bits
	if errHi != nil {
		return Zero, errHi
	}

	return Uint256{lo: lo, hi: hi}, nil
}
