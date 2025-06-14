package decimal

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/qntx/decimal/uint128"
	"github.com/qntx/decimal/uint256"
)

var (
	// defaultPrec is the default number of digits after the decimal point
	// if not specified
	defaultPrec uint8 = 19

	// maxPrec is the maximum number of digits after the decimal point
	maxPrec uint8 = 19

	// maxStrLen is the maximum length of string input when using Parse/MustParse
	// set it to 200 so string length value can fit in 1 byte (for MarshalBinary).
	// Also such that big number (more than 200 digits) is unrealistic in financial system
	// which this library is mainly designed for
	maxStrLen = 200
)

// pre-computed values
var pow10 = [39]uint128.Uint128{
	uint128.NewFromUint64(1),                                        // 10^0
	uint128.NewFromUint64(10),                                       // 10^1
	uint128.NewFromUint64(1e2),                                      // 10^2
	uint128.NewFromUint64(1e3),                                      // 10^3
	uint128.NewFromUint64(1e4),                                      // 10^4
	uint128.NewFromUint64(1e5),                                      // 10^5
	uint128.NewFromUint64(1e6),                                      // 10^6
	uint128.NewFromUint64(1e7),                                      // 10^7
	uint128.NewFromUint64(1e8),                                      // 10^8
	uint128.NewFromUint64(1e9),                                      // 10^9
	uint128.NewFromUint64(1e10),                                     // 10^10
	uint128.NewFromUint64(1e11),                                     // 10^11
	uint128.NewFromUint64(1e12),                                     // 10^12
	uint128.NewFromUint64(1e13),                                     // 10^13
	uint128.NewFromUint64(1e14),                                     // 10^14
	uint128.NewFromUint64(1e15),                                     // 10^15
	uint128.NewFromUint64(1e16),                                     // 10^16
	uint128.NewFromUint64(1e17),                                     // 10^17
	uint128.NewFromUint64(1e18),                                     // 10^18
	uint128.NewFromUint64(1e19),                                     // 10^19
	uint128.New(7_766_279_631_452_241_920, 5),                       // 10^20
	uint128.New(3_875_820_019_684_212_736, 54),                      // 10^21
	uint128.New(1_864_712_049_423_024_128, 542),                     // 10^22
	uint128.New(200_376_420_520_689_664, 5_421),                     // 10^23
	uint128.New(2_003_764_205_206_896_640, 54_210),                  // 10^24
	uint128.New(1_590_897_978_359_414_784, 542_101),                 // 10^25
	uint128.New(15_908_979_783_594_147_840, 5_421_010),              // 10^26
	uint128.New(11_515_845_246_265_065_472, 54_210_108),             // 10^27
	uint128.New(4_477_988_020_393_345_024, 542_101_086),             // 10^28
	uint128.New(7_886_392_056_514_347_008, 5_421_010_862),           // 10^29
	uint128.New(5_076_944_270_305_263_616, 54_210_108_624),          // 10^30
	uint128.New(1_387_595_455_563_353_2928, 542_101_086_242),        // 10^31
	uint128.New(9_632_337_040_368_467_968, 5_421_010_862_427),       // 10^32
	uint128.New(4_089_650_035_136_921_600, 54_210_108_624_275),      // 10^33
	uint128.New(4_003_012_203_950_112_768, 542_101_086_242_752),     // 10^34
	uint128.New(3_136_633_892_082_024_448, 5_421_010_862_427_522),   // 10^35
	uint128.New(12_919_594_847_110_692_864, 54_210_108_624_275_221), // 10^36
	uint128.New(68_739_955_140_067_328, 542_101_086_242_752_217),    // 10^37
	uint128.New(687_399_551_400_673_280, 5_421_010_862_427_522_170), // 10^38
}

var pow10Big = [20]*big.Int{
	big.NewInt(1),    // 10^0
	big.NewInt(10),   // 10^1
	big.NewInt(1e2),  // 10^2
	big.NewInt(1e3),  // 10^3
	big.NewInt(1e4),  // 10^4
	big.NewInt(1e5),  // 10^5
	big.NewInt(1e6),  // 10^6
	big.NewInt(1e7),  // 10^7
	big.NewInt(1e8),  // 10^8
	big.NewInt(1e9),  // 10^9
	big.NewInt(1e10), // 10^10
	big.NewInt(1e11), // 10^11
	big.NewInt(1e12), // 10^12
	big.NewInt(1e13), // 10^13
	big.NewInt(1e14), // 10^14
	big.NewInt(1e15), // 10^15
	big.NewInt(1e16), // 10^16
	big.NewInt(1e17), // 10^17
	big.NewInt(1e18), // 10^18
	pow10[19].Big(),  // 10^19
}

var (
	errOverflow = fmt.Errorf("overflow")

	// ErrPrecOutOfRange is returned when the decimal precision is greater than the default precision
	// default precision can be configured using SetDefaultPrecision, and its value is up to 19
	ErrPrecOutOfRange = fmt.Errorf("precision out of range. Only support maximum %d digits after the decimal point", defaultPrec)

	// ErrEmptyString is returned when the input string is empty
	ErrEmptyString = fmt.Errorf("can't parse empty string")

	// ErrMaxStrLen is returned when the input string exceeds the maximum length
	// Maximum length is arbitrarily set to 200 so string length value can fit in 1 byte (for MarshalBinary).
	// Also such that big number (more than 200 digits) is unrealistic in financial system
	// which this library is mainly designed for.
	ErrMaxStrLen = fmt.Errorf("string input exceeds maximum length %d", maxStrLen)

	// ErrInvalidFormat is returned when the input string is not in the correct format
	// It doesn't support scientific notation, such as 1e-2, 1.23e4, etc.
	ErrInvalidFormat = fmt.Errorf("invalid format")

	// ErrDivideByZero is returned when dividing by zero
	ErrDivideByZero = fmt.Errorf("can't divide by zero")

	// ErrSqrtNegative is returned when calculating square root of negative number
	ErrSqrtNegative = fmt.Errorf("can't calculate square root of negative number")

	// ErrInvalidBinaryData is returned when unmarshalling invalid binary data
	// The binary data should follow the format as described in MarshalBinary
	ErrInvalidBinaryData = fmt.Errorf("invalid binary data")

	// ErrZeroPowNegative is returned when raising zero to a negative power
	ErrZeroPowNegative = fmt.Errorf("can't raise zero to a negative power")

	// ErrExponentTooLarge is returned when the exponent is too large and becomes impractical.
	ErrExponentTooLarge = fmt.Errorf("exponent is too large. Must be less than or equal math.MaxInt32")

	// ErrIntPartOverflow is returned when the integer part of the decimal is too large to fit in int64
	ErrIntPartOverflow = fmt.Errorf("integer part is too large to fit in int64")
)

var (
	Zero    = Decimal{}
	One     = MustFromInt64(1, 0)
	oneUnit = MustFromUint64(1, 19)
)

// Decimal represents a fixed-point decimal number.
// The number is represented as a struct with three fields: coef, neg, and prec.
//
//   - coef: the coefficient of the decimal number
//   - neg: true if the number is negative
//   - prec: the number of digits after the decimal point (0 to 19)
//
// Decimal numbers are immutable and can be used in arithmetic operations such as addition, subtraction, multiplication, and division.
type Decimal struct {
	coef bint
	neg  bool // true if number is negative
	prec uint8
}

// SetDefaultPrecision changes the default precision for decimal numbers in the package.
// Max precision is 19 and is also default.
//
// This function is particularly useful when you want to have your precision of the deicmal smaller than 19
// across the whole application. It should be called only once at the beginning of your application
//
// Panics if the new precision is greater than 19 (maxPrec) or new precision is 0
func SetDefaultPrecision(prec uint8) {
	if prec > maxPrec {
		panic(fmt.Sprintf("precision out of range. Only allow maximum %d digits after the decimal points", maxPrec))
	}

	if prec == 0 {
		panic("prec must be greater than 0")
	}

	defaultPrec = prec
}

// NewFromHiLo returns a decimal from 128-bit unsigned integer (hi,lo)
func NewFromHiLo(neg bool, hi uint64, lo uint64, prec uint8) (Decimal, error) {
	if prec > defaultPrec {
		return Decimal{}, ErrPrecOutOfRange
	}

	coef := uint128.New(lo, hi)
	return newDecimal(neg, bintFromU128(coef), prec), nil
}

// newDecimal return the decimal.
// This function should be used internally to create a new decimal
// to ensure the Zero value is consistent and avoid unexpected cases.
func newDecimal(neg bool, coef bint, prec uint8) Decimal {
	if coef.IsZero() {
		// make Zero consistent and avoid unexpected cases, such as:
		// - coef = 0 and neg is true
		// - coef = 0 and prec != 0
		// These cases results in incorrect comparison between zero values
		return Zero
	}

	return Decimal{neg: neg, coef: coef, prec: prec}
}

// NewFromUint64 returns a decimal which equals to coef / 10^prec and coef is an uint64
// Trailing zeros wll be removed and the prec will also be adjusted
func NewFromUint64(coef uint64, prec uint8) (Decimal, error) {
	if prec > defaultPrec {
		return Decimal{}, ErrPrecOutOfRange
	}

	return newDecimal(false, bintFromU64(coef), prec), nil
}

// MustFromUint64 similars to NewFromUint64, but panics instead of returning error
func MustFromUint64(coef uint64, prec uint8) Decimal {
	d, err := NewFromUint64(coef, prec)
	if err != nil {
		panic(err)
	}

	return d
}

// NewFromInt64 returns a decimal which equals to coef / 10^prec and coef is an int64.
// Trailing zeros wll be removed and the prec will also be adjusted
func NewFromInt64(coef int64, prec uint8) (Decimal, error) {
	var neg bool
	if coef < 0 {
		neg = true
		coef = -coef
	}

	if prec > defaultPrec {
		return Decimal{}, ErrPrecOutOfRange
	}

	//nolint:gosec // coef is positive, so it's safe to convert to uint64
	return newDecimal(neg, bintFromU64(uint64(coef)), prec), nil
}

// MustFromInt64 similars to NewFromInt64, but panics instead of returning error
func MustFromInt64(coef int64, prec uint8) Decimal {
	d, err := NewFromInt64(coef, prec)
	if err != nil {
		panic(err)
	}

	return d
}

// NewFromFloat64 returns a decimal from float64.
//
// **NOTE**: you'll expect to lose some precision for this method due to FormatFloat. See: https://github.com/golang/go/issues/29491
//
// This method is only suitable for small numbers with low precision. e.g. 1.0001, 0.0001, -123.456, -1000000.123456.
// You should avoid using this method if your input number has high precision.
//
// Returns error when:
//  1. f is NaN or Inf
//  2. error when parsing float to string and then to decimal
func NewFromFloat64(f float64) (Decimal, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return Decimal{}, fmt.Errorf("%w: can't parse float '%v' to Decimal", ErrInvalidFormat, f)
	}

	s := strconv.FormatFloat(f, 'f', -1, 64)
	d, err := Parse(s)
	if err != nil {
		return Decimal{}, fmt.Errorf("can't parse float: %w", err)
	}

	return d, nil
}

// MustFromFloat64 similars to NewFromFloat64, but panics instead of returning error
func MustFromFloat64(f float64) Decimal {
	d, err := NewFromFloat64(f)
	if err != nil {
		panic(err)
	}

	return d
}

// Int64 returns the integer part of the decimal.
// Return error if the decimal is too large to fit in int64.
func (d Decimal) Int64() (int64, error) {
	d1 := d.Trunc(0)

	if d1.coef.overflow() {
		return 0, ErrIntPartOverflow
	}

	if d1.coef.u128.Cmp64(math.MaxInt64) > 0 {
		return 0, ErrIntPartOverflow
	}

	//nolint:gosec // can be safely converted as we already checked if coef.u128 is less than math.MaxInt64 above
	int64Part := int64(d1.coef.u128.Low())
	if d1.neg {
		int64Part = -int64Part
	}

	return int64Part, nil
}

// InexactFloat64 returns the float64 representation of the decimal.
// The result may not be 100% accurate due to the limitation of float64 (less decimal precision).
//
// Caution: this method will not return the exact number if the decimal is too large.
//
//	e.g. 123456789012345678901234567890123456789.9999999999999999999 -> 123456789012345680000000000000000000000
func (d Decimal) InexactFloat64() float64 {
	f, _ := strconv.ParseFloat(d.String(), 64)
	return f
}

// ToHiLo returns the internal representation of the decimal.
func (d Decimal) ToHiLo() (neg bool, hi uint64, lo uint64, prec uint8, ok bool) {
	if d.coef.bigInt != nil {
		return
	}
	return d.neg, d.coef.u128.High(), d.coef.u128.Low(), d.prec, true
}

// Parse parses a number in string to a decimal.
// The string must be in the format of: [+-]d{1,19}[.d{1,19}]
//
// Returns error if:
//  1. empty/invalid string
//  2. the number has more than 19 digits after the decimal point
//  3. string length exceeds maxStrLen (which is 200 characters. See [ErrMaxStrLen] for more details)
func Parse(s string) (Decimal, error) {
	return parseBytes(unsafeStringToBytes(s))
}

func parseBytes(b []byte) (Decimal, error) {
	neg, bint, prec, err := parseBint(b)
	if err != nil {
		return Decimal{}, err
	}

	return newDecimal(neg, bint, prec), nil
}

// MustParse similars to Parse, but pacnis instead of returning error.
func MustParse(s string) Decimal {
	d, err := Parse(s)
	if err != nil {
		panic(err)
	}

	return d
}

// Add returns d + e
func (d Decimal) Add(e Decimal) Decimal {
	dcoef, ecoef := d.coef, e.coef

	var (
		prec uint8
	)

	switch {
	case d.prec == e.prec:
		prec = d.prec
	case d.prec > e.prec:
		prec = d.prec
		ecoef = ecoef.Mul(bintFromU128(pow10[d.prec-e.prec]))
	case d.prec < e.prec:
		prec = e.prec
		dcoef = dcoef.Mul(bintFromU128(pow10[e.prec-d.prec]))
	}

	if d.neg == e.neg {
		return newDecimal(d.neg, dcoef.Add(ecoef), prec)
	}

	// different sign
	switch dcoef.Cmp(ecoef) {
	case 1:
		// dcoef > ecoef, subtract can't overflow
		coef, _ := dcoef.Sub(ecoef)
		return newDecimal(d.neg, coef, prec)
	default:
		// dcoef <= ecoef
		coef, _ := ecoef.Sub(dcoef)
		return newDecimal(e.neg, coef, prec)
	}
}

// Add64 returns d + e where e is a uint64
func (d Decimal) Add64(e uint64) Decimal {
	ecoef := bintFromU64(e).Mul(bintFromU128(pow10[d.prec]))

	if d.neg {
		var (
			dcoef bint
			neg   bool
		)

		if d.coef.GT(ecoef) {
			// can ignore the error as we already check if dcoef > ecoef
			dcoef, _ = d.coef.Sub(ecoef)
			neg = true
		} else {
			dcoef, _ = ecoef.Sub(d.coef)
			neg = false
		}

		return newDecimal(neg, dcoef, d.prec)
	}

	dcoef := d.coef.Add(ecoef)
	return newDecimal(false, dcoef, d.prec)
}

// Sub returns d - e
func (d Decimal) Sub(e Decimal) Decimal {
	dcoef, ecoef := d.coef, e.coef

	var (
		prec uint8
	)

	switch {
	case d.prec == e.prec:
		prec = d.prec
	case d.prec > e.prec:
		prec = d.prec
		ecoef = ecoef.Mul(bintFromU128(pow10[d.prec-e.prec]))
	case d.prec < e.prec:
		prec = e.prec
		dcoef = dcoef.Mul(bintFromU128(pow10[e.prec-d.prec]))
	}

	if d.neg != e.neg {
		// different sign
		coef := dcoef.Add(ecoef)
		return newDecimal(d.neg, coef, prec)
	}

	// same sign
	switch dcoef.Cmp(ecoef) {
	case 1:
		// dcoef > ecoef, subtract can't overflow
		coef, _ := dcoef.Sub(ecoef)
		return newDecimal(d.neg, coef, prec)
	default:
		// dcoef <= ecoef
		coef, _ := ecoef.Sub(dcoef)
		return newDecimal(!d.neg, coef, prec)
	}
}

// Sub64 returns d - e where e is a uint64
func (d Decimal) Sub64(e uint64) Decimal {
	ecoef := bintFromU64(e).Mul(bintFromU128(pow10[d.prec]))

	if !d.neg {
		var (
			dcoef bint
			neg   bool
		)

		if d.coef.GT(ecoef) {
			// dcoef > ecoef, subtract can't overflow
			dcoef, _ = d.coef.Sub(ecoef)
			neg = false
		} else {
			dcoef, _ = ecoef.Sub(d.coef)
			neg = true
		}

		return newDecimal(neg, dcoef, d.prec)
	}

	return newDecimal(true, d.coef.Add(ecoef), d.prec)
}

// Mul returns d * e.
// The result will have at most defaultPrec digits after the decimal point.
func (d Decimal) Mul(e Decimal) Decimal {
	prec := d.prec + e.prec
	neg := d.neg != e.neg

	v, err := tryMulU128(d, e, neg, prec)
	if err == nil {
		return v
	}

	// overflow, try with *big.Int
	dBig := d.coef.GetBig()
	eBig := e.coef.GetBig()

	dBig.Mul(dBig, eBig)
	if prec <= defaultPrec {
		return newDecimal(neg, bintFromBigInt(dBig), prec)
	}

	q, _ := new(big.Int).QuoRem(dBig, pow10[prec-defaultPrec].Big(), new(big.Int))
	return newDecimal(neg, bintFromBigInt(q), defaultPrec)
}

func tryMulU128(d, e Decimal, neg bool, prec uint8) (Decimal, error) {
	if d.coef.overflow() || e.coef.overflow() {
		return Decimal{}, errOverflow
	}

	// Use MulFull to get high and low 128-bit parts of the product.
	hiProd, loProd := d.coef.u128.MulFull(e.coef.u128)

	if prec <= defaultPrec {
		// If hiProd is not zero, the product overflows 128 bits.
		if !hiProd.IsZero() {
			return Decimal{}, errOverflow
		}
		// The coefficient is the low 128-bit part of the product.
		return newDecimal(neg, bintFromU128(loProd), prec), nil
	}

	// For prec > defaultPrec, we need to divide the 256-bit product.
	// Construct the 256-bit number from hiProd and loProd.
	r256 := uint256.New(loProd, hiProd)

	// Get the divisor (pow10[prec-defaultPrec]), which is uint128.Uint128.
	// Ensure prec-defaultPrec is a valid index for pow10, this logic is preserved from original.
	divisor128 := pow10[prec-defaultPrec]

	// Convert divisor to Uint256 for QuoRem.
	divisor256 := uint256.NewFromUint128(divisor128)

	// Perform division: r256 / divisor256.
	q256, _, err := r256.QuoRem(divisor256)
	if err != nil {
		// Handle division errors (e.g., division by zero, though pow10 should be safe).
		return Decimal{}, err
	}

	// The quotient q256 is Uint256. Check if it fits into Uint128.
	if !q256.High().IsZero() {
		// Quotient is too large to fit in Uint128, consider it an overflow.
		return Decimal{}, errOverflow
	}

	// The final 128-bit quotient.
	q128 := q256.Low()

	return newDecimal(neg, bintFromU128(q128), defaultPrec), nil
}

// Mul64 returns d * e where e is a uint64.
// The result will have at most defaultPrec digits after the decimal point.
func (d Decimal) Mul64(v uint64) Decimal {
	if v == 0 {
		return Decimal{}
	}

	if v == 1 {
		return d
	}

	if !d.coef.overflow() {
		coef, err := d.coef.u128.Mul64(v)
		if err == nil {
			return newDecimal(d.neg, bintFromU128(coef), d.prec)
		}
	}

	// overflow, try with *big.Int
	dBig := d.coef.GetBig()
	dBig.Mul(dBig, new(big.Int).SetUint64(v))

	return newDecimal(d.neg, bintFromBigInt(dBig), d.prec)
}

// Div returns d / e.
// If the result has more than defaultPrec fraction digits, it will be truncated to defaultPrec digits.
//
// Returns divide by zero error when e is zero
func (d Decimal) Div(e Decimal) (Decimal, error) {
	if e.coef.IsZero() {
		return Decimal{}, ErrDivideByZero
	}

	neg := d.neg != e.neg

	q, err := tryDivU128(d, e, neg)
	if err == nil {
		return q, nil
	}

	// Need to multiply divident with factor
	// to make sure the total decimal number after the decimal point is defaultPrec
	factor := defaultPrec - (d.prec - e.prec)

	// overflow, try with *big.Int
	dBig := d.coef.GetBig()
	eBig := e.coef.GetBig()

	dBig.Mul(dBig, pow10[factor].Big())
	dBig.Div(dBig, eBig)
	return newDecimal(neg, bintFromBigInt(dBig), defaultPrec), nil
}

func tryDivU128(d, e Decimal, neg bool) (Decimal, error) {
	if d.coef.overflow() || e.coef.overflow() {
		return Decimal{}, errOverflow
	}

	// Need to multiply dividend with factor
	// to make sure the total decimal number after the decimal point is defaultPrec
	factor := defaultPrec - (d.prec - e.prec)

	// Calculate the 256-bit dividend: d.coef.u128 * pow10[factor]
	hiProd, loProd := d.coef.u128.MulFull(pow10[factor])
	dividend256 := uint256.New(loProd, hiProd)

	// The divisor is e.coef.u128. Convert it to Uint256 for QuoRem.
	divisorU128 := e.coef.u128
	// Handle division by zero explicitly, as NewFromUint128(Zero).QuoRem(anything) might not be what's expected
	// or QuoRem itself might panic or return an error for zero divisor.
	if divisorU128.IsZero() {
		return Decimal{}, ErrDivideByZero // Or an appropriate error for division by zero in this context
	}
	divisor256 := uint256.NewFromUint128(divisorU128)

	// Perform division: dividend256 / divisor256
	q256, _, err := dividend256.QuoRem(divisor256)
	if err != nil {
		// This could be ErrDivideByZero if not caught above, or other potential errors from QuoRem.
		return Decimal{}, err
	}

	// The quotient q256 is Uint256. Check if it fits into Uint128.
	if !q256.High().IsZero() {
		// Quotient is too large to fit in Uint128, consider it an overflow.
		return Decimal{}, errOverflow
	}

	// The final 128-bit quotient.
	finalQuo128 := q256.Low()

	return newDecimal(neg, bintFromU128(finalQuo128), defaultPrec), nil
}

// Div64 returns d / e where e is a uint64.
// If the result has more than defaultPrec fraction digits, it will be truncated to defaultPrec digits.
//
// Returns divide by zero error when e is zero
func (d Decimal) Div64(v uint64) (Decimal, error) {
	if v == 0 {
		return Decimal{}, ErrDivideByZero
	}

	if v == 1 {
		return d, nil
	}

	if !d.coef.overflow() {
		// Calculate the scaled dividend: d.coef.u128 * pow10[defaultPrec-d.prec]
		// Ensure defaultPrec-d.prec is a valid index for pow10.
		scaleFactor := defaultPrec - d.prec
		// Add a check for scaleFactor validity if necessary, assuming it's valid based on original code.
		hiProd, loProd := d.coef.u128.MulFull(pow10[scaleFactor])
		dividend256 := uint256.New(loProd, hiProd)

		// Prepare the divisor as Uint256.
		// v is guaranteed to be non-zero here.
		divisor256 := uint256.NewFromUint64(v)

		// Perform division: dividend256 / divisor256.
		// For v > 0, QuoRem is expected to return a nil error.
		q256, _, errQuoRem := dividend256.QuoRem(divisor256)

		// If division was successful (errQuoRem is nil, expected here)
		// and the quotient fits within a uint128.Uint128.
		if errQuoRem == nil && q256.High().IsZero() {
			finalQuo128 := q256.Low()
			return newDecimal(d.neg, bintFromU128(finalQuo128), defaultPrec), nil
		}

		// If quotient overflows uint128 (q256.High() != 0) or if errQuoRem was unexpectedly non-nil,
		// fall through to *big.Int arithmetic.
		// overflow, try with *big.Int
	}

	// Fallback path: d.coef already overflowed, or the fast path above resulted in overflow.
	// try with *big.Int
	dBig := d.coef.GetBig()
	// Ensure defaultPrec-d.prec is a valid index for pow10.
	scaleFactor := defaultPrec - d.prec
	dBig.Mul(dBig, pow10[scaleFactor].Big())
	dBig.Div(dBig, new(big.Int).SetUint64(v))

	return newDecimal(d.neg, bintFromBigInt(dBig), defaultPrec), nil
}

// QuoRem returns q and r where
// - q = d / e and  q is an integer
// - r = d - q * e (r < e and r has the same sign as d)
//
// The implementation is similar to C's fmod function.
// Returns divide by zero error when e is zero
func (d Decimal) QuoRem(e Decimal) (Decimal, Decimal, error) {
	if e.coef.IsZero() {
		return Decimal{}, Decimal{}, ErrDivideByZero
	}

	q, r, err := tryQuoRemU128(d, e)
	if err == nil {
		return q, r, nil
	}

	factor := max(d.prec, e.prec)

	// overflow, try with *big.Int
	dBig := d.coef.GetBig()
	eBig := e.coef.GetBig()

	dBig.Mul(dBig, pow10[factor-d.prec].Big())
	eBig.Mul(eBig, pow10[factor-e.prec].Big())

	qBig, rBig := new(big.Int), new(big.Int)
	qBig.QuoRem(dBig, eBig, rBig)

	q = newDecimal(d.neg != e.neg, bintFromBigInt(qBig), 0)
	r = newDecimal(d.neg, bintFromBigInt(rBig), factor)
	return q, r, nil
}

func tryQuoRemU128(d, e Decimal) (Decimal, Decimal, error) {
	if d.coef.overflow() || e.coef.overflow() {
		return Decimal{}, Decimal{}, errOverflow
	}

	var (
		factor      uint8
		dividend256 uint256.Uint256
		divisorU128 uint128.Uint128 // Scaled divisor, must fit in 128 bits for this path
		err         error
	)

	if d.prec == e.prec {
		factor = d.prec
		// Dividend is d's coefficient, as a 256-bit number (though high part is zero)
		dividend256 = uint256.NewFromUint128(d.coef.u128)
		// Divisor is e's coefficient
		divisorU128 = e.coef.u128
	} else {
		factor = max(d.prec, e.prec)

		// Scale dividend d to common precision 'factor'
		scaleByD := pow10[factor-d.prec]
		hiProdD, loProdD := d.coef.u128.MulFull(scaleByD)
		dividend256 = uint256.New(loProdD, hiProdD)

		// Scale divisor e to common precision 'factor'
		scaleByE := pow10[factor-e.prec]
		divisorU128, err = e.coef.u128.Mul(scaleByE)
		if err != nil {
			// Scaled divisor overflows uint128. This path cannot proceed.
			return Decimal{}, Decimal{}, err // err is ErrOverflow from uint128.Mul
		}
	}

	// Check for division by zero using the 128-bit scaled divisor
	if divisorU128.IsZero() {
		return Decimal{}, Decimal{}, ErrDivideByZero
	}
	// Convert 128-bit divisor to 256-bit for QuoRem method
	divisor256 := uint256.NewFromUint128(divisorU128)

	// Perform 256-bit by 256-bit division
	q256, r256, err := dividend256.QuoRem(divisor256)
	if err != nil {
		// This typically only happens if divisor256 was zero, already checked by divisorU128.IsZero().
		// Propagate error if it occurs for other reasons.
		return Decimal{}, Decimal{}, err
	}

	// Results from QuoRem are uint256.Uint256.
	// For this "fast path", quotient and remainder must fit in uint128.Uint128.
	if !q256.High().IsZero() || !r256.High().IsZero() {
		// Quotient or remainder is too large to fit in uint128. Signal fallback.
		return Decimal{}, Decimal{}, errOverflow
	}

	qFinal128 := q256.Low()
	// The remainder r256 must also fit in uint128, which is checked above.
	// Additionally, the remainder r should be less than the original scaled divisorU128.
	// The QuoRem contract ensures r < divisor, so r256.Lt(divisor256) is true.
	// Since divisor256 was created from divisorU128 (which fits in 128 bits),
	// and r256.High().IsZero() is true, r256.Low() is the correct 128-bit remainder.
	rFinal128 := r256.Low()

	// Construct result Decimals
	// Quotient q has precision 0.
	q := newDecimal(d.neg != e.neg, bintFromU128(qFinal128), 0)
	// Remainder r has precision 'factor'.
	r := newDecimal(d.neg, bintFromU128(rFinal128), factor)

	return q, r, nil
}

// Mod is similar to [Decimal.QuoRem] but only returns the remainder
func (d Decimal) Mod(e Decimal) (Decimal, error) {
	_, r, err := d.QuoRem(e)
	return r, err
}

// Prec returns decimal precision as an integer
func (d Decimal) Prec() int {
	return int(d.prec)
}

// PrecUint returns decimal precision as uint8
// Useful when you want to use the precision
// in other functions like [Decimal.RoundBank] or [Decimal.Trunc] because they accept uint8
//
// Example:
//
//	u := MustParse("0.000001")
//	d := MustParse("123.4567891") // 123.456, prec = 3
//	d = d.Trunc(u.PrecUint()) // 123.456789
func (d Decimal) PrecUint() uint8 {
	return d.prec
}

// Cmp compares two decimals d,e and returns:
//
//	-1 if d < e
//	 0 if d == e
//	+1 if d > e
func (d Decimal) Cmp(e Decimal) int {
	if d.neg && !e.neg {
		return -1
	}

	if !d.neg && e.neg {
		return 1
	}

	// d.neg = e.neg
	if d.neg {
		// both are negative, return the opposite
		return -d.cmpDecSameSign(e)
	}

	return d.cmpDecSameSign(e)
}

// Equal reports whether the two decimals d and e are equal.
func (d Decimal) Equal(e Decimal) bool {
	return d.Cmp(e) == 0
}

// LessThan reports whether d < e.
func (d Decimal) LessThan(e Decimal) bool {
	return d.Cmp(e) == -1
}

// LessThanOrEqual reports whether d <= e.
func (d Decimal) LessThanOrEqual(e Decimal) bool {
	return d.Cmp(e) <= 0
}

// GreaterThan reports whether d > e.
func (d Decimal) GreaterThan(e Decimal) bool {
	return d.Cmp(e) == 1
}

// GreaterThanOrEqual reports whether d >= e.
func (d Decimal) GreaterThanOrEqual(e Decimal) bool {
	return d.Cmp(e) >= 0
}

// Max returns the maximum decimal from the list of decimals.
func Max(a Decimal, b ...Decimal) Decimal {
	result := a
	for _, v := range b {
		if v.GreaterThan(result) {
			result = v
		}
	}

	return result
}

func Min(a Decimal, b ...Decimal) Decimal {
	result := a
	for _, v := range b {
		if v.LessThan(result) {
			result = v
		}
	}

	return result
}

func (d Decimal) cmpDecSameSign(e Decimal) int {
	result, err := tryCmpU128(d, e)
	if err == nil {
		return result
	}

	// overflow, fallback to big.Int
	dBig := d.coef.GetBig()
	eBig := e.coef.GetBig()

	if d.prec == e.prec {
		return dBig.Cmp(eBig)
	}

	if d.prec < e.prec {
		dBig.Mul(dBig, pow10[e.prec-d.prec].Big())
	} else {
		eBig.Mul(eBig, pow10[d.prec-e.prec].Big())
	}

	return dBig.Cmp(eBig)
}

func tryCmpU128(d, e Decimal) (int, error) {
	if d.coef.overflow() || e.coef.overflow() {
		return 0, errOverflow // Fast path not applicable if coefficients are already big.Ints
	}

	if d.prec == e.prec {
		return d.coef.u128.Cmp(e.coef.u128), nil
	}

	// Precisions are different, scale one of the numbers to match the other's precision.
	if d.prec < e.prec {
		// e has more fractional digits (e.g., d=1.2, e=1.23). Scale d.
		// Compare d * 10^(e.prec - d.prec) with e.
		scaleFactor := pow10[e.prec-d.prec]
		hiProdD, loProdD := d.coef.u128.MulFull(scaleFactor)
		valD256 := uint256.New(loProdD, hiProdD)
		// valD256 is (d * scaleFactor), e.coef.u128 is e's coefficient.
		// Cmp128(v uint128.Uint128) compares the Uint256 receiver with v.
		return valD256.Cmp128(e.coef.u128), nil
	} else { // d.prec > e.prec
		// d has more fractional digits (e.g., d=1.23, e=1.2). Scale e.
		// Compare d with e * 10^(d.prec - e.prec).
		scaleFactor := pow10[d.prec-e.prec]
		hiProdE, loProdE := e.coef.u128.MulFull(scaleFactor)
		valE256 := uint256.New(loProdE, hiProdE)
		// valE256 is (e * scaleFactor), d.coef.u128 is d's coefficient.
		// We want to calculate d.coef.u128 Cmp valE256.
		// valE256.Cmp128(d.coef.u128) calculates (valE256 Cmp d.coef.u128).
		// So, the result needs to be negated.
		return -valE256.Cmp128(d.coef.u128), nil
	}
}

// Rescale returns the decimal with the new prec only if the new prec is greater than the current prec.
// Useful when you want to increase the prec of the decimal for display purposes.
//
// Example:
//
//	d := MustParse("123.456") // 123.456, prec = 3
//	d.rescale(5) // 123.45600, prec = 5
func (d Decimal) rescale(prec uint8) Decimal {
	dTrim := d.trimTrailingZeros()

	if prec > maxPrec {
		prec = maxPrec
	}

	if prec <= dTrim.prec {
		return dTrim
	}

	diff := prec - dTrim.prec
	coef := dTrim.coef.Mul(bintFromU128(pow10[diff]))

	// only this case we're not using  newDecimal to apply exact precision to zero value
	// this happens when calling StringFixed with 0: 0.StringFixed(5) -> "0.00000"
	// If we use newDecimal, the precision will always be 0, then 0.StringFixed(5) -> "0"
	return Decimal{neg: dTrim.neg, coef: coef, prec: prec}
}

// Neg returns -d
func (d Decimal) Neg() Decimal {
	return newDecimal(!d.neg, d.coef, d.prec)
}

// Abs returns |d|
func (d Decimal) Abs() Decimal {
	return newDecimal(false, d.coef, d.prec)
}

// Sign returns:
//
//	-1 if d < 0
//	 0 if d == 0
//	+1 if d > 0
func (d Decimal) Sign() int {
	// check this first
	// because we allow parsing "-0" into decimal, which results in d.neg = true and d.coef = 0
	if d.coef.IsZero() {
		return 0
	}

	if d.neg {
		return -1
	}

	return 1
}

// IsZero returns
//
//	true if d == 0
//	false if d != 0
func (d Decimal) IsZero() bool {
	return d.coef.IsZero()
}

// IsNeg returns
//
//	true if d < 0
//	false if d >= 0
func (d Decimal) IsNeg() bool {
	return d.neg && !d.coef.IsZero()
}

// IsPos returns
//
//	true if d > 0
//	false if d <= 0
func (d Decimal) IsPos() bool {
	return !d.neg && !d.coef.IsZero()
}

// RoundBank uses half up to even (banker's rounding) to round the decimal to the specified prec.
//
// Examples:
//
//	RoundBank(1.12345, 4) = 1.1234
//	RoundBank(1.12335, 4) = 1.1234
//	RoundBank(1.5, 0) = 2
//	RoundBank(-1.5, 0) = -2
func (d Decimal) RoundBank(prec uint8) Decimal {
	if prec >= d.prec {
		return d
	}

	factor := pow10[d.prec-prec]
	lo := factor.Low() / 2

	if !d.coef.overflow() {
		var err error
		q, r := d.coef.u128.QuoRem64(factor.Low())
		if lo < r || (lo == r && q.Low()%2 == 1) {
			q, err = q.Add64(1)
		}

		// no overflow, return the result
		if err == nil {
			return newDecimal(d.neg, bintFromU128(q), prec)
		}
	}

	// overflow, fallback to big.Int
	dBig := d.coef.GetBig()
	q, r := new(big.Int).QuoRem(dBig, factor.Big(), new(big.Int))

	loBig := new(big.Int).SetUint64(lo)
	if r.Cmp(loBig) > 0 || (r.Cmp(loBig) == 0 && q.Bit(0) == 1) {
		q.Add(q, bigOne)
	}

	return newDecimal(d.neg, bintFromBigInt(q), prec)
}

// RoundAwayFromZero rounds the decimal to the specified prec using AWAY FROM ZERO method (https://en.wikipedia.org/wiki/Rounding#Rounding_away_from_zero).
// If differs from HALF AWAY FROM ZERO in a way that the number is always rounded away from zero (or to infinity) no matter if is 0.5 or not.
// In other libraries or languages, this method is also known as ROUND_UP.
//
// Examples:
//
//	Round(1.12, 1) = 1.2
//	Round(1.15, 1) = 1.2
//	Round(-1.12, 1) = -1.12
//	Round(-1.15, 1) = -1.12
func (d Decimal) RoundAwayFromZero(prec uint8) Decimal {
	if prec >= d.prec {
		return d
	}

	factor := pow10[d.prec-prec]

	if !d.coef.overflow() {
		var err error
		q, r := d.coef.u128.QuoRem64(factor.Low())

		if r != 0 {
			q, err = q.Add64(1)
		}

		if err == nil {
			return newDecimal(d.neg, bintFromU128(q), prec)
		}
	}

	// overflow, fallback to big.Int
	dBig := d.coef.GetBig()
	q, r := new(big.Int).QuoRem(dBig, factor.Big(), new(big.Int))

	if r.Cmp(bigZero) != 0 {
		q.Add(q, bigOne)
	}

	return newDecimal(d.neg, bintFromBigInt(q), prec)
}

// RoundHAZ rounds the decimal to the specified prec using HALF AWAY FROM ZERO method (https://en.wikipedia.org/wiki/Rounding#Rounding_half_away_from_zero).
//
// Examples:
//
//	Round(1.12345, 4) = 1.1235
//	Round(1.12335, 4) = 1.1234
//	Round(1.5, 0) = 2
//	Round(-1.5, 0) = -2
func (d Decimal) RoundHAZ(prec uint8) Decimal {
	if prec >= d.prec {
		return d
	}

	factor := pow10[d.prec-prec]
	half, _ := factor.QuoRem64(2)

	if !d.coef.overflow() {
		var err error
		q, r := d.coef.u128.QuoRem64(factor.Low())
		if half.Cmp64(r) <= 0 {
			q, err = q.Add64(1)
		}

		if err == nil {
			return newDecimal(d.neg, bintFromU128(q), prec)
		}
	}

	// overflow, fallback to big.Int
	dBig := d.coef.GetBig()
	q, r := new(big.Int).QuoRem(dBig, factor.Big(), new(big.Int))

	loBig := half.Big()
	if r.Cmp(loBig) >= 0 {
		q.Add(q, bigOne)
	}

	return newDecimal(d.neg, bintFromBigInt(q), prec)
}

// RoundHTZ rounds the decimal to the specified prec using HALF TOWARD ZERO method (https://en.wikipedia.org/wiki/Rounding#Rounding_half_toward_zero).
//
// Examples:
//
//	Round(1.12345, 4) = 1.1234
//	Round(1.12335, 4) = 1.1233
//	Round(1.5, 0) = 1
//	Round(-1.5, 0) = -1
func (d Decimal) RoundHTZ(prec uint8) Decimal {
	if prec >= d.prec {
		return d
	}

	factor := pow10[d.prec-prec]
	half, _ := factor.QuoRem64(2)

	if !d.coef.overflow() {
		var err error
		q, r := d.coef.u128.QuoRem64(factor.Low())
		if half.Cmp64(r) < 0 {
			q, err = q.Add64(1)
		}

		if err == nil {
			return newDecimal(d.neg, bintFromU128(q), prec)
		}
	}

	// overflow, fallback to big.Int
	dBig := d.coef.GetBig()
	q, r := new(big.Int).QuoRem(dBig, factor.Big(), new(big.Int))

	loBig := half.Big()
	if r.Cmp(loBig) > 0 {
		q.Add(q, bigOne)
	}

	return newDecimal(d.neg, bintFromBigInt(q), prec)
}

// Floor returns the largest integer value less than or equal to d.
func (d Decimal) Floor() Decimal {
	if d.prec == 0 {
		return d
	}

	if !d.coef.overflow() {
		var err error
		q, r := d.coef.u128.QuoRem64(pow10[d.prec].Low())

		// add 1 if it's negative and there's a remainder, e.g. -1.5 -> -2
		if d.neg && r != 0 {
			q, err = q.Add64(1)
		}

		if err == nil {
			return newDecimal(d.neg, bintFromU128(q), 0)
		}
	}

	// overflow, fallback to big.Int
	dBig := d.coef.GetBig()
	q, r := new(big.Int).QuoRem(dBig, pow10[d.prec].Big(), new(big.Int))

	// add 1 if it's negative and there's a remainder, e.g. -1.5 -> -2
	if d.neg && r.Cmp(bigZero) != 0 {
		q.Add(q, bigOne)
	}

	return newDecimal(d.neg, bintFromBigInt(q), 0)
}

// Ceil returns the smallest integer value greater than or equal to d.
func (d Decimal) Ceil() Decimal {
	if d.prec == 0 {
		return d
	}

	if !d.coef.overflow() {
		var err error
		q, r := d.coef.u128.QuoRem64(pow10[d.prec].Low())

		// add 1 if it's positive and there's a remainder, e.g. 1.5 -> 2
		if !d.neg && r != 0 {
			q, err = q.Add64(1)
		}

		if err == nil {
			return newDecimal(d.neg, bintFromU128(q), 0)
		}
	}

	// overflow, fallback to big.Int
	dBig := d.coef.GetBig()
	q, r := new(big.Int).QuoRem(dBig, pow10[d.prec].Big(), new(big.Int))

	// add 1 if it's positive and there's a remainder, e.g. 1.5 -> 2
	if !d.neg && r.Cmp(bigZero) != 0 {
		q.Add(q, bigOne)
	}

	return newDecimal(d.neg, bintFromBigInt(q), 0)
}

// Trunc returns d after truncating the decimal to the specified prec.
//
// Examples:
//
//	Trunc(1.12345, 4) = 1.1234
//	Trunc(1.12335, 4) = 1.1233
func (d Decimal) Trunc(prec uint8) Decimal {
	if prec >= d.prec {
		return d
	}

	factor := pow10[d.prec-prec]

	if !d.coef.overflow() {
		q, _ := d.coef.u128.QuoRem64(factor.Low())
		return newDecimal(d.neg, bintFromU128(q), prec)
	}

	// overflow, fallback to big.Int
	dBig := d.coef.GetBig()
	q := new(big.Int).Quo(dBig, factor.Big())
	return newDecimal(d.neg, bintFromBigInt(q), prec)
}

func (d Decimal) trimTrailingZeros() Decimal {
	if d.coef.overflow() {
		zeros := trailingZerosBigInt(d.coef.bigInt)

		var (
			dBig = d.coef.GetBig()
			prec uint8
		)

		if zeros == 0 {
			return newDecimal(d.neg, bintFromBigInt(dBig), d.prec)
		}

		if zeros >= d.prec {
			dBig.Div(dBig, pow10[d.prec].Big())
			prec = 0
		} else {
			prec = d.prec - uint8(zeros)
			dBig.Div(dBig, pow10[zeros].Big())
		}

		return newDecimal(d.neg, bintFromBigInt(dBig), prec)
	}

	zeros := trailingZerosU128(d.coef.u128)
	if zeros == 0 {
		return newDecimal(d.neg, bintFromU128(d.coef.u128), d.prec)
	}

	var (
		coef uint128.Uint128
		prec uint8
	)

	if zeros >= d.prec {
		coef, _, _ = d.coef.u128.QuoRem(pow10[d.prec])
		prec = 0
	} else {
		prec = d.prec - zeros
		coef, _, _ = d.coef.u128.QuoRem(pow10[zeros])
	}

	d.coef = bintFromU128(coef)
	d.prec = prec
	return d
}

func trailingZerosBigInt(n *big.Int) uint8 {
	var (
		zeros uint8
		z, m  = new(big.Int), new(big.Int)
	)

	_, m = z.QuoRem(n, pow10Big[16], m)
	if m.Cmp(bigZero) == 0 {
		zeros += 16

		// shortcut because maxPrec = 19
		_, m = z.QuoRem(n, pow10Big[zeros+2], m)
		if m.Cmp(bigZero) == 0 {
			zeros += 2
		}

		_, m = z.QuoRem(n, pow10Big[zeros+1], m)
		if m.Cmp(bigZero) == 0 {
			zeros++
		}

		return zeros
	}

	_, m = z.QuoRem(n, pow10Big[8], m)
	if m.Cmp(bigZero) == 0 {
		zeros += 8
	}

	_, m = z.QuoRem(n, pow10Big[zeros+4], m)
	if m.Cmp(bigZero) == 0 {
		zeros += 4
	}

	_, m = z.QuoRem(n, pow10Big[zeros+2], m)
	if m.Cmp(bigZero) == 0 {
		zeros += 2
	}

	_, m = z.QuoRem(n, pow10Big[zeros+1], m)
	if m.Cmp(bigZero) == 0 {
		zeros++
	}

	return zeros
}

func trailingZerosU128(n uint128.Uint128) uint8 {
	var zeros uint8

	_, rem := n.QuoRem64(1e16)
	if rem == 0 {
		zeros += 16

		_, rem = n.QuoRem64(pow10[zeros+2].Low())
		if rem == 0 {
			zeros += 2
		}

		_, rem = n.QuoRem64(pow10[zeros+1].Low())
		if rem == 0 {
			zeros++
		}

		return zeros
	}

	_, rem = n.QuoRem64(1e8)
	if rem == 0 {
		zeros += 8
	}

	_, rem = n.QuoRem64(pow10[zeros+4].Low())
	if rem == 0 {
		zeros += 4
	}

	_, rem = n.QuoRem64(pow10[zeros+2].Low())
	if rem == 0 {
		zeros += 2
	}

	_, rem = n.QuoRem64(pow10[zeros+1].Low())
	if rem == 0 {
		zeros++
	}

	return zeros
}

// PowToIntPart raises the decimal d to the power of integer part of e (d^int(e)).
// This is useful when the exponent is an integer but stored in [Decimal].
//
// Returns error if:
//   - d is zero and e is a negative integer.
//   - |int(e)| > [math.MaxInt32] (because MaxInt32 is already ~2 billion, supporting more than that value is not practical and unnecessary).
//
// Special cases:
//   - 0^0 = 1
//   - 0^(any negative integer) results in [ErrZeroPowNegative]
//
// Examples:
//
//	PowInt32(0, 0) = 1
//	PowInt32(2, 0) = 1
//	PowInt32(0, 1) = 0
//	PowInt32(0, -1) results in an error
//	PowInt32(2.5, 2.6) = 2.5^2 = 6.25
//	PowInt32(2.5, -2.123) = 2.5^(-2) = 0.16
func (d Decimal) PowToIntPart(e Decimal) (Decimal, error) {
	if d.coef.IsZero() && e.neg {
		return Decimal{}, ErrZeroPowNegative
	}

	eInt := e.Trunc(0)
	if eInt.coef.overflow() || eInt.coef.u128.Cmp64(math.MaxInt32) > 0 {
		return Decimal{}, ErrExponentTooLarge
	}

	// convert eInt to int32
	var exponent int32

	//nolint:gosec // coef is positive, so it's safe to convert to int32
	exponent = int32(eInt.coef.u128.Low())

	if eInt.neg {
		exponent = -exponent
	}

	return d.PowInt32(exponent)
}

// Deprecated: Use [PowInt32] instead for correct handling of 0^0 and negative exponents.
// This function treats 0 raised to any power as 0, which may not align with mathematical conventions
// but is practical in certain cases. See: https://github.com/quagmt/udecimal/issues/25.
//
// PowInt raises the decimal d to the integer power e (d^e).
//
// Special cases:
//   - 0^e = 0 for any integer e
//   - d^0 = 1 for any decimal d ≠ 0
//
// Examples:
//
//	PowInt(0, 0)    = 0
//	PowInt(0, 1)    = 0
//	PowInt(0, -1)   = 0
//	PowInt(2, 0)    = 1
//	PowInt(2.5, 2)  = 6.25
//	PowInt(2.5, -2) = 0.16
func (d Decimal) PowInt(e int) Decimal {
	// check 0 first to avoid 0^0 = 1
	if d.coef.IsZero() {
		return Zero
	}

	if e == 0 {
		return One
	}

	if e == 1 {
		return d
	}

	// rescale fist to remove trailing zeros
	dTrim := d.trimTrailingZeros()

	if e < 0 {
		return dTrim.powIntInverse(-e)
	}

	// e > 1 && d != 0
	q, err := dTrim.tryPowIntU128(e)
	if err == nil {
		return q
	}

	// overflow, fallback to big.Int
	dBig := dTrim.coef.GetBig()
	factor := 0
	powPrecision := int(dTrim.prec) * e
	if powPrecision >= int(defaultPrec) {
		factor = powPrecision - int(defaultPrec)
		powPrecision = int(defaultPrec)
	}

	m := new(big.Int).Exp(bigTen, big.NewInt(int64(factor)), nil)
	dBig = new(big.Int).Exp(dBig, big.NewInt(int64(e)), nil)
	qBig := dBig.Quo(dBig, m)

	neg := d.neg
	if e%2 == 0 {
		neg = false
	}

	//nolint:gosec // powPrecision <= defaultPrec, so it's safe to convert to uint8
	return newDecimal(neg, bintFromBigInt(qBig), uint8(powPrecision))
}

// PowInt32 returns d raised to the power of e, where e is an int32.
//
// Returns:
//   - The result of d raised to the power of e.
//   - An error if d is zero and e is a negative integer.
//
// Special cases:
//   - 0^0 = 1
//   - 0^(any negative integer) results in [ErrZeroPowNegative]
//
// Examples:
//
//	PowInt32(0, 0) = 1
//	PowInt32(2, 0) = 1
//	PowInt32(0, 1) = 0
//	PowInt32(0, -1) results in an error
//	PowInt32(2.5, 2) = 6.25
//	PowInt32(2.5, -2) = 0.16
func (d Decimal) PowInt32(e int32) (Decimal, error) {
	// special case: 0 raised to a negative power
	if d.coef.IsZero() && e < 0 {
		return Decimal{}, ErrZeroPowNegative
	}

	if e == 0 {
		return One, nil
	}

	if e == 1 {
		return d, nil
	}

	// Rescale first to remove trailing zeros
	dTrim := d.trimTrailingZeros()

	if e < 0 {
		return dTrim.powIntInverse(int(-e)), nil
	}

	// e > 1 && d != 0
	q, err := dTrim.tryPowIntU128(int(e))
	if err == nil {
		return q, nil
	}

	// overflow, fallback to big.Int
	dBig := dTrim.coef.GetBig()

	var factor int32
	powPrecision := int32(dTrim.prec) * e
	if powPrecision >= int32(defaultPrec) {
		factor = powPrecision - int32(defaultPrec)
		powPrecision = int32(defaultPrec)
	}

	m := new(big.Int).Exp(bigTen, big.NewInt(int64(factor)), nil)
	dBig = new(big.Int).Exp(dBig, big.NewInt(int64(e)), nil)
	qBig := dBig.Quo(dBig, m)

	neg := d.neg
	if e%2 == 0 {
		neg = false
	}

	//nolint:gosec // powPrecision <= defaultPrec, so it's safe to convert to uint8
	return newDecimal(neg, bintFromBigInt(qBig), uint8(powPrecision)), nil
}

// powIntInverse returns d^(-e), with e > 0
func (d Decimal) powIntInverse(e int) Decimal {
	q, err := d.tryInversePowIntU128(e)
	if err == nil {
		return q
	}

	// overflow, fallback to big.Int
	dBig := d.coef.GetBig()
	powPrecision := int(d.prec) * e

	// d^(-e) = 10^(defaultPrec + e) / d^e (with defaultPrec digits after the decimal point)
	m := new(big.Int).Exp(bigTen, big.NewInt(int64(powPrecision+int(defaultPrec))), nil)
	dBig = new(big.Int).Exp(dBig, big.NewInt(int64(e)), nil)
	qBig := dBig.Quo(m, dBig)

	neg := d.neg
	if e%2 == 0 {
		neg = false
	}

	return newDecimal(neg, bintFromBigInt(qBig), defaultPrec)
}

func (d Decimal) tryPowIntU128(e int) (Decimal, error) {
	if d.coef.overflow() {
		return Decimal{}, errOverflow
	}

	// Handle e=0 and e=1 early as they are common and simple.
	if e == 0 {
		// d^0 = 1. The precision of 1 is typically 0.
		// Or, if d is 0, 0^0 is often 1 in this context, but could be an error.
		// Assuming 1 with precision 0 for d^0.
		return One, nil
	}
	if e == 1 {
		return d, nil
	}

	// The uint256.Pow method expects a uint64 exponent.
	// If e can be negative, this function needs more logic (e.g., 1/d^(-e)).
	// Assuming e is positive based on typical integer power usage in such optimized paths.
	if e < 0 {
		// Or handle d^(-e) by calculating 1 / (d^|e|), which likely falls back to big.Int arithmetic.
		return Decimal{}, errors.New("negative exponent not supported in tryPowIntU128 fast path")
	}

	// Original check: if d.coef.u128.High() (meaning d.coef.u128.hi the uint64 part) != 0 && e >= 4
	// This implies if the number is >= 2^64, and exponent is >=4, result is >= (2^64)^4 = 2^256.
	// This check is still relevant for early exit.
	if d.coef.u128.High() != 0 && e >= 4 { // Assuming HiUint64() gives access to the old .hi field's value for this check
		return Decimal{}, errOverflow
	}

	neg := d.neg
	if e%2 == 0 {
		neg = false
	}

	// Calculate the precision of the result before any scaling for defaultPrec
	// exponentPrec is the precision that d^e would naturally have.
	exponentPrec := int(d.prec) * e
	if exponentPrec > int(defaultPrec)+38 { // Max index for pow10 for scaling
		return Decimal{}, errOverflow
	}

	// Base for the power operation is d's coefficient, as a Uint256.
	base256 := uint256.NewFromUint128(d.coef.u128)

	// Calculate base256 ^ e. uint256.Pow takes uint64 exponent.
	result256, err := base256.Pow(uint64(e))
	if err != nil {
		// This could be an overflow if intermediate products in Pow exceed Uint256.Max
		return Decimal{}, err
	}

	// Case 1: The natural precision of d^e is within defaultPrecision.
	if exponentPrec <= int(defaultPrec) {
		// The result256 (which is d^e without decimal point consideration yet)
		// must fit into a uint128 to be represented by bintFromU128.
		if !result256.High().IsZero() {
			// Quotient is too large to fit in Uint128, consider it an overflow.
			return Decimal{}, errOverflow
		}
		finalCoeff128 := result256.Low()
		return newDecimal(neg, bintFromU128(finalCoeff128), uint8(exponentPrec)), nil
	}

	// Case 2: The natural precision of d^e is greater than defaultPrecision.
	// We need to scale down result256 by dividing by 10^(exponentPrec - defaultPrec).
	scaleDownFactorVal := exponentPrec - int(defaultPrec)
	// pow10 array is indexed up to 38, checked by (exponentPrec > int(defaultPrec)+38)
	divisorU128 := pow10[scaleDownFactorVal]
	divisor256 := uint256.NewFromUint128(divisorU128)

	if divisor256.IsZero() { // Should not happen if scaleDownFactorVal is valid for pow10
		return Decimal{}, ErrDivideByZero
	}
	// Perform division: result256 / divisor256
	q256, _, err := result256.QuoRem(divisor256)
	if err != nil {
		return Decimal{}, err
	}

	// The scaled quotient must fit into uint128.
	if !q256.High().IsZero() {
		return Decimal{}, errOverflow
	}
	finalCoeff128 := q256.Low()
	return newDecimal(neg, bintFromU128(finalCoeff128), defaultPrec), nil
}

func (d Decimal) tryInversePowIntU128(e int) (Decimal, error) {
	if d.coef.overflow() {
		return Decimal{}, errOverflow
	}

	if e == 0 {
		// d^0 = 1, so 1/d^0 = 1. Note: if d is 0, 0^0 is 1, 1/1 = 1.
		// If d is 1, 1/1^e = 1.
		// If d is -1, 1/(-1)^e depends on e. (-1)^e is 1 if e is even, -1 if e is odd.
		// So 1/(-1)^e is 1 if e is even, -1 if e is odd.
		// This matches the sign logic below for d=-1.
		// For d=0, (0)^e is 0 for e > 0, division by zero. (0)^0=1.
		if d.coef.IsZero() { // Handle 1/0^e for e > 0
			return Decimal{}, ErrDivideByZero
		}
		return One, nil // 1/(d^0) = 1/1 = 1, assuming d != 0.
	}

	// Assuming e > 0 based on typical usage for inverse power integer exponent.
	if e < 0 { // e.g. d^(-(-2)) = d^2, should be handled by tryPowIntU128
		return Decimal{}, errors.New("negative exponent in tryInversePowIntU128 means positive power, use tryPowIntU128")
	}

	// Original check for (d.coef)^e potentially overflowing uint256
	if d.coef.u128.High() != 0 && e >= 4 {
		return Decimal{}, errOverflow
	}

	neg := d.neg
	if e%2 == 0 { // (-d)^(-e_even) = 1/(-d)^(e_even) = 1/(d^(e_even)) -> positive
		neg = false
	}

	// Calculate P = d.prec * e + defaultPrec for the numerator 10^P
	exponentNumerator := int(d.prec)*e + int(defaultPrec)
	// This limit ensures 10^P can be constructed using pow10, possibly as pow10[P-38]*pow10[38]
	if exponentNumerator > 76 || exponentNumerator < 0 { // exponentNumerator can be <0 if d.prec*e is very negative, though e is positive.
		return Decimal{}, errOverflow // Numerator 10^P is too large or invalid
	}

	// Calculate denominator B = (d.coef.u128)^e
	baseForPow := uint256.NewFromUint128(d.coef.u128)
	denominatorB_256, err := baseForPow.Pow(uint64(e))
	if err != nil {
		return Decimal{}, err // Error from Pow (e.g. overflow within Pow)
	}

	// Denominator B must fit in uint128 for this fast path
	if !denominatorB_256.High().IsZero() {
		return Decimal{}, errOverflow // (d.coef.u128)^e is >= 2^128
	}
	denominatorB_128 := denominatorB_256.Low()

	if denominatorB_128.IsZero() { // Division by zero if (d.coef.u128)^e = 0
		return Decimal{}, ErrDivideByZero
	}

	var finalQuotient128 uint128.Uint128

	if exponentNumerator <= 38 {
		// Numerator is 10^exponentNumerator (fits in uint128)
		numeratorN_128 := pow10[exponentNumerator]
		// Division: (128-bit N) / (128-bit B)
		q, _, errDiv := numeratorN_128.QuoRem(denominatorB_128)
		if errDiv != nil { // Includes division by zero, though checked above for B_128
			return Decimal{}, errDiv
		}
		finalQuotient128 = q
	} else {
		// Numerator is 10^exponentNumerator = pow10[exponentNumerator-38] * pow10[38] (may be 256-bit)
		term1_128 := pow10[exponentNumerator-38]
		term2_128 := pow10[38]
		hiProdN, loProdN := term1_128.MulFull(term2_128)
		numeratorN_256 := uint256.New(loProdN, hiProdN)

		// Division: (256-bit N) / (128-bit B)
		// Convert 128-bit denominator to 256-bit for QuoRem
		denominatorB_forDiv_256 := uint256.NewFromUint128(denominatorB_128)
		q256, _, errDiv := numeratorN_256.QuoRem(denominatorB_forDiv_256)
		if errDiv != nil {
			return Decimal{}, errDiv
		}
		// Resultant quotient must fit in uint128
		if !q256.High().IsZero() {
			return Decimal{}, errOverflow
		}
		finalQuotient128 = q256.Low()
	}

	return newDecimal(neg, bintFromU128(finalQuotient128), defaultPrec), nil
}

// Sqrt returns the square root of d using Newton-Raphson method. (https://en.wikipedia.org/wiki/Newton%27s_method)
// The result will have at most defaultPrec digits after the decimal point.
// Returns error if d < 0
//
// Examples:
//
//	Sqrt(4) = 2
//	Sqrt(2) = 1.4142135623730950488
func (d Decimal) Sqrt() (Decimal, error) {
	if d.neg {
		return Decimal{}, ErrSqrtNegative
	}

	if d.coef.IsZero() {
		return Zero, nil
	}

	if d.Cmp(One) == 0 {
		return One, nil
	}

	if !d.coef.overflow() {
		q, err := d.sqrtU128()
		if err == nil {
			return q, nil
		}
	}

	// overflow, fallback to big.Int
	dBig := d.coef.GetBig()
	factor := 2*defaultPrec - d.prec
	coef := dBig.Mul(dBig, pow10[factor].Big())
	return newDecimal(false, bintFromBigInt(coef.Sqrt(coef)), defaultPrec), nil
}
func (d Decimal) sqrtU128() (Decimal, error) {
	// factor for scaling: d.coef * 10^factor to align precision for sqrt calculation
	// The result of sqrt will have defaultPrec, so (sqrt_coef * 10^-defaultPrec)^2 = d.coef * 10^-d.prec
	// sqrt_coef^2 * 10^(-2*defaultPrec) = d.coef * 10^-d.prec
	// sqrt_coef^2 = d.coef * 10^(2*defaultPrec - d.prec)
	// So, we calculate sqrt(d.coef * 10^factor) where factor = 2*defaultPrec - d.prec
	factor := int(2*defaultPrec) - int(d.prec)

	if factor < 0 || factor >= len(pow10) {
		// This scaling factor is problematic, implies d.prec is too large relative to 2*defaultPrec
		// or factor is too large for pow10 array. This scenario might need fallback to big.Int.
		return Decimal{}, errors.New("scaling factor out of range for sqrtU128")
	}
	scaleVal := pow10[factor]

	// Calculate scaled coefficient: d.coef.u128 * scaleVal
	hiProd, loProd := d.coef.u128.MulFull(scaleVal)
	scaledCoef256 := uint256.New(loProd, hiProd)

	// Original code had a check: coef.carry.hi != 0. coef.carry was u128.
	// This means the high 64 bits of the high 128 bits of the 256-bit product.
	// Equivalent to checking if scaledCoef256 >= 2^(128+64) = 2^192.
	if scaledCoef256.High().High() != 0 { // scaledCoef256.High() is u128, its .High() is uint64
		return Decimal{}, errOverflow // Scaled coefficient is too large
	}

	// Initial guess for Newton-Raphson method
	// x_0 approx sqrt(scaledCoef256)
	bitLen := scaledCoef256.BitLen() // Total bits in scaledCoef256
	if bitLen == 0 {
		return Zero, nil // sqrt(0) is 0
	}

	// Initial guess x0. For sqrt(N), a common guess is 2^(bitLen(N)/2).
	// x_u128 will store our guess, should be uint128 as final result is scaled to defaultPrec.
	var x_u128 uint128.Uint128
	shiftAmount := (uint(bitLen) + 1) / 2
	if shiftAmount >= 128 { // Initial guess itself would overflow uint128 or be max value
		// This case implies scaledCoef256 is very large, near 2^256.
		// sqrt(2^256) = 2^128. So x_u128 should be max uint128 or handle this edge.
		// For simplicity, if initial guess is >= 2^128, it might indicate an issue or need big.Int path.
		// However, the result of sqrt is expected to fit in defaultPrec, which implies x_u128 should be < 2^128.
		// Let's cap the shift if it's too large for Lsh on uint128(1).
		if shiftAmount > 127 {
			x_u128 = uint128.Max // A very large initial guess if N is huge
		} else {
			x_u128 = uint128.New(1, 0).Lsh(shiftAmount)
		}
	} else {
		x_u128 = uint128.New(1, 0).Lsh(shiftAmount)
	}

	// Newton-Raphson iteration: x_new = (x_old + scaledCoef256 / x_old) / 2
	// We expect x_u128 to converge to a value whose square is scaledCoef256.
	// The final result is then scaled by 10^-defaultPrec.
	for {
		if x_u128.IsZero() { // Avoid division by zero if guess becomes 0
			// This shouldn't happen if scaledCoef256 > 0 and initial guess > 0.
			// If scaledCoef256 was 0, we'd have returned already.
			return Decimal{}, errors.New("sqrt iteration led to zero guess")
		}

		// y = scaledCoef256 / x_u128
		// Divisor x_u128 is uint128, convert to Uint256 for QuoRem
		x_as_256 := uint256.NewFromUint128(x_u128)
		y_u256, _, err := scaledCoef256.QuoRem(x_as_256)
		if err != nil { // e.g. division by zero, though x_u128.IsZero() is checked
			return Decimal{}, err
		}

		// y must fit in uint128 because x is converging to sqrt(scaledCoef256)
		// and sqrt(scaledCoef256) should be representable as uint128 (after scaling)
		if !y_u256.High().IsZero() {
			// This implies x_u128 was too small, making y_u256 too large.
			// This could happen if scaledCoef256 is very large, close to (2^128)^2 = 2^256.
			// If y overflows u128, the iteration might be unstable or input is too big for u128 result.
			return Decimal{}, errOverflow // y = scaledCoef/x does not fit in u128
		}
		y_u128 := y_u256.Low()

		// x1 = (x + y) / 2
		x1_u128_sum, carry := x_u128.AddCarry(y_u128, 0) // Assuming AddOverflow exists or use Add and check manually
		if carry != 0 {                                  // (x+y) overflowed uint128
			// This is problematic. If x and y are both large u128, their sum can exceed u128.Max.
			// Then Rsh(1) would be on an incorrect sum.
			// Fallback or error if sum overflows.
			return Decimal{}, errOverflow // (x+y) overflows uint128
		}

		x1_u128 := x1_u128_sum.Rsh(1)

		// Check for convergence
		if x1_u128.Cmp(x_u128) == 0 {
			x_u128 = x1_u128 // ensure the latest value is used if x1 slightly rounded down to x
			break
		}
		x_u128 = x1_u128
	}

	// The result x_u128 is the integer square root of (d.coef * 10^factor).
	// This value corresponds to a Decimal with 'defaultPrec' precision.
	return newDecimal(false, bintFromU128(x_u128), defaultPrec), nil
}
