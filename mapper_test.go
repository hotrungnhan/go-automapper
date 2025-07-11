package mapper

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Test types for various scenarios
type (
	Person struct {
		Name string
		Age  int
	}

	PersonDTO struct {
		FullName string
		Years    int
	}

	EmptyStruct struct{}

	NestedStruct struct {
		Inner Person
		Value string
	}

	GenericStruct[T any] struct {
		Data T
	}
)

// Helper mapping functions
func personToDTO(p Person) PersonDTO {
	return PersonDTO{
		FullName: p.Name,
		Years:    p.Age,
	}
}

func stringToInt(s string) int {
	if s == "zero" {
		return 0
	}
	return len(s)
}

func intToString(i int) string {
	return fmt.Sprintf("number-%d", i)
}

func emptyToEmpty(e EmptyStruct) EmptyStruct {
	return e
}

// TestNewMapper tests the constructor
func TestNewMapper(t *testing.T) {
	t.Run("CreatesNewMapperWithInitializedRegistry", func(t *testing.T) {
		mapper := New()

		if mapper.registry == nil {
			t.Error("Expected registry to be initialized")
		}

		if len(mapper.registry) != 0 {
			t.Errorf("Expected empty registry, got %d items", len(mapper.registry))
		}
	})

	t.Run("MultipleInstancesAreIndependent", func(t *testing.T) {
		mapper1 := New()
		mapper2 := New()

		Register(mapper1, stringToInt)

		if Has[string, int](mapper1) == Has[string, int](mapper2) {
			t.Error("Expected mappers to be independent")
		}
	})
}

// TestRegister tests the registration functionality
func TestRegister(t *testing.T) {
	t.Run("RegisterSimpleMapping", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		if !Has[string, int](mapper) {
			t.Error("Expected mapping to be registered")
		}
	})

	t.Run("RegisterStructMapping", func(t *testing.T) {
		mapper := New()
		Register(mapper, personToDTO)

		if !Has[Person, PersonDTO](mapper) {
			t.Error("Expected struct mapping to be registered")
		}
	})

	t.Run("RegisterEmptyStructMapping", func(t *testing.T) {
		mapper := New()
		Register(mapper, emptyToEmpty)

		if !Has[EmptyStruct, EmptyStruct](mapper) {
			t.Error("Expected empty struct mapping to be registered")
		}
	})

	t.Run("RegisterMultipleMappings", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Register(mapper, personToDTO)

		if !Has[string, int](mapper) {
			t.Error("Expected string->int mapping")
		}
		if !Has[int, string](mapper) {
			t.Error("Expected int->string mapping")
		}
		if !Has[Person, PersonDTO](mapper) {
			t.Error("Expected Person->PersonDTO mapping")
		}
	})

	t.Run("RegisterSameMappingTwiceOverwrites", func(t *testing.T) {
		mapper := New()

		// Register first function
		Register(mapper, func(s string) int { return 1 })
		result1, _ := Map[string, int](mapper, "test")

		// Register second function with same signature
		Register(mapper, func(s string) int { return 2 })
		result2, _ := Map[string, int](mapper, "test")

		if result1 == result2 {
			t.Error("Expected second registration to overwrite first")
		}
		if result2 != 2 {
			t.Errorf("Expected result 2, got %d", result2)
		}
	})

	t.Run("RegisterBidirectionalMappings", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		Register(mapper, intToString)

		// Should be able to map both ways
		if !Has[string, int](mapper) {
			t.Error("Expected string->int mapping")
		}
		if !Has[int, string](mapper) {
			t.Error("Expected int->string mapping")
		}
	})
}

// TestMap tests the mapping functionality
func TestMap(t *testing.T) {
	t.Run("MapRegisteredFunction", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		result, err := Map[string, int](mapper, "hello")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != 5 {
			t.Errorf("Expected 5, got %d", result)
		}
	})

	t.Run("MapUnregisteredFunctionReturnsError", func(t *testing.T) {
		mapper := New()

		_, err := Map[string, int](mapper, "test")
		if err == nil {
			t.Error("Expected error for unregistered mapping")
		}

		expectedError := "no mapping function registered"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
		}
	})

	t.Run("MapZeroValues", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		result, err := Map[string, int](mapper, "zero")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})
	t.Run("MapNilInput", func(t *testing.T) {
		mapper := New()
		Register(mapper, func(p Person) PersonDTO {
			return PersonDTO{
				FullName: p.Name,
				Years:    p.Age,
			}
		})

		result, err := Map[*Person, PersonDTO](mapper, nil)

		if err != nil {
			t.Errorf("Unexpected error for nil input: %v", err)
		}
		if result != (PersonDTO{}) {
			t.Errorf("Expected zero struct, got %+v", result)
		}
	})

	t.Run("MapNilPointerToPointerReturnsNil", func(t *testing.T) {
		mapper := New()
		Register(mapper, func(p Person) PersonDTO {
			return PersonDTO{
				FullName: p.Name,
				Years:    p.Age,
			}
		})

		result, err := Map[*Person, *PersonDTO](mapper, nil)
		if err != nil {
			t.Errorf("Unexpected error for nil input: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result for nil input, got %+v", result)
		}
	})

	t.Run("MapEmptyString", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		result, err := Map[string, int](mapper, "")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})

	t.Run("MapEmptyStruct", func(t *testing.T) {
		mapper := New()
		Register(mapper, emptyToEmpty)

		empty := EmptyStruct{}
		result, err := Map[EmptyStruct, EmptyStruct](mapper, empty)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !reflect.DeepEqual(result, empty) {
			t.Error("Expected empty struct to map to itself")
		}
	})

	t.Run("MapReturnsZeroValueOnError", func(t *testing.T) {
		mapper := New()

		result, err := Map[string, int](mapper, "test")
		if err == nil {
			t.Error("Expected error")
		}
		if result != 0 {
			t.Errorf("Expected zero value (0), got %d", result)
		}
	})

	t.Run("MapStructToStruct", func(t *testing.T) {
		mapper := New()
		Register(mapper, personToDTO)

		person := Person{Name: "John", Age: 30}
		result, err := Map[Person, PersonDTO](mapper, person)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.FullName != "John" || result.Years != 30 {
			t.Errorf("Expected {John 30}, got %+v", result)
		}
	})
	t.Run("MapPointerTypes", func(t *testing.T) {
		mapper := New()
		Register(mapper, func(s string) int {
			length := len(s)
			return length
		})

		str := "hello"
		result, err := Map[*string, *int](mapper, &str)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil || *result != 5 {
			t.Errorf("Expected pointer to 5, got %v", result)
		}
	})

	t.Run("MapPointerToPointer", func(t *testing.T) {
		mapper := New()
		Register(mapper, func(p Person) PersonDTO {
			return PersonDTO{
				FullName: p.Name,
				Years:    p.Age,
			}
		})

		person := &Person{Name: "Alice", Age: 25}
		result, err := Map[*Person, *PersonDTO](mapper, person)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil || result.FullName != "Alice" || result.Years != 25 {
			t.Errorf("Expected &{Alice 25}, got %+v", result)
		}

		// Test nil pointer
		result, err = Map[*Person, *PersonDTO](mapper, nil)
		if err != nil {
			t.Errorf("Unexpected error for nil input: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result for nil input, got %+v", result)
		}
	})

	t.Run("MapPointerToStruct", func(t *testing.T) {
		mapper := New()
		Register(mapper, func(p Person) PersonDTO {
			return PersonDTO{
				FullName: p.Name,
				Years:    p.Age,
			}
		})

		person := &Person{Name: "Bob", Age: 40}
		result, err := Map[*Person, PersonDTO](mapper, person)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.FullName != "Bob" || result.Years != 40 {
			t.Errorf("Expected {Bob 40}, got %+v", result)
		}

		// Test nil pointer
		result, err = Map[*Person, PersonDTO](mapper, nil)
		if err != nil {
			t.Errorf("Unexpected error for nil input: %v", err)
		}
		if result != (PersonDTO{}) {
			t.Errorf("Expected zero value for nil input, got %+v", result)
		}
	})

	t.Run("MapStructToPointer", func(t *testing.T) {
		mapper := New()
		Register(mapper, func(p Person) PersonDTO {
			return PersonDTO{
				FullName: p.Name,
				Years:    p.Age,
			}
		})

		person := Person{Name: "Eve", Age: 22}
		result, err := Map[Person, *PersonDTO](mapper, person)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil || result.FullName != "Eve" || result.Years != 22 {
			t.Errorf("Expected &{Eve 22}, got %+v", result)
		}
	})
}

// TestMapSlice tests the MapSlice functionality for all 4 pointer/value slice mapping cases
func TestMapSlice(t *testing.T) {
	type A struct{ V int }
	type B struct{ W int }

	// Mapping function: value-to-value
	aToB := func(a A) B { return B{W: a.V + 1} }

	t.Run("SliceValueToValue", func(t *testing.T) {
		mapper := New()
		Register(mapper, aToB)

		src := []A{{1}, {2}, {3}}
		got, err := MapSlice[[]A, []B](mapper, src)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []B{{2}, {3}, {4}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v, got %v", want, got)
		}
	})

	t.Run("SlicePointerToPointer", func(t *testing.T) {
		mapper := New()
		Register(mapper, aToB)

		src := []*A{{1}, nil, {3}}
		got, err := MapSlice[[]*A, []*B](mapper, src)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []*B{{2}, nil, {4}}
		for i := range want {
			if src[i] == nil {
				if got[i] != nil {
					t.Errorf("expected nil at %d, got %v", i, got[i])
				}
			} else {
				if got[i] == nil || got[i].W != want[i].W {
					t.Errorf("expected %v at %d, got %v", want[i], i, got[i])
				}
			}
		}
	})

	t.Run("SlicePointerToValue", func(t *testing.T) {
		mapper := New()
		Register(mapper, aToB)

		src := []*A{{10}, nil, {30}}
		got, err := MapSlice[[]*A, []B](mapper, src)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []B{{11}, {}, {31}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v, got %v", want, got)
		}
	})

	t.Run("SliceValueToPointer", func(t *testing.T) {
		mapper := New()
		Register(mapper, aToB)

		src := []A{{100}, {200}}
		got, err := MapSlice[[]A, []*B](mapper, src)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(src) {
			t.Fatalf("expected length %d, got %d", len(src), len(got))
		}
		for i, v := range src {
			if got[i] == nil || got[i].W != v.V+1 {
				t.Errorf("expected pointer to {W:%d} at %d, got %v", v.V+1, i, got[i])
			}
		}
	})
}


// TestHasMapping tests the mapping existence check
func TestHasMapping(t *testing.T) {
	t.Run("HasMappingReturnsTrueForRegistered", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		if !Has[string, int](mapper) {
			t.Error("Expected HasMapping to return true")
		}
	})

	t.Run("HasMappingReturnsFalseForUnregistered", func(t *testing.T) {
		mapper := New()

		if Has[string, int](mapper) {
			t.Error("Expected HasMapping to return false")
		}
	})

	t.Run("HasMappingDifferentTypeCombinations", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		// Registered mapping
		if !Has[string, int](mapper) {
			t.Error("Expected string->int mapping")
		}

		// Reverse mapping not registered
		if Has[int, string](mapper) {
			t.Error("Expected int->string mapping to not exist")
		}

		// Completely different types
		if Has[Person, PersonDTO](mapper) {
			t.Error("Expected Person->PersonDTO mapping to not exist")
		}
	})

	t.Run("HasMappingEmptyStruct", func(t *testing.T) {
		mapper := New()
		Register(mapper, emptyToEmpty)

		if !Has[EmptyStruct, EmptyStruct](mapper) {
			t.Error("Expected EmptyStruct->EmptyStruct mapping")
		}
	})

	t.Run("HasMappingAfterRemoval", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		if !Has[string, int](mapper) {
			t.Error("Expected mapping before removal")
		}

		Remove[string, int](mapper)

		if Has[string, int](mapper) {
			t.Error("Expected mapping to not exist after removal")
		}
	})
}

// TestRemoveMapping tests the mapping removal functionality
func TestRemoveMapping(t *testing.T) {
	t.Run("RemoveExistingMapping", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		if !Has[string, int](mapper) {
			t.Error("Expected mapping to exist before removal")
		}

		Remove[string, int](mapper)

		if Has[string, int](mapper) {
			t.Error("Expected mapping to be removed")
		}
	})

	t.Run("RemoveNonexistentMappingNoPanic", func(t *testing.T) {
		mapper := New()

		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RemoveMapping panicked: %v", r)
			}
		}()

		Remove[string, int](mapper)
	})

	t.Run("RemoveOneOfMultipleMappings", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Register(mapper, personToDTO)

		Remove[string, int](mapper)

		if Has[string, int](mapper) {
			t.Error("Expected string->int mapping to be removed")
		}
		if !Has[int, string](mapper) {
			t.Error("Expected int->string mapping to remain")
		}
		if !Has[Person, PersonDTO](mapper) {
			t.Error("Expected Person->PersonDTO mapping to remain")
		}
	})

	t.Run("RemoveAndReRegister", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		Remove[string, int](mapper)
		Register(mapper, stringToInt)

		if !Has[string, int](mapper) {
			t.Error("Expected mapping to exist after re-registration")
		}

		result, err := Map[string, int](mapper, "test")
		if err != nil {
			t.Errorf("Unexpected error after re-registration: %v", err)
		}
		if result != 4 {
			t.Errorf("Expected 4, got %d", result)
		}
	})
}

// TestList tests the mapping listing functionality
func TestList(t *testing.T) {
	t.Run("ListEmptyMappings", func(t *testing.T) {
		mapper := New()
		mappings := List(mapper)

		if len(mappings) != 0 {
			t.Errorf("Expected 0 mappings, got %d", len(mappings))
		}
	})

	t.Run("ListSingleMapping", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		mappings := List(mapper)

		if len(mappings) != 1 {
			t.Errorf("Expected 1 mapping, got %d", len(mappings))
		}

		expected := "string-int"
		if mappings[0] != expected {
			t.Errorf("Expected '%s', got '%s'", expected, mappings[0])
		}
	})

	t.Run("ListMultipleMappings", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Register(mapper, personToDTO)

		mappings := List(mapper)
		sort.Strings(mappings)

		expected := []string{
			"int-string",
			"mapper.Person-mapper.PersonDTO",
			"string-int",
		}
		sort.Strings(expected)

		if len(mappings) != len(expected) {
			t.Errorf("Expected %d mappings, got %d", len(expected), len(mappings))
		}

		for i, mapping := range mappings {
			if mapping != expected[i] {
				t.Errorf("Expected '%s', got '%s'", expected[i], mapping)
			}
		}
	})

	t.Run("ListMappingsAfterRemoval", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		Register(mapper, intToString)

		Remove[string, int](mapper)
		mappings := List(mapper)

		if len(mappings) != 1 {
			t.Errorf("Expected 1 mapping after removal, got %d", len(mappings))
		}

		expected := "int-string"
		if mappings[0] != expected {
			t.Errorf("Expected '%s', got '%s'", expected, mappings[0])
		}
	})

	t.Run("ListMappingsReturnsCopy", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		mappings1 := List(mapper)
		mappings2 := List(mapper)

		// Modify one slice
		if len(mappings1) > 0 {
			mappings1[0] = "modified"
		}

		// Other slice should be unchanged
		if len(mappings2) > 0 && mappings2[0] == "modified" {
			t.Error("Expected List to return independent copies")
		}
	})
}
