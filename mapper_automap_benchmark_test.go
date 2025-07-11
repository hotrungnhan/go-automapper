package mapper

import (
	"testing"
)

// Benchmark tests for AutoMap and RegisterAutoMap functionality

// BenchmarkAutoMapSimpleStruct measures performance of automatic mapping between simple structs
func BenchmarkAutoMapSimpleStruct(b *testing.B) {
	type Source struct {
		Name  string
		Age   int
		Email string
	}
	type Dest struct {
		Name  string
		Age   int
		Email string
	}

	src := Source{Name: "John Doe", Age: 30, Email: "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = autoMap[Source, Dest](src)
	}
}

// BenchmarkAutoMapComplexStruct measures performance of automatic mapping between complex nested structs
func BenchmarkAutoMapComplexStruct(b *testing.B) {
	type Address struct {
		Street string
		City   string
		ZIP    string
	}
	type Contact struct {
		Email string
		Phone string
	}
	type Source struct {
		Name    string
		Age     int
		Address Address
		Contact Contact
		Tags    []string
		Scores  map[string]int
	}
	type Dest struct {
		Name    string
		Age     int
		Address Address
		Contact Contact
		Tags    []string
		Scores  map[string]int
	}

	src := Source{
		Name: "John Doe",
		Age:  30,
		Address: Address{
			Street: "123 Main Street",
			City:   "New York",
			ZIP:    "10001",
		},
		Contact: Contact{
			Email: "john@example.com",
			Phone: "555-1234",
		},
		Tags:   []string{"developer", "golang", "backend"},
		Scores: map[string]int{"math": 95, "science": 88, "english": 92},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = autoMap[Source, Dest](src)
	}
}

// BenchmarkAutoMapWithPointers measures performance of automatic mapping between structs with pointer fields
func BenchmarkAutoMapWithPointers(b *testing.B) {
	type Source struct {
		Name  *string
		Age   *int
		Email *string
	}
	type Dest struct {
		Name  *string
		Age   *int
		Email *string
	}

	name := "John Doe"
	age := 30
	email := "john@example.com"
	src := Source{Name: &name, Age: &age, Email: &email}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = autoMap[Source, Dest](src)
	}
}

func BenchmarkAutoMapPrimitive(b *testing.B) {
	src := "hello world"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = autoMap[string, string](src)
	}
}

// TODO: Improved this benchmark to only test RegisterAutoMap with under an knowned amount of registered mappings
func BenchmarkRegisterAutoMap(b *testing.B) {
	type Source struct {
		Name string
		Age  int
	}
	type Dest struct {
		Name string
		Age  int
	}
	mapper := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		RegisterAutoMap[Source, Dest](mapper)
	}
}

func BenchmarkMapWithAutoMap(b *testing.B) {
	type Source struct {
		Name  string
		Age   int
		Email string
	}
	type Dest struct {
		Name  string
		Age   int
		Email string
	}

	mapper := New()
	RegisterAutoMap[Source, Dest](mapper)
	src := Source{Name: "John Doe", Age: 30, Email: "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Map[Source, Dest](mapper, src)
	}
}

func BenchmarkMapWithAutoMapComplex(b *testing.B) {
	type Address struct {
		Street string
		City   string
	}
	type Source struct {
		Name    string
		Age     int
		Address Address
		Tags    []string
	}
	type Dest struct {
		Name    string
		Age     int
		Address Address
		Tags    []string
	}

	mapper := New()
	RegisterAutoMap[Source, Dest](mapper)
	src := Source{
		Name:    "John Doe",
		Age:     30,
		Address: Address{Street: "123 Main St", City: "NYC"},
		Tags:    []string{"dev", "go", "backend"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Map[Source, Dest](mapper, src)
	}
}

// Benchmark AutoMap with different struct sizes
func BenchmarkAutoMapStructSizes(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		type Small struct {
			A string
			B int
		}
		src := Small{A: "test", B: 42}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = autoMap[Small, Small](src)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		type Medium struct {
			A, B, C, D, E string
			F, G, H, I, J int
		}
		src := Medium{
			A: "a", B: "b", C: "c", D: "d", E: "e",
			F: 1, G: 2, H: 3, I: 4, J: 5,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = autoMap[Medium, Medium](src)
		}
	})

	b.Run("Large", func(b *testing.B) {
		type Large struct {
			A, B, C, D, E, F, G, H, I, J string
			K, L, M, N, O, P, Q, R, S, T int
			U, V, W, X, Y, Z             bool
		}
		src := Large{
			A: "a", B: "b", C: "c", D: "d", E: "e",
			F: "f", G: "g", H: "h", I: "i", J: "j",
			K: 1, L: 2, M: 3, N: 4, O: 5,
			P: 6, Q: 7, R: 8, S: 9, T: 10,
			U: true, V: false, W: true, X: false, Y: true, Z: false,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = autoMap[Large, Large](src)
		}
	})
}
