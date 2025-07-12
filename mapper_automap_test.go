package mapper

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoMap(t *testing.T) {
	t.Run("AutoMapStructWithSameFieldNames", func(t *testing.T) {
		type Source struct {
			Name string
			Age  int
		}
		type Dest struct {
			Name string
			Age  int
		}

		src := Source{Name: "John", Age: 30}
		result := autoMap[Source, Dest](src)

		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("AutoMapStructWithDifferentFieldNames", func(t *testing.T) {
		type Source struct {
			Name string
			Age  int
		}
		type Dest struct {
			FullName string
			Years    int
		}

		src := Source{Name: "John", Age: 30}
		result := autoMap[Source, Dest](src)

		assert.Zero(t, result.FullName)
		assert.Zero(t, result.Years)
	})

	t.Run("AutoMapStructPartialMatch", func(t *testing.T) {
		type Source struct {
			Name  string
			Age   int
			Email string
		}
		type Dest struct {
			Name string
			Age  int
			City string
		}

		src := Source{Name: "John", Age: 30, Email: "john@example.com"}
		result := autoMap[Source, Dest](src)

		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.Zero(t, result.City)
	})

	t.Run("AutoMapNestedStructs", func(t *testing.T) {
		type Address struct {
			Street string
			City   string
		}
		type Source struct {
			Name    string
			Address Address
		}
		type Dest struct {
			Name    string
			Address Address
		}

		src := Source{
			Name:    "John",
			Address: Address{Street: "123 Main St", City: "NYC"},
		}
		result := autoMap[Source, Dest](src)

		assert.Equal(t, "John", result.Name)
		assert.Equal(t, "123 Main St", result.Address.Street)
		assert.Equal(t, "NYC", result.Address.City)
	})

	t.Run("AutoMapSliceFields", func(t *testing.T) {
		type Source struct {
			Name string
			Tags []string
			Nums []int
		}
		type Dest struct {
			Name string
			Tags []string
			Nums []int
		}

		src := Source{
			Name: "John",
			Tags: []string{"dev", "go"},
			Nums: []int{1, 2, 3},
		}
		result := autoMap[Source, Dest](src)

		assert.Equal(t, "John", result.Name)
		assert.Equal(t, []string{"dev", "go"}, result.Tags)
		assert.Equal(t, []int{1, 2, 3}, result.Nums)
	})

	t.Run("AutoMapPointerFields", func(t *testing.T) {
		type Source struct {
			Name *string
			Age  *int
		}
		type Dest struct {
			Name *string
			Age  *int
		}

		name := "John"
		age := 30
		src := Source{Name: &name, Age: &age}
		result := autoMap[Source, Dest](src)

		assert.NotNil(t, result.Name)
		assert.NotNil(t, result.Age)
		assert.Equal(t, "John", *result.Name)
		assert.Equal(t, 30, *result.Age)
		assert.NotSame(t, src.Name, result.Name)
	})

	t.Run("AutoMapMapFields", func(t *testing.T) {
		type Source struct {
			Name   string
			Attrs  map[string]string
			Counts map[string]int
		}
		type Dest struct {
			Name   string
			Attrs  map[string]string
			Counts map[string]int
		}

		src := Source{
			Name:   "John",
			Attrs:  map[string]string{"role": "dev", "team": "backend"},
			Counts: map[string]int{"projects": 5, "bugs": 2},
		}
		result := autoMap[Source, Dest](src)

		assert.Equal(t, "John", result.Name)
		assert.Equal(t, map[string]string{"role": "dev", "team": "backend"}, result.Attrs)
		assert.Equal(t, map[string]int{"projects": 5, "bugs": 2}, result.Counts)
	})

	t.Run("AutoMapInterfaceFields", func(t *testing.T) {
		type Source struct {
			Name  string
			Value interface{}
		}
		type Dest struct {
			Name  string
			Value interface{}
		}

		src := Source{Name: "John", Value: 42}
		result := autoMap[Source, Dest](src)

		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 42, result.Value)
	})

	t.Run("AutoMapEmptyStruct", func(t *testing.T) {
		type Empty struct{}

		src := Empty{}
		result := autoMap[Empty, Empty](src)

		assert.True(t, reflect.DeepEqual(src, result))
	})

	t.Run("AutoMapPrimitiveTypes", func(t *testing.T) {
		src := "hello"
		result := autoMap[string, string](src)
		assert.Equal(t, "hello", result)

		srcInt := 42
		resultInt := autoMap[int, int](srcInt)
		assert.Equal(t, 42, resultInt)
	})

	t.Run("AutoMapDifferentPrimitiveTypes", func(t *testing.T) {
		src := "hello"
		result := autoMap[string, int](src)
		assert.Zero(t, result)
	})

	t.Run("AutoMapWithEmbeddedStructs", func(t *testing.T) {
		type BaseInfo struct {
			ID   int
			Name string
		}
		type Source struct {
			BaseInfo
			Email string
		}
		type Dest struct {
			BaseInfo
			Email string
		}

		src := Source{
			BaseInfo: BaseInfo{ID: 1, Name: "John"},
			Email:    "john@example.com",
		}
		result := autoMap[Source, Dest](src)

		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, "john@example.com", result.Email)
	})
}

func TestRegisterAutoMap(t *testing.T) {
	t.Run("RegisterAutoMapSimpleStruct", func(t *testing.T) {
		type Source struct {
			Name string
			Age  int
		}
		type Dest struct {
			Name string
			Age  int
		}

		mapper := New()
		RegisterAutoMap[Source, Dest](mapper)

		assert.True(t, Has[Source, Dest](mapper))

		src := Source{Name: "John", Age: 30}
		result, err := Map[Source, Dest](mapper, src)

		assert.NoError(t, err)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("RegisterAutoMapPartialMatch", func(t *testing.T) {
		type Source struct {
			Name  string
			Age   int
			Email string
		}
		type Dest struct {
			Name string
			Age  int
			City string
		}

		mapper := New()
		RegisterAutoMap[Source, Dest](mapper)

		src := Source{Name: "John", Age: 30, Email: "john@example.com"}
		result, err := Map[Source, Dest](mapper, src)

		assert.NoError(t, err)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.Zero(t, result.City)
	})

	t.Run("RegisterAutoMapMultipleTypes", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		type PersonDTO struct {
			Name string
			Age  int
		}
		type Address struct {
			Street string
			City   string
		}
		type AddressDTO struct {
			Street string
			City   string
		}

		mapper := New()
		RegisterAutoMap[Person, PersonDTO](mapper)
		RegisterAutoMap[Address, AddressDTO](mapper)

		person := Person{Name: "John", Age: 30}
		personResult, err := Map[Person, PersonDTO](mapper, person)
		assert.NoError(t, err)
		assert.Equal(t, "John", personResult.Name)
		assert.Equal(t, 30, personResult.Age)

		addr := Address{Street: "123 Main St", City: "NYC"}
		addrResult, err := Map[Address, AddressDTO](mapper, addr)
		assert.NoError(t, err)
		assert.Equal(t, "123 Main St", addrResult.Street)
		assert.Equal(t, "NYC", addrResult.City)
	})

	t.Run("RegisterAutoMapOverwritesExisting", func(t *testing.T) {
		type Source struct {
			Name string
			Age  int
		}
		type Dest struct {
			Name string
			Age  int
		}

		mapper := New()

		Register(mapper, func(s Source) Dest {
			return Dest{Name: "Manual: " + s.Name, Age: s.Age + 10}
		})

		src := Source{Name: "John", Age: 30}
		result1, _ := Map[Source, Dest](mapper, src)

		RegisterAutoMap[Source, Dest](mapper)
		result2, _ := Map[Source, Dest](mapper, src)

		assert.NotEqual(t, result1.Name, result2.Name)
		assert.Equal(t, "John", result2.Name)
		assert.Equal(t, 30, result2.Age)
	})

	t.Run("RegisterAutoMapComplexNestedStructs", func(t *testing.T) {
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
		}
		type Dest struct {
			Name    string
			Age     int
			Address Address
			Contact Contact
			Tags    []string
		}

		mapper := New()
		RegisterAutoMap[Source, Dest](mapper)

		src := Source{
			Name: "John",
			Age:  30,
			Address: Address{
				Street: "123 Main St",
				City:   "NYC",
				ZIP:    "10001",
			},
			Contact: Contact{
				Email: "john@example.com",
				Phone: "555-1234",
			},
			Tags: []string{"developer", "golang"},
		}

		result, err := Map[Source, Dest](mapper, src)

		assert.NoError(t, err)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
		assert.Equal(t, "123 Main St", result.Address.Street)
		assert.Equal(t, "NYC", result.Address.City)
		assert.Equal(t, "john@example.com", result.Contact.Email)
		assert.Equal(t, "555-1234", result.Contact.Phone)
		assert.Equal(t, []string{"developer", "golang"}, result.Tags)
	})

	t.Run("RegisterAutoMapWithPointers", func(t *testing.T) {
		type Source struct {
			Name *string
			Age  *int
		}
		type Dest struct {
			Name *string
			Age  *int
		}

		mapper := New()
		RegisterAutoMap[Source, Dest](mapper)

		name := "John"
		age := 30
		src := Source{Name: &name, Age: &age}

		result, err := Map[Source, Dest](mapper, src)

		assert.NoError(t, err)
		assert.NotNil(t, result.Name)
		assert.NotNil(t, result.Age)
		assert.Equal(t, "John", *result.Name)
		assert.Equal(t, 30, *result.Age)
	})

	t.Run("RegisterAutoMapBidirectional", func(t *testing.T) {
		type PersonA struct {
			Name string
			Age  int
		}
		type PersonB struct {
			Name string
			Age  int
		}

		mapper := New()
		RegisterAutoMap[PersonA, PersonB](mapper)
		RegisterAutoMap[PersonB, PersonA](mapper)

		personA := PersonA{Name: "John", Age: 30}
		resultB, err := Map[PersonA, PersonB](mapper, personA)
		assert.NoError(t, err)
		assert.Equal(t, "John", resultB.Name)
		assert.Equal(t, 30, resultB.Age)

		personB := PersonB{Name: "Jane", Age: 25}
		resultA, err := Map[PersonB, PersonA](mapper, personB)
		assert.NoError(t, err)
		assert.Equal(t, "Jane", resultA.Name)
		assert.Equal(t, 25, resultA.Age)
	})

	t.Run("RegisterAutoMapReverseMap", func(t *testing.T) {
		type Foo struct {
			X int
			Y string
		}
		type Bar struct {
			X int
			Y string
		}

		mapper := New()
		RegisterAutoMap[Foo, Bar](mapper)
		RegisterAutoMap[Bar, Foo](mapper)

		foo := Foo{X: 42, Y: "hello"}
		bar, err := Map[Foo, Bar](mapper, foo)
		assert.NoError(t, err)
		assert.Equal(t, 42, bar.X)
		assert.Equal(t, "hello", bar.Y)

		foo2, err := Map[Bar, Foo](mapper, bar)
		assert.NoError(t, err)
		assert.Equal(t, 42, foo2.X)
		assert.Equal(t, "hello", foo2.Y)
	})
}
