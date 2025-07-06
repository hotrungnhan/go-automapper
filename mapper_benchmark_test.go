package mapper

import (
	"fmt"
	"testing"
)

// Benchmark tests
func BenchmarkNewMapper(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMapper()
	}
}

// TODO: Improved this benchmark to only test RegisterAutoMap with under an knowned amount of registered mappings
func BenchmarkRegister(b *testing.B) {
	mapper := NewMapper()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Register(mapper, stringToInt)
	}
}

// TODO: Improved this benchmark to only test RegisterAutoMap with under an knowned amount of registered mappings
func BenchmarkRegisterDifferentTypes(b *testing.B) {
	mapper := NewMapper()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Register(mapper, personToDTO)
		Register(mapper, emptyToEmpty)
	}

}

func BenchmarkMap(b *testing.B) {
	mapper := NewMapper()
	Register(mapper, stringToInt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Map[string, int](mapper, "benchmark")
	}
}

func BenchmarkMapStruct(b *testing.B) {
	mapper := NewMapper()
	Register(mapper, personToDTO)
	person := Person{Name: "Benchmark", Age: 25}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Map[Person, PersonDTO](mapper, person)
	}
}

func BenchmarkMapNotFound(b *testing.B) {
	mapper := NewMapper()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Map[string, int](mapper, "test")
	}
}

func BenchmarkHasMapping(b *testing.B) {
	mapper := NewMapper()
	Register(mapper, stringToInt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HasMapping[string, int](mapper)
	}
}

func BenchmarkHasMappingNotFound(b *testing.B) {
	mapper := NewMapper()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HasMapping[string, int](mapper)
	}
}

func BenchmarkRemoveMapping(b *testing.B) {
	mapper := NewMapper()
	Register(mapper, stringToInt)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RemoveMapping[string, int](mapper)
	}
}

func BenchmarkList(b *testing.B) {
	mapper := NewMapper()
	Register(mapper, stringToInt)
	Register(mapper, intToString)
	Register(mapper, personToDTO)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = List(mapper)
	}
}

func BenchmarkListLarge(b *testing.B) {
	mapper := NewMapper()

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
		mapper := NewMapper()

		// Register multiple mappings
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Register(mapper, personToDTO)

		// Perform mappings
		_, _ = Map[string, int](mapper, "test")
		_, _ = Map[int, string](mapper, 42)
		_, _ = Map[Person, PersonDTO](mapper, Person{Name: "Test", Age: 30})

		// Check existence
		_ = HasMapping[string, int](mapper)
		_ = HasMapping[int, string](mapper)

		// List mappings
		_ = List(mapper)

		// Remove one mapping
		RemoveMapping[string, int](mapper)
	}
}

func BenchmarkMapperVsManual(b *testing.B) {
	mapper := NewMapper()
	Register(mapper, stringToInt)

	b.Run("Mapper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Map[string, int](mapper, "benchmark")
		}
	})

	b.Run("Manual", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = stringToInt("benchmark")
		}
	})
}
