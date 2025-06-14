package main

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/govalues/decimal"
	qdecimal "github.com/qntx/decimal"
	"github.com/quagmt/udecimal"
)

// --- 测试数据准备 ---

var (
	// 场景1: 常规数字
	s1 = "12345.6789"
	s2 = "9876.54321"

	// 场景2: 高精度小数
	l1 = "12345.1234567890"
	l2 = "9876.0987654321"

	// 场景3: 一个极大值和一个极小值
	b1 = "12345678901234567890.123456789"
	b2 = "0.000000000000000000123456789"
)

// --- 1. big.Float 基准测试 ---

// big.Float 使用的精度设置
// 对于加减法，64位精度通常足够
// 对于乘除法，需要更高的精度来容纳结果
const (
	precSimple  = 64
	precComplex = 256
)

// --- 简单运算 ---

func BenchmarkBigFloat_Simple_Add(b *testing.B) {
	x, _ := new(big.Float).SetPrec(precSimple).SetString(s1)
	y, _ := new(big.Float).SetPrec(precSimple).SetString(s2)
	res := new(big.Float).SetPrec(precSimple)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res.Add(x, y)
	}
}

func BenchmarkBigFloat_Simple_Mul(b *testing.B) {
	x, _ := new(big.Float).SetPrec(precSimple).SetString(s1)
	y, _ := new(big.Float).SetPrec(precSimple).SetString(s2)
	res := new(big.Float).SetPrec(precSimple)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res.Mul(x, y)
	}
}

// --- 高精度运算 ---

func BenchmarkBigFloat_Complex_Add(b *testing.B) {
	x, _ := new(big.Float).SetPrec(precComplex).SetString(l1)
	y, _ := new(big.Float).SetPrec(precComplex).SetString(l2)
	res := new(big.Float).SetPrec(precComplex)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res.Add(x, y)
	}
}

func BenchmarkBigFloat_Complex_Mul(b *testing.B) {
	x, _ := new(big.Float).SetPrec(precComplex).SetString(l1)
	y, _ := new(big.Float).SetPrec(precComplex).SetString(l2)
	res := new(big.Float).SetPrec(precComplex)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res.Mul(x, y)
	}
}

// --- 大数与小数运算 ---

func BenchmarkBigFloat_Scale_Div(b *testing.B) {
	x, _ := new(big.Float).SetPrec(precComplex).SetString(b1)
	y, _ := new(big.Float).SetPrec(precComplex).SetString(b2)
	res := new(big.Float).SetPrec(precComplex)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res.Quo(x, y)
	}
}

// --- 并发性能 ---

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

// --- 2. govalues/decimal 基准测试 ---

// --- 简单运算 ---

func BenchmarkGovaluesDecimal_Simple_Add(b *testing.B) {
	x, _ := decimal.Parse(s1)
	y, _ := decimal.Parse(s2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = x.Add(y)
	}
}

func BenchmarkGovaluesDecimal_Simple_Mul(b *testing.B) {
	x, _ := decimal.Parse(s1)
	y, _ := decimal.Parse(s2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = x.Mul(y)
	}
}

// --- 高精度运算 ---

func BenchmarkGovaluesDecimal_Complex_Add(b *testing.B) {
	x, _ := decimal.Parse(l1)
	y, _ := decimal.Parse(l2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = x.Add(y)
	}
}

func BenchmarkGovaluesDecimal_Complex_Mul(b *testing.B) {
	x, _ := decimal.Parse(l1)
	y, _ := decimal.Parse(l2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = x.Mul(y)
	}
}

// --- 大数与小数运算 ---

func BenchmarkGovaluesDecimal_Scale_Div(b *testing.B) {
	x, _ := decimal.Parse(b1)
	y, _ := decimal.Parse(b2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// govalues/decimal 除法需要显式指定精度
		_, _ = x.QuoExact(y, 38)
	}
}

// --- 并发性能 ---

func BenchmarkGovaluesDecimal_Parallel_Add(b *testing.B) {
	x, _ := decimal.Parse(s1)
	y, _ := decimal.Parse(s2)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = x.Add(y)
		}
	})
}

// --- 3. quagmt/udecimal 基准测试 ---

// --- 简单运算 ---

func BenchmarkUdecimal_Simple_Add(b *testing.B) {
	x, _ := udecimal.Parse(s1)
	y, _ := udecimal.Parse(s2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x.Add(y)
	}
}

func BenchmarkUdecimal_Simple_Mul(b *testing.B) {
	x, _ := udecimal.Parse(s1)
	y, _ := udecimal.Parse(s2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x.Mul(y)
	}
}

// --- 高精度运算 ---

func BenchmarkUdecimal_Complex_Add(b *testing.B) {
	x, _ := udecimal.Parse(l1)
	y, _ := udecimal.Parse(l2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x.Add(y)
	}
}

func BenchmarkUdecimal_Complex_Mul(b *testing.B) {
	x, _ := udecimal.Parse(l1)
	y, _ := udecimal.Parse(l2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x.Mul(y)
	}
}

// --- 大数与小数运算 ---

func BenchmarkUdecimal_Scale_Div(b *testing.B) {
	x, _ := udecimal.Parse(b1)
	y, _ := udecimal.Parse(b2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// udecimal 除法自动处理精度
		_, _ = x.Div(y)
	}
}

// --- 并发性能 ---

func BenchmarkUdecimal_Parallel_Add(b *testing.B) {
	x, _ := udecimal.Parse(s1)
	y, _ := udecimal.Parse(s2)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = x.Add(y)
		}
	})
}

// --- 4. qntx/decimal 基准测试 ---

// --- 简单运算 ---

func BenchmarkQdecimal_Simple_Add(b *testing.B) {
	x, _ := qdecimal.Parse(s1)
	y, _ := qdecimal.Parse(s2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x.Add(y)
	}
}

func BenchmarkQdecimal_Simple_Mul(b *testing.B) {
	x, _ := qdecimal.Parse(s1)
	y, _ := qdecimal.Parse(s2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x.Mul(y)
	}
}

// --- 高精度运算 ---

func BenchmarkQdecimal_Complex_Add(b *testing.B) {
	x, _ := qdecimal.Parse(l1)
	y, _ := qdecimal.Parse(l2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x.Add(y)
	}
}

func BenchmarkQdecimal_Complex_Mul(b *testing.B) {
	x, _ := qdecimal.Parse(l1)
	y, _ := qdecimal.Parse(l2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x.Mul(y)
	}
}

// --- 大数与小数运算 ---

func BenchmarkQdecimal_Scale_Div(b *testing.B) {
	x, _ := qdecimal.Parse(b1)
	y, _ := qdecimal.Parse(b2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// udecimal 除法自动处理精度
		_, _ = x.Div(y)
	}
}

// --- 并发性能 ---

func BenchmarkQdecimal_Parallel_Add(b *testing.B) {
	x, _ := qdecimal.Parse(s1)
	y, _ := qdecimal.Parse(s2)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = x.Add(y)
		}
	})
}

// --- 5. float64 基准测试 ---

// --- 简单运算 ---

func BenchmarkFloat64_Simple_Add(b *testing.B) {
	x, _ := strconv.ParseFloat(s1, 64)
	y, _ := strconv.ParseFloat(s2, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x + y
	}
}

func BenchmarkFloat64_Simple_Mul(b *testing.B) {
	x, _ := strconv.ParseFloat(s1, 64)
	y, _ := strconv.ParseFloat(s2, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x * y
	}
}

// --- 高精度运算 (float64 精度固定) ---

func BenchmarkFloat64_Complex_Add(b *testing.B) {
	x, _ := strconv.ParseFloat(l1, 64)
	y, _ := strconv.ParseFloat(l2, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x + y
	}
}

func BenchmarkFloat64_Complex_Mul(b *testing.B) {
	x, _ := strconv.ParseFloat(l1, 64)
	y, _ := strconv.ParseFloat(l2, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x * y
	}
}

// --- 大数与小数运算 ---

func BenchmarkFloat64_Scale_Div(b *testing.B) {
	x, _ := strconv.ParseFloat(b1, 64)
	y, _ := strconv.ParseFloat(b2, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x / y
	}
}

// --- 并发性能 ---

func BenchmarkFloat64_Parallel_Add(b *testing.B) {
	x, _ := strconv.ParseFloat(s1, 64)
	y, _ := strconv.ParseFloat(s2, 64)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = x + y
		}
	})
}
