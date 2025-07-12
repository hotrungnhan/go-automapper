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

func BenchmarkMapSliceUnsafe(b *testing.B) {
	m := New()
	Register(m, func(s string) int { return len(s) })

	input := []string{"hello", "world", "go", "programming", "test"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := MapSliceUnsafe[[]string, []int](m, input)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkMapSliceOriginal(b *testing.B) {
	m := New()
	Register(m, func(s string) int { return len(s) })

	input := []string{"hello", "world", "go", "programming", "test"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := MapSlice[[]string, []int](m, input)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

func BenchmarkManualSliceMap(b *testing.B) {
	fn := func(s string) int { return len(s) }
	input := []string{"hello", "world", "go", "programming", "test"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := make([]int, len(input))
		for j, s := range input {
			result[j] = fn(s)
		}
		_ = result
	}
}

// Test larger slice performance
func BenchmarkMapSliceLarge(b *testing.B) {
	m := New()
	Register(m, func(s string) int { return len(s) })

	// Create a larger slice for more realistic performance testing
	input := make([]string, 1000)
	for i := range input {
		input[i] = "test string for benchmarking purposes"
	}

	b.Run("Unsafe", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result, err := MapSliceUnsafe[[]string, []int](m, input)
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		}
	})

	b.Run("Original", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result, err := MapSlice[[]string, []int](m, input)
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		}
	})

	b.Run("Manual", func(b *testing.B) {
		fn := func(s string) int { return len(s) }
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			result := make([]int, len(input))
			for j, s := range input {
				result[j] = fn(s)
			}
			_ = result
		}
	})
}
