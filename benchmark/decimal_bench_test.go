package main

import (
	"math/big"
	"strconv"
	"testing"

	govalues "github.com/govalues/decimal"
	qntx "github.com/qntx/decimal"
	quagmt "github.com/quagmt/udecimal"
	shopspring "github.com/shopspring/decimal"
)

// Test data.
var (
	// Regular numbers.
	s1 = "12345.6789"
	s2 = "9876.54321"

	// High-precision decimals.
	l1 = "12345.1234567890123456789012345678901234567890"
	l2 = "9876.0987654321123456789012345678901234567890"

	// Large and small numbers.
	b1 = "12345678901234567890.123456789"
	b2 = "0.000000000000000000123456789"
)

// big.Float precision.
const (
	precSimple  = 64
	precComplex = 256
)

// big.Float benchmarks.
func BenchmarkBigFloat_Simple_Add(b *testing.B) {
	x, _ := new(big.Float).SetPrec(precSimple).SetString(s1)
	y, _ := new(big.Float).SetPrec(precSimple).SetString(s2)
	res := new(big.Float).SetPrec(precSimple)

	b.ResetTimer()

	for range b.N {
		res.Add(x, y)
	}
}

func BenchmarkBigFloat_Simple_Mul(b *testing.B) {
	x, _ := new(big.Float).SetPrec(precSimple).SetString(s1)
	y, _ := new(big.Float).SetPrec(precSimple).SetString(s2)
	res := new(big.Float).SetPrec(precSimple)

	b.ResetTimer()

	for range b.N {
		res.Mul(x, y)
	}
}

func BenchmarkBigFloat_Complex_Add(b *testing.B) {
	x, _ := new(big.Float).SetPrec(precComplex).SetString(l1)
	y, _ := new(big.Float).SetPrec(precComplex).SetString(l2)
	res := new(big.Float).SetPrec(precComplex)

	b.ResetTimer()

	for range b.N {
		res.Add(x, y)
	}
}

func BenchmarkBigFloat_Complex_Mul(b *testing.B) {
	x, _ := new(big.Float).SetPrec(precComplex).SetString(l1)
	y, _ := new(big.Float).SetPrec(precComplex).SetString(l2)
	res := new(big.Float).SetPrec(precComplex)

	b.ResetTimer()

	for range b.N {
		res.Mul(x, y)
	}
}

func BenchmarkBigFloat_Scale_Div(b *testing.B) {
	x, _ := new(big.Float).SetPrec(precComplex).SetString(b1)
	y, _ := new(big.Float).SetPrec(precComplex).SetString(b2)
	res := new(big.Float).SetPrec(precComplex)

	b.ResetTimer()

	for range b.N {
		res.Quo(x, y)
	}
}

func BenchmarkBigFloat_Parallel_Add(b *testing.B) {
	x, _ := new(big.Float).SetPrec(precSimple).SetString(s1)
	y, _ := new(big.Float).SetPrec(precSimple).SetString(s2)

	b.RunParallel(func(pb *testing.PB) {
		res := new(big.Float).SetPrec(precSimple)
		for pb.Next() {
			res.Add(x, y)
		}
	})
}

// govalues/decimal benchmarks.
func BenchmarkGovalues_Simple_Add(b *testing.B) {
	x, _ := govalues.Parse(s1)
	y, _ := govalues.Parse(s2)

	b.ResetTimer()

	for range b.N {
		_, _ = x.Add(y)
	}
}

func BenchmarkGovalues_Simple_Mul(b *testing.B) {
	x, _ := govalues.Parse(s1)
	y, _ := govalues.Parse(s2)

	b.ResetTimer()

	for range b.N {
		_, _ = x.Mul(y)
	}
}

func BenchmarkGovalues_Complex_Add(b *testing.B) {
	x, _ := govalues.Parse(l1)
	y, _ := govalues.Parse(l2)

	b.ResetTimer()

	for range b.N {
		_, _ = x.Add(y)
	}
}

func BenchmarkGovalues_Complex_Mul(b *testing.B) {
	x, _ := govalues.Parse(l1)
	y, _ := govalues.Parse(l2)

	b.ResetTimer()

	for range b.N {
		_, _ = x.Mul(y)
	}
}

func BenchmarkGovalues_Scale_Div(b *testing.B) {
	x, _ := govalues.Parse(b1)
	y, _ := govalues.Parse(b2)

	b.ResetTimer()

	for range b.N {
		_, _ = x.QuoExact(y, 38)
	}
}

func BenchmarkGovalues_Parallel_Add(b *testing.B) {
	x, _ := govalues.Parse(s1)
	y, _ := govalues.Parse(s2)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = x.Add(y)
		}
	})
}

// quagmt/udecimal benchmarks.
func BenchmarkUdecimal_Simple_Add(b *testing.B) {
	x, _ := quagmt.Parse(s1)
	y, _ := quagmt.Parse(s2)

	b.ResetTimer()

	for range b.N {
		_ = x.Add(y)
	}
}

func BenchmarkUdecimal_Simple_Mul(b *testing.B) {
	x, _ := quagmt.Parse(s1)
	y, _ := quagmt.Parse(s2)

	b.ResetTimer()

	for range b.N {
		_ = x.Mul(y)
	}
}

func BenchmarkUdecimal_Complex_Add(b *testing.B) {
	x, _ := quagmt.Parse(l1)
	y, _ := quagmt.Parse(l2)

	b.ResetTimer()

	for range b.N {
		_ = x.Add(y)
	}
}

func BenchmarkUdecimal_Complex_Mul(b *testing.B) {
	x, _ := quagmt.Parse(l1)
	y, _ := quagmt.Parse(l2)

	b.ResetTimer()

	for range b.N {
		_ = x.Mul(y)
	}
}

func BenchmarkUdecimal_Scale_Div(b *testing.B) {
	x, _ := quagmt.Parse(b1)
	y, _ := quagmt.Parse(b2)

	b.ResetTimer()

	for range b.N {
		_, _ = x.Div(y)
	}
}

func BenchmarkUdecimal_Parallel_Add(b *testing.B) {
	x, _ := quagmt.Parse(s1)
	y, _ := quagmt.Parse(s2)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = x.Add(y)
		}
	})
}

// qntx/decimal benchmarks.
func BenchmarkQntx_Simple_Add(b *testing.B) {
	x, _ := qntx.Parse(s1)
	y, _ := qntx.Parse(s2)

	b.ResetTimer()

	for range b.N {
		_ = x.Add(y)
	}
}

func BenchmarkQntx_Simple_Mul(b *testing.B) {
	x, _ := qntx.Parse(s1)
	y, _ := qntx.Parse(s2)

	b.ResetTimer()

	for range b.N {
		_ = x.Mul(y)
	}
}

func BenchmarkQntx_Complex_Add(b *testing.B) {
	x, _ := qntx.Parse(l1)
	y, _ := qntx.Parse(l2)

	b.ResetTimer()

	for range b.N {
		_ = x.Add(y)
	}
}

func BenchmarkQntx_Complex_Mul(b *testing.B) {
	x, _ := qntx.Parse(l1)
	y, _ := qntx.Parse(l2)

	b.ResetTimer()

	for range b.N {
		_ = x.Mul(y)
	}
}

func BenchmarkQntx_Scale_Div(b *testing.B) {
	x, _ := qntx.Parse(b1)
	y, _ := qntx.Parse(b2)

	b.ResetTimer()

	for range b.N {
		_, _ = x.Div(y)
	}
}

func BenchmarkQntx_Parallel_Add(b *testing.B) {
	x, _ := qntx.Parse(s1)
	y, _ := qntx.Parse(s2)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = x.Add(y)
		}
	})
}

// shopspring/decimal benchmarks.
func BenchmarkShopspring_Simple_Add(b *testing.B) {
	x, _ := shopspring.NewFromString(s1)
	y, _ := shopspring.NewFromString(s2)

	b.ResetTimer()

	for range b.N {
		_ = x.Add(y)
	}
}

func BenchmarkShopspring_Simple_Mul(b *testing.B) {
	x, _ := shopspring.NewFromString(s1)
	y, _ := shopspring.NewFromString(s2)

	b.ResetTimer()

	for range b.N {
		_ = x.Mul(y)
	}
}

func BenchmarkShopspring_Complex_Add(b *testing.B) {
	x, _ := shopspring.NewFromString(l1)
	y, _ := shopspring.NewFromString(l2)

	b.ResetTimer()

	for range b.N {
		_ = x.Add(y)
	}
}

func BenchmarkShopspring_Complex_Mul(b *testing.B) {
	x, _ := shopspring.NewFromString(l1)
	y, _ := shopspring.NewFromString(l2)

	b.ResetTimer()

	for range b.N {
		_ = x.Mul(y)
	}
}

func BenchmarkShopspring_Scale_Div(b *testing.B) {
	x, _ := shopspring.NewFromString(b1)
	y, _ := shopspring.NewFromString(b2)

	b.ResetTimer()

	for range b.N {
		_ = x.DivRound(y, 38)
	}
}

func BenchmarkShopspring_Parallel_Add(b *testing.B) {
	x, _ := shopspring.NewFromString(s1)
	y, _ := shopspring.NewFromString(s2)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = x.Add(y)
		}
	})
}

// float64 benchmarks.
func BenchmarkFloat64_Simple_Add(b *testing.B) {
	x, _ := strconv.ParseFloat(s1, 64)
	y, _ := strconv.ParseFloat(s2, 64)

	b.ResetTimer()

	for range b.N {
		_ = x + y
	}
}

func BenchmarkFloat64_Simple_Mul(b *testing.B) {
	x, _ := strconv.ParseFloat(s1, 64)
	y, _ := strconv.ParseFloat(s2, 64)

	b.ResetTimer()

	for range b.N {
		_ = x * y
	}
}

func BenchmarkFloat64_Complex_Add(b *testing.B) {
	x, _ := strconv.ParseFloat(l1, 64)
	y, _ := strconv.ParseFloat(l2, 64)

	b.ResetTimer()

	for range b.N {
		_ = x + y
	}
}

func BenchmarkFloat64_Complex_Mul(b *testing.B) {
	x, _ := strconv.ParseFloat(l1, 64)
	y, _ := strconv.ParseFloat(l2, 64)

	b.ResetTimer()

	for range b.N {
		_ = x * y
	}
}

func BenchmarkFloat64_Scale_Div(b *testing.B) {
	x, _ := strconv.ParseFloat(b1, 64)
	y, _ := strconv.ParseFloat(b2, 64)

	b.ResetTimer()

	for range b.N {
		_ = x / y
	}
}

func BenchmarkFloat64_Parallel_Add(b *testing.B) {
	x, _ := strconv.ParseFloat(s1, 64)
	y, _ := strconv.ParseFloat(s2, 64)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = x + y
		}
	})
}
