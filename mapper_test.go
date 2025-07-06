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
	t.Run("creates_new_mapper", func(t *testing.T) {
		mapper := NewMapper()

		if mapper.registry == nil {
			t.Error("Expected registry to be initialized")
		}

		if len(mapper.registry) != 0 {
			t.Errorf("Expected empty registry, got %d items", len(mapper.registry))
		}
	})

	t.Run("multiple_instances_are_independent", func(t *testing.T) {
		mapper1 := NewMapper()
		mapper2 := NewMapper()

		Register(mapper1, stringToInt)

		if HasMapping[string, int](mapper1) == HasMapping[string, int](mapper2) {
			t.Error("Expected mappers to be independent")
		}
	})
}

// TestRegister tests the registration functionality
func TestRegister(t *testing.T) {
	t.Run("register_simple_mapping", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)

		if !HasMapping[string, int](mapper) {
			t.Error("Expected mapping to be registered")
		}
	})

	t.Run("register_struct_mapping", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, personToDTO)

		if !HasMapping[Person, PersonDTO](mapper) {
			t.Error("Expected struct mapping to be registered")
		}
	})

	t.Run("register_empty_struct_mapping", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, emptyToEmpty)

		if !HasMapping[EmptyStruct, EmptyStruct](mapper) {
			t.Error("Expected empty struct mapping to be registered")
		}
	})

	t.Run("register_multiple_mappings", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Register(mapper, personToDTO)

		if !HasMapping[string, int](mapper) {
			t.Error("Expected string->int mapping")
		}
		if !HasMapping[int, string](mapper) {
			t.Error("Expected int->string mapping")
		}
		if !HasMapping[Person, PersonDTO](mapper) {
			t.Error("Expected Person->PersonDTO mapping")
		}
	})

	t.Run("register_same_mapping_twice_overwrites", func(t *testing.T) {
		mapper := NewMapper()

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

	t.Run("register_bidirectional_mappings", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)
		Register(mapper, intToString)

		// Should be able to map both ways
		if !HasMapping[string, int](mapper) {
			t.Error("Expected string->int mapping")
		}
		if !HasMapping[int, string](mapper) {
			t.Error("Expected int->string mapping")
		}
	})
}

// TestMap tests the mapping functionality
func TestMap(t *testing.T) {
	t.Run("map_registered_function", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)

		result, err := Map[string, int](mapper, "hello")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != 5 {
			t.Errorf("Expected 5, got %d", result)
		}
	})

	t.Run("map_struct_conversion", func(t *testing.T) {
		mapper := NewMapper()
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

	t.Run("map_unregistered_function_returns_error", func(t *testing.T) {
		mapper := NewMapper()

		_, err := Map[string, int](mapper, "test")
		if err == nil {
			t.Error("Expected error for unregistered mapping")
		}

		expectedError := "no mapping function registered"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
		}
	})

	t.Run("map_zero_values", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)

		result, err := Map[string, int](mapper, "zero")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})

	t.Run("map_empty_string", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)

		result, err := Map[string, int](mapper, "")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})

	t.Run("map_empty_struct", func(t *testing.T) {
		mapper := NewMapper()
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

	t.Run("map_returns_zero_value_on_error", func(t *testing.T) {
		mapper := NewMapper()

		result, err := Map[string, int](mapper, "test")
		if err == nil {
			t.Error("Expected error")
		}
		if result != 0 {
			t.Errorf("Expected zero value (0), got %d", result)
		}
	})

	t.Run("map_pointer_types", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, func(s *string) *int {
			if s == nil {
				return nil
			}
			length := len(*s)
			return &length
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
}

// TestHasMapping tests the mapping existence check
func TestHasMapping(t *testing.T) {
	t.Run("has_mapping_returns_true_for_registered", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)

		if !HasMapping[string, int](mapper) {
			t.Error("Expected HasMapping to return true")
		}
	})

	t.Run("has_mapping_returns_false_for_unregistered", func(t *testing.T) {
		mapper := NewMapper()

		if HasMapping[string, int](mapper) {
			t.Error("Expected HasMapping to return false")
		}
	})

	t.Run("has_mapping_different_type_combinations", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)

		// Registered mapping
		if !HasMapping[string, int](mapper) {
			t.Error("Expected string->int mapping")
		}

		// Reverse mapping not registered
		if HasMapping[int, string](mapper) {
			t.Error("Expected int->string mapping to not exist")
		}

		// Completely different types
		if HasMapping[Person, PersonDTO](mapper) {
			t.Error("Expected Person->PersonDTO mapping to not exist")
		}
	})

	t.Run("has_mapping_empty_struct", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, emptyToEmpty)

		if !HasMapping[EmptyStruct, EmptyStruct](mapper) {
			t.Error("Expected EmptyStruct->EmptyStruct mapping")
		}
	})

	t.Run("has_mapping_after_removal", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)

		if !HasMapping[string, int](mapper) {
			t.Error("Expected mapping before removal")
		}

		RemoveMapping[string, int](mapper)

		if HasMapping[string, int](mapper) {
			t.Error("Expected mapping to not exist after removal")
		}
	})
}

// TestRemoveMapping tests the mapping removal functionality
func TestRemoveMapping(t *testing.T) {
	t.Run("remove_existing_mapping", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)

		if !HasMapping[string, int](mapper) {
			t.Error("Expected mapping to exist before removal")
		}

		RemoveMapping[string, int](mapper)

		if HasMapping[string, int](mapper) {
			t.Error("Expected mapping to be removed")
		}
	})

	t.Run("remove_nonexistent_mapping_no_panic", func(t *testing.T) {
		mapper := NewMapper()

		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RemoveMapping panicked: %v", r)
			}
		}()

		RemoveMapping[string, int](mapper)
	})

	t.Run("remove_one_of_multiple_mappings", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Register(mapper, personToDTO)

		RemoveMapping[string, int](mapper)

		if HasMapping[string, int](mapper) {
			t.Error("Expected string->int mapping to be removed")
		}
		if !HasMapping[int, string](mapper) {
			t.Error("Expected int->string mapping to remain")
		}
		if !HasMapping[Person, PersonDTO](mapper) {
			t.Error("Expected Person->PersonDTO mapping to remain")
		}
	})

	t.Run("remove_and_re_register", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)

		RemoveMapping[string, int](mapper)
		Register(mapper, stringToInt)

		if !HasMapping[string, int](mapper) {
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
	t.Run("list_empty_mappings", func(t *testing.T) {
		mapper := NewMapper()
		mappings := List(mapper)

		if len(mappings) != 0 {
			t.Errorf("Expected 0 mappings, got %d", len(mappings))
		}
	})

	t.Run("list_single_mapping", func(t *testing.T) {
		mapper := NewMapper()
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

	t.Run("list_multiple_mappings", func(t *testing.T) {
		mapper := NewMapper()
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

	t.Run("list_mappings_after_removal", func(t *testing.T) {
		mapper := NewMapper()
		Register(mapper, stringToInt)
		Register(mapper, intToString)

		RemoveMapping[string, int](mapper)
		mappings := List(mapper)

		if len(mappings) != 1 {
			t.Errorf("Expected 1 mapping after removal, got %d", len(mappings))
		}

		expected := "int-string"
		if mappings[0] != expected {
			t.Errorf("Expected '%s', got '%s'", expected, mappings[0])
		}
	})

	t.Run("list_mappings_returns_copy", func(t *testing.T) {
		mapper := NewMapper()
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
