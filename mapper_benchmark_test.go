package mapper

import (
	"fmt"
	"testing"
)

// BenchmarkMapperCreation measures the performance of creating new mapper instances
func BenchmarkMapperCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}

// BenchmarkMappingRegistration measures the performance of registering mapping functions
func BenchmarkMappingRegistration(b *testing.B) {
	mapper := New()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Register(mapper, stringToInt)
	}
}

// BenchmarkMultipleTypeRegistration measures the performance of registering different mapping types
func BenchmarkMultipleTypeRegistration(b *testing.B) {
	mapper := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Register(mapper, personToDTO)
		Register(mapper, emptyToEmpty)
	}

}

// BenchmarkSimpleMapping measures the performance of mapping simple types (string to int)
func BenchmarkSimpleMapping(b *testing.B) {
	mapper := New()
	Register(mapper, stringToInt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Map[string, int](mapper, "benchmark")
	}
}

// BenchmarkStructMapping measures the performance of mapping between struct types
func BenchmarkStructMapping(b *testing.B) {
	mapper := New()
	Register(mapper, personToDTO)
	person := Person{Name: "Benchmark", Age: 25}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Map[Person, PersonDTO](mapper, person)
	}
}

// BenchmarkSingleElementSliceMapping measures the performance of mapping a slice with one element
func BenchmarkSingleElementSliceMapping(b *testing.B) {
	mapper := New()
	Register(mapper, personToDTO)

	// Prepare a slice of Person
	persons := make([]Person, 1)
	for i := 0; i < 1; i++ {
		persons[i] = Person{Name: fmt.Sprintf("Person%d", i), Age: 20 + i}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MapSlice[[]Person, []PersonDTO](mapper, persons)
	}
}

// BenchmarkLargeSliceMapping measures the performance of mapping a slice with 100 elements
func BenchmarkLargeSliceMapping(b *testing.B) {
	mapper := New()
	Register(mapper, personToDTO)

	// Prepare a slice of Person
	persons := make([]Person, 100)
	for i := 0; i < len(persons); i++ {
		persons[i] = Person{Name: fmt.Sprintf("Person%d", i), Age: 20 + i}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MapSlice[[]Person, []PersonDTO](mapper, persons)
	}
}

// BenchmarkMappingNotFound measures the performance when no mapping function is registered
func BenchmarkMappingNotFound(b *testing.B) {
	mapper := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Map[string, int](mapper, "test")
	}
}

func BenchmarkHasMapping(b *testing.B) {
	mapper := New()
	Register(mapper, stringToInt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Has[string, int](mapper)
	}
}

func BenchmarkHasMappingNotFound(b *testing.B) {
	mapper := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Has[string, int](mapper)
	}
}

func BenchmarkRemoveMapping(b *testing.B) {
	mapper := New()
	Register(mapper, stringToInt)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Remove[string, int](mapper)
	}
}

func BenchmarkList(b *testing.B) {
	mapper := New()
	Register(mapper, stringToInt)
	Register(mapper, intToString)
	Register(mapper, personToDTO)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = List(mapper)
	}
}

func BenchmarkListLarge(b *testing.B) {
	mapper := New()

	// Register many mappings
	for i := 0; i < 1000; i++ {
		Register(mapper, func(i int) string { return fmt.Sprintf("%d", i) })
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = List(mapper)
	}
}

// Complex scenario benchmarks
func BenchmarkComplexWorkflow(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapper := New()

		// Register multiple mappings
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Register(mapper, personToDTO)

		// Perform mappings
		_, _ = Map[string, int](mapper, "test")
		_, _ = Map[int, string](mapper, 42)
		_, _ = Map[Person, PersonDTO](mapper, Person{Name: "Test", Age: 30})

		// Check existence
		_ = Has[string, int](mapper)
		_ = Has[int, string](mapper)

		// List mappings
		_ = List(mapper)

		// Remove one mapping
		Remove[string, int](mapper)
	}
}
