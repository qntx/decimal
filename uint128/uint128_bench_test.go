package uint128

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"testing"
)

func BenchmarkArithmetic(b *testing.B) {
	randBuf := make([]byte, 17)
	randUint128 := func() Uint128 {
		rand.Read(randBuf)

		var lo, hi uint64
		if randBuf[16]&1 != 0 {
			lo = binary.LittleEndian.Uint64(randBuf[:8])
		}

		if randBuf[16]&2 != 0 {
			hi = binary.LittleEndian.Uint64(randBuf[8:])
		}

		return New(lo, hi)
	}
	x, y := randUint128(), randUint128()

	b.Run("MustAdd native", func(b *testing.B) {
		for range b.N {
			_ = x.lo * y.lo
		}
	})

	b.Run("MustAdd", func(b *testing.B) {
		for range b.N {
			x.MustAdd(y)
		}
	})

	b.Run("Sub", func(b *testing.B) {
		for range b.N {
			x.Sub(y)
		}
	})

	b.Run("Mul", func(b *testing.B) {
		for range b.N {
			x.Mul(y)
		}
	})

	b.Run("Lsh", func(b *testing.B) {
		for range b.N {
			x.Lsh(17)
		}
	})

	b.Run("Rsh", func(b *testing.B) {
		for range b.N {
			x.Rsh(17)
		}
	})

	b.Run("Cmp64", func(b *testing.B) {
		for range b.N {
			x.Cmp64(y.lo)
		}
	})
}

func BenchmarkDivision(b *testing.B) {
	randBuf := make([]byte, 8)
	randU64 := func() uint64 {
		rand.Read(randBuf)

		return binary.LittleEndian.Uint64(randBuf) | 3 // avoid divide-by-zero
	}
	x64 := NewFromUint64(randU64())
	y64 := NewFromUint64(randU64())
	x128 := New(randU64(), randU64())
	y128 := New(randU64(), randU64())

	b.Run("native 64/64", func(b *testing.B) {
		for range b.N {
			_ = x64.lo / y64.lo
		}
	})
	b.Run("Div64 64/64", func(b *testing.B) {
		for range b.N {
			x64.Div64(y64.lo)
		}
	})
	b.Run("Div64 128/64", func(b *testing.B) {
		for range b.N {
			x128.Div64(y64.lo)
		}
	})
	b.Run("Div 64/64", func(b *testing.B) {
		for range b.N {
			x64.Div(y64)
		}
	})
	b.Run("Div 128/64-Lo", func(b *testing.B) {
		x := x128
		x.hi = y64.lo - 1

		for range b.N {
			x.Div(y64)
		}
	})
	b.Run("Div 128/64-Hi", func(b *testing.B) {
		x := x128
		x.hi = y64.lo + 1

		for range b.N {
			x.Div(y64)
		}
	})
	b.Run("Div 128/128", func(b *testing.B) {
		for range b.N {
			x128.Div(y128)
		}
	})
	b.Run("big.Int 128/64", func(b *testing.B) {
		xb, yb := x128.BigInt(), y64.BigInt()
		q := new(big.Int)

		for range b.N {
			q = q.Div(xb, yb)
		}
	})
	b.Run("big.Int 128/128", func(b *testing.B) {
		xb, yb := x128.BigInt(), y128.BigInt()
		q := new(big.Int)

		for range b.N {
			q = q.Div(xb, yb)
		}
	})
}

func BenchmarkString(b *testing.B) {
	buf := make([]byte, 16)
	rand.Read(buf)
	x := New(
		binary.LittleEndian.Uint64(buf[:8]),
		binary.LittleEndian.Uint64(buf[8:]),
	)
	xb := x.BigInt()

	b.Run("Uint128", func(b *testing.B) {
		b.ReportAllocs()

		for range b.N {
			_ = x.String()
		}
	})
	b.Run("big.Int", func(b *testing.B) {
		for range b.N {
			_ = xb.String()
		}
	})
}
