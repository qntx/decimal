package uint256

import (
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// toBig converts a Uint256 to a big.Int.
func toBig(u Uint256) *big.Int {
	// (hi << 128) | lo
	hi := new(big.Int).SetUint64(u.hi.High())
	hi.Lsh(hi, 64)
	hi.Or(hi, new(big.Int).SetUint64(u.hi.Low()))
	hi.Lsh(hi, 128)

	lo := new(big.Int).SetUint64(u.lo.High())
	lo.Lsh(lo, 64)
	lo.Or(lo, new(big.Int).SetUint64(u.lo.Low()))

	return hi.Or(hi, lo)
}

func fromBig(b *big.Int) Uint256 {
	val, err := NewFromBigInt(b)
	if err != nil {
		// This should not happen in tests if we construct big.Int correctly
		panic(err)
	}
	return val
}

func TestDiv(t *testing.T) {
	testCases := []struct {
		name string
		u, v Uint256
	}{
		{"simple", NewFromUint64(100), NewFromUint64(10)},
		{"zero_dividend", NewFromUint64(0), NewFromUint64(10)},
		{"div_by_one", NewFromUint64(12345), NewFromUint64(1)},
		{"u_lt_v", NewFromUint64(10), NewFromUint64(100)},
		{"large_result", MustFromDecimal(decimal.RequireFromString("1e38")), NewFromUint64(10)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, r, err := tc.u.QuoRem(tc.v)
			if err != nil {
				t.Fatalf("QuoRem failed: %v", err)
			}

			bu, bv := toBig(tc.u), toBig(tc.v)
			expQ, expR := new(big.Int).QuoRem(bu, bv, new(big.Int))

			if toBig(q).Cmp(expQ) != 0 {
				t.Errorf("Quotient mismatch:\n got %v\nwant %v", q.String(), expQ.String())
			}
			if toBig(r).Cmp(expR) != 0 {
				t.Errorf("Remainder mismatch:\n got %v\nwant %v", r.String(), expR.String())
			}
		})
	}

	t.Run("div_by_zero", func(t *testing.T) {
		_, _, err := NewFromUint64(100).QuoRem(NewFromUint64(0))
		if err == nil {
			t.Fatal("Expected error for division by zero, but got nil")
		}
	})

	t.Run("randomized", func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < 1000; i++ {
			uBytes := make([]byte, 32)
			vBytes := make([]byte, 32)
			rand.Read(uBytes)
			rand.Read(vBytes)

			bu := new(big.Int).SetBytes(uBytes)
			bv := new(big.Int).SetBytes(vBytes)

			if bv.Sign() == 0 {
				continue // Skip division by zero
			}

			u := fromBig(bu)
			v := fromBig(bv)

			q, r, err := u.QuoRem(v)
			if err != nil {
				t.Fatalf("QuoRem failed for u=%v, v=%v: %v", bu, bv, err)
			}

			expQ, expR := new(big.Int).QuoRem(bu, bv, new(big.Int))

			if toBig(q).Cmp(expQ) != 0 {
				t.Errorf("Quotient mismatch for u=%v, v=%v:\n got %v\nwant %v", bu, bv, q.String(), expQ.String())
			}
			if toBig(r).Cmp(expR) != 0 {
				t.Errorf("Remainder mismatch for u=%v, v=%v:\n got %v\nwant %v", bu, bv, r.String(), expR.String())
			}
		}
	})
}
