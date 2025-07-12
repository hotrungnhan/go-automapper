package mapper

import (
	"testing"
)

func BenchmarkMapUnsafe(b *testing.B) {
	m := New()
	Register(m, func(s string) int { return len(s) })

	input := "hello world"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := MapUnsafe[string, int](m, input)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkMapOriginal(b *testing.B) {
	m := New()
	Register(m, func(s string) int { return len(s) })

	input := "hello world"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := Map[string, int](m, input)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkDirectFunctionCall(b *testing.B) {
	fn := func(s string) int { return len(s) }
	input := "hello world"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := fn(input)
		_ = result
	}
}

// Test the fast path in our optimized Map function
func BenchmarkMapOptimizedFastPath(b *testing.B) {
	m := New()
	// Register exact type that will hit the fast path
	Register[string, int](m, func(s string) int { return len(s) })

	input := "hello world"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := Map[string, int](m, input)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}
