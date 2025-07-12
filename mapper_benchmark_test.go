package mapper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// BenchmarkMapperCreation measures the performance of creating new mapper instances
func BenchmarkMapperCreation(b *testing.B) {
	b.ResetTimer()
	var m Mapper
	for i := 0; i < b.N; i++ {
		m = New()
	}
	b.StopTimer()
	assert.NotNil(b, m)
}

// BenchmarkMappingRegistration measures the performance of registering mapping functions
func BenchmarkMappingRegistration(b *testing.B) {
	mapper := New()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Register(mapper, stringToInt)
	}
	b.StopTimer()
	assert.True(b, Has[string, int](mapper), "Mapping should be registered")
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
	b.StopTimer()

	assert.True(b, Has[string, int](mapper), "Mapping string to int should be registered")
	assert.True(b, Has[int, string](mapper), "Mapping int to string should be registered")
	assert.True(b, Has[Person, PersonDTO](mapper), "Mapping Person to PersonDTO should be registered")
	assert.True(b, Has[EmptyStruct, EmptyStruct](mapper), "Mapping Empty to Empty should be registered")
}

// BenchmarkSimpleMapping measures the performance of mapping simple types (string to int)
func BenchmarkSimpleMapping(b *testing.B) {
	mapper := New()
	Register(mapper, stringToInt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Map[string, int](mapper, "benchmark")
	}
	b.StopTimer()
	assert.NotNil(b, mapper, "Mapper should not be nil")
	assert.True(b, Has[string, int](mapper), "Mapping string to int should be registered")
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
	b.StopTimer()
	assert.NotNil(b, mapper, "Mapper should not be nil")
	assert.True(b, Has[Person, PersonDTO](mapper), "Mapping Person to PersonDTO should be registered")
	assert.Equal(b, person.Name, "Benchmark", "Person name should match")
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
	b.StopTimer()
	assert.NotNil(b, mapper, "Mapper should not be nil")
	assert.True(b, Has[Person, PersonDTO](mapper), "Mapping []Person to []PersonDTO should be registered")
	assert.Len(b, persons, 1, "Slice should contain one element")
	assert.Equal(b, persons[0].Name, "Person0", "First element name should match")
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
	b.StopTimer()
	assert.NotNil(b, mapper, "Mapper should not be nil")
	assert.True(b, Has[Person, PersonDTO](mapper), "Mapping []Person to []PersonDTO should be registered")
	assert.Len(b, persons, 100, "Slice should contain 100 elements")
	assert.Equal(b, persons[0].Name, "Person0", "First element name should match")
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

	var result bool
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = Has[string, int](mapper)
	}
	b.StopTimer()
	assert.True(b, result, "Mapping string to int should be registered")
}

func BenchmarkHasMappingNotFound(b *testing.B) {
	mapper := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Has[string, int](mapper)
	}
	b.StopTimer()
	assert.False(b, Has[string, int](mapper), "Mapping string to int should not be registered")
}

func BenchmarkRemoveMapping(b *testing.B) {
	mapper := New()
	Register(mapper, stringToInt)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Remove[string, int](mapper)
	}
	b.StopTimer()
	assert.False(b, Has[string, int](mapper), "Mapping string to int should be removed")
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
	b.StopTimer()
	assert.NotEmpty(b, List(mapper), "List of mappings should not be empty")
}

func BenchmarkListLarge(b *testing.B) {
	mapper := New()

	Register(mapper, func(i int) string { return fmt.Sprintf("%d", i) })

	var result []string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = List(mapper)
	}
	b.StopTimer()

	assert.Len(b, result, 1, "List of mappings should contain 1000 elements")
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
		mappedInt, err1 := Map[string, int](mapper, "test")
		assert.NoError(b, err1, "Mapping string to int should not error")
		assert.Equal(b, mappedInt, stringToInt("test"), "Mapped int should match expected")

		mappedStr, err2 := Map[int, string](mapper, 42)
		assert.NoError(b, err2, "Mapping int to string should not error")
		assert.Equal(b, mappedStr, intToString(42), "Mapped string should match expected")

		mappedDTO, err3 := Map[Person, PersonDTO](mapper, Person{Name: "Test", Age: 30})
		assert.NoError(b, err3, "Mapping Person to PersonDTO should not error")
		assert.Equal(b, mappedDTO, personToDTO(Person{Name: "Test", Age: 30}), "Mapped DTO should match expected")

		// Check existence
		assert.True(b, Has[string, int](mapper), "Mapping string to int should exist")
		assert.True(b, Has[int, string](mapper), "Mapping int to string should exist")
		assert.True(b, Has[Person, PersonDTO](mapper), "Mapping Person to PersonDTO should exist")

		// List mappings
		mappings := List(mapper)
		assert.NotEmpty(b, mappings, "List of mappings should not be empty")
		assert.GreaterOrEqual(b, len(mappings), 3, "Should have at least 3 mappings")

		// Remove one mapping
		Remove[string, int](mapper)
		assert.False(b, Has[string, int](mapper), "Mapping string to int should be removed")
	}
}
