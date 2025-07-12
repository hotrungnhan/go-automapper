package mapper

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestNewMapper(t *testing.T) {
	t.Run("CreatesNewMapperWithInitializedRegistry", func(t *testing.T) {
		mapper := New()
		assert.NotNil(t, mapper.registry, "Expected registry to be initialized")
		assert.Empty(t, mapper.registry, "Expected empty registry")
	})

	t.Run("MultipleInstancesAreIndependent", func(t *testing.T) {
		mapper1 := New()
		mapper2 := New()

		Register(mapper1, stringToInt)

		assert.NotEqual(t, Has[string, int](mapper1), Has[string, int](mapper2), "Expected mappers to be independent")
	})
}

func TestRegister(t *testing.T) {
	t.Run("RegisterSimpleMapping", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		assert.True(t, Has[string, int](mapper), "Expected mapping to be registered")
	})

	t.Run("RegisterStructMapping", func(t *testing.T) {
		mapper := New()
		Register(mapper, personToDTO)
		assert.True(t, Has[Person, PersonDTO](mapper), "Expected struct mapping to be registered")
	})

	t.Run("RegisterEmptyStructMapping", func(t *testing.T) {
		mapper := New()
		Register(mapper, emptyToEmpty)
		assert.True(t, Has[EmptyStruct, EmptyStruct](mapper), "Expected empty struct mapping to be registered")
	})

	t.Run("RegisterMultipleMappings", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Register(mapper, personToDTO)

		assert.True(t, Has[string, int](mapper), "Expected string->int mapping")
		assert.True(t, Has[int, string](mapper), "Expected int->string mapping")
		assert.True(t, Has[Person, PersonDTO](mapper), "Expected Person->PersonDTO mapping")
	})

	t.Run("RegisterSameMappingTwiceOverwrites", func(t *testing.T) {
		mapper := New()
		Register(mapper, func(s string) int { return 1 })
		result1, _ := Map[string, int](mapper, "test")
		Register(mapper, func(s string) int { return 2 })
		result2, _ := Map[string, int](mapper, "test")

		assert.NotEqual(t, result1, result2, "Expected second registration to overwrite first")
		assert.Equal(t, 2, result2, "Expected result 2")
	})

	t.Run("RegisterBidirectionalMappings", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		Register(mapper, intToString)

		assert.True(t, Has[string, int](mapper), "Expected string->int mapping")
		assert.True(t, Has[int, string](mapper), "Expected int->string mapping")
	})
}

func TestMap(t *testing.T) {
	t.Run("MapRegisteredFunction", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)

		result, err := Map[string, int](mapper, "hello")
		assert.NoError(t, err)
		assert.Equal(t, 5, result)
	})

	t.Run("MapUnregisteredFunctionReturnsError", func(t *testing.T) {
		mapper := New()
		_, err := Map[string, int](mapper, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no mapping function registered")
	})

	t.Run("MapZeroValues", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		result, err := Map[string, int](mapper, "zero")
		assert.NoError(t, err)
		assert.Equal(t, 0, result)
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
		assert.NoError(t, err)
		assert.Equal(t, PersonDTO{}, result)
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
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("MapEmptyString", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		result, err := Map[string, int](mapper, "")
		assert.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("MapEmptyStruct", func(t *testing.T) {
		mapper := New()
		Register(mapper, emptyToEmpty)
		empty := EmptyStruct{}
		result, err := Map[EmptyStruct, EmptyStruct](mapper, empty)
		assert.NoError(t, err)
		assert.Equal(t, empty, result)
	})

	t.Run("MapReturnsZeroValueOnError", func(t *testing.T) {
		mapper := New()
		result, err := Map[string, int](mapper, "test")
		assert.Error(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("MapStructToStruct", func(t *testing.T) {
		mapper := New()
		Register(mapper, personToDTO)
		person := Person{Name: "John", Age: 30}
		result, err := Map[Person, PersonDTO](mapper, person)
		assert.NoError(t, err)
		assert.Equal(t, PersonDTO{FullName: "John", Years: 30}, result)
	})

	t.Run("MapPointerTypes", func(t *testing.T) {
		mapper := New()
		Register(mapper, func(s string) int {
			return len(s)
		})

		str := "hello"
		result, err := Map[*string, *int](mapper, &str)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 5, *result)
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
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, &PersonDTO{FullName: "Alice", Years: 25}, result)

		// Test nil pointer
		result, err = Map[*Person, *PersonDTO](mapper, nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
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
		assert.NoError(t, err)
		assert.Equal(t, PersonDTO{FullName: "Bob", Years: 40}, result)

		// Test nil pointer
		result, err = Map[*Person, PersonDTO](mapper, nil)
		assert.NoError(t, err)
		assert.Equal(t, PersonDTO{}, result)
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
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, &PersonDTO{FullName: "Eve", Years: 22}, result)
	})
}

// TestMapSlice tests the MapSlice functionality for all 4 pointer/value slice mapping cases
func TestMapSlice(t *testing.T) {
	type A struct{ V int }
	type B struct{ W int }

	aToB := func(a A) B { return B{W: a.V + 1} }

	t.Run("SliceValueToValue", func(t *testing.T) {
		mapper := New()
		Register(mapper, aToB)

		src := []A{{1}, {2}, {3}}
		got, err := MapSlice[[]A, []B](mapper, src)
		assert.NoError(t, err)
		want := []B{{2}, {3}, {4}}
		assert.Equal(t, want, got)
	})

	t.Run("SlicePointerToPointer", func(t *testing.T) {
		mapper := New()
		Register(mapper, aToB)

		src := []*A{{1}, nil, {3}}
		got, err := MapSlice[[]*A, []*B](mapper, src)
		assert.NoError(t, err)
		want := []*B{{2}, nil, {4}}
		for i := range want {
			if src[i] == nil {
				assert.Nil(t, got[i], "expected nil at %d", i)
			} else {
				assert.NotNil(t, got[i], "expected non-nil at %d", i)
				assert.Equal(t, want[i].W, got[i].W, "expected %v at %d, got %v", want[i], i, got[i])
			}
		}
	})

	t.Run("SlicePointerToValue", func(t *testing.T) {
		mapper := New()
		Register(mapper, aToB)

		src := []*A{{10}, nil, {30}}
		got, err := MapSlice[[]*A, []B](mapper, src)
		assert.NoError(t, err)
		want := []B{{11}, {}, {31}}
		assert.Equal(t, want, got)
	})

	t.Run("SliceValueToPointer", func(t *testing.T) {
		mapper := New()
		Register(mapper, aToB)

		src := []A{{100}, {200}}
		got, err := MapSlice[[]A, []*B](mapper, src)
		assert.NoError(t, err)
		assert.Len(t, got, len(src))
		for i, v := range src {
			assert.NotNil(t, got[i], "expected pointer at %d", i)
			assert.Equal(t, v.V+1, got[i].W, "expected pointer to {W:%d} at %d, got %v", v.V+1, i, got[i])
		}
	})
}

// TestHasMapping tests the mapping existence check
func TestHasMapping(t *testing.T) {
	t.Run("HasMappingReturnsTrueForRegistered", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		assert.True(t, Has[string, int](mapper), "Expected HasMapping to return true")
	})

	t.Run("HasMappingReturnsFalseForUnregistered", func(t *testing.T) {
		mapper := New()
		assert.False(t, Has[string, int](mapper), "Expected HasMapping to return false")
	})

	t.Run("HasMappingDifferentTypeCombinations", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		assert.True(t, Has[string, int](mapper), "Expected string->int mapping")
		assert.False(t, Has[int, string](mapper), "Expected int->string mapping to not exist")
		assert.False(t, Has[Person, PersonDTO](mapper), "Expected Person->PersonDTO mapping to not exist")
	})

	t.Run("HasMappingEmptyStruct", func(t *testing.T) {
		mapper := New()
		Register(mapper, emptyToEmpty)
		assert.True(t, Has[EmptyStruct, EmptyStruct](mapper), "Expected EmptyStruct->EmptyStruct mapping")
	})

	t.Run("HasMappingAfterRemoval", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		assert.True(t, Has[string, int](mapper), "Expected mapping before removal")
		Remove[string, int](mapper)
		assert.False(t, Has[string, int](mapper), "Expected mapping to not exist after removal")
	})
}

func TestRemoveMapping(t *testing.T) {
	t.Run("RemoveExistingMapping", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		assert.True(t, Has[string, int](mapper), "Expected mapping to exist before removal")
		Remove[string, int](mapper)
		assert.False(t, Has[string, int](mapper), "Expected mapping to be removed")
	})

	t.Run("RemoveNonexistentMappingNoPanic", func(t *testing.T) {
		mapper := New()
		assert.NotPanics(t, func() {
			Remove[string, int](mapper)
		}, "RemoveMapping panicked")
	})

	t.Run("RemoveOneOfMultipleMappings", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Register(mapper, personToDTO)

		Remove[string, int](mapper)

		assert.False(t, Has[string, int](mapper), "Expected string->int mapping to be removed")
		assert.True(t, Has[int, string](mapper), "Expected int->string mapping to remain")
		assert.True(t, Has[Person, PersonDTO](mapper), "Expected Person->PersonDTO mapping to remain")
	})

	t.Run("RemoveAndReRegister", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		Remove[string, int](mapper)
		Register(mapper, stringToInt)
		assert.True(t, Has[string, int](mapper), "Expected mapping to exist after re-registration")
		result, err := Map[string, int](mapper, "test")
		assert.NoError(t, err)
		assert.Equal(t, 4, result)
	})
}

func TestList(t *testing.T) {
	t.Run("ListEmptyMappings", func(t *testing.T) {
		mapper := New()
		mappings := List(mapper)
		assert.Empty(t, mappings, "Expected 0 mappings")
	})

	t.Run("ListSingleMapping", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		mappings := List(mapper)
		assert.Len(t, mappings, 1)
		assert.Equal(t, "string-int", mappings[0])
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
		assert.Equal(t, expected, mappings)
	})

	t.Run("ListMappingsAfterRemoval", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		Register(mapper, intToString)
		Remove[string, int](mapper)
		mappings := List(mapper)
		assert.Len(t, mappings, 1)
		assert.Equal(t, "int-string", mappings[0])
	})

	t.Run("ListMappingsReturnsCopy", func(t *testing.T) {
		mapper := New()
		Register(mapper, stringToInt)
		mappings1 := List(mapper)
		mappings2 := List(mapper)
		if len(mappings1) > 0 {
			mappings1[0] = "modified"
		}
		if len(mappings2) > 0 {
			assert.NotEqual(t, "modified", mappings2[0], "Expected List to return independent copies")
		}
	})
}
