package mapper

import (
	"reflect"
	"testing"
)

// TestAutoMap tests the automatic mapping functionality
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

		if result.Name != "John" || result.Age != 30 {
			t.Errorf("Expected {John 30}, got %+v", result)
		}
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

		// Should only copy matching field names, others remain zero values
		if result.FullName != "" || result.Years != 0 {
			t.Errorf("Expected zero values for non-matching fields, got %+v", result)
		}
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
			City string // No matching field in source
		}

		src := Source{Name: "John", Age: 30, Email: "john@example.com"}
		result := autoMap[Source, Dest](src)

		if result.Name != "John" || result.Age != 30 {
			t.Errorf("Expected matching fields to be copied, got %+v", result)
		}
		if result.City != "" {
			t.Errorf("Expected City to be zero value, got %s", result.City)
		}
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

		if result.Name != "John" {
			t.Errorf("Expected Name to be copied, got %s", result.Name)
		}
		if result.Address.Street != "123 Main St" || result.Address.City != "NYC" {
			t.Errorf("Expected nested struct to be copied, got %+v", result.Address)
		}
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

		if result.Name != "John" {
			t.Errorf("Expected Name to be copied, got %s", result.Name)
		}
		if len(result.Tags) != 2 || result.Tags[0] != "dev" || result.Tags[1] != "go" {
			t.Errorf("Expected Tags to be copied, got %+v", result.Tags)
		}
		if len(result.Nums) != 3 || result.Nums[0] != 1 || result.Nums[2] != 3 {
			t.Errorf("Expected Nums to be copied, got %+v", result.Nums)
		}
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

		if result.Name == nil || *result.Name != "John" {
			t.Errorf("Expected Name pointer to be copied, got %v", result.Name)
		}
		if result.Age == nil || *result.Age != 30 {
			t.Errorf("Expected Age pointer to be copied, got %v", result.Age)
		}

		// Verify deep copy (different pointers)
		if result.Name == src.Name {
			t.Error("Expected deep copy, but pointers are the same")
		}
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

		if result.Name != "John" {
			t.Errorf("Expected Name to be copied, got %s", result.Name)
		}
		if len(result.Attrs) != 2 || result.Attrs["role"] != "dev" {
			t.Errorf("Expected Attrs to be copied, got %+v", result.Attrs)
		}
		if len(result.Counts) != 2 || result.Counts["projects"] != 5 {
			t.Errorf("Expected Counts to be copied, got %+v", result.Counts)
		}
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

		if result.Name != "John" {
			t.Errorf("Expected Name to be copied, got %s", result.Name)
		}
		if result.Value != 42 {
			t.Errorf("Expected Value to be copied, got %v", result.Value)
		}
	})

	t.Run("AutoMapEmptyStruct", func(t *testing.T) {
		type Empty struct{}

		src := Empty{}
		result := autoMap[Empty, Empty](src)

		if !reflect.DeepEqual(src, result) {
			t.Error("Expected empty structs to be equal")
		}
	})

	t.Run("AutoMapPrimitiveTypes", func(t *testing.T) {
		// String to string
		src := "hello"
		result := autoMap[string, string](src)
		if result != "hello" {
			t.Errorf("Expected 'hello', got %s", result)
		}

		// Int to int
		srcInt := 42
		resultInt := autoMap[int, int](srcInt)
		if resultInt != 42 {
			t.Errorf("Expected 42, got %d", resultInt)
		}
	})

	t.Run("AutoMapDifferentPrimitiveTypes", func(t *testing.T) {
		// This should fail or return zero value since types don't match
		src := "hello"
		result := autoMap[string, int](src)
		if result != 0 {
			t.Errorf("Expected zero value for incompatible types, got %d", result)
		}
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

		if result.ID != 1 || result.Name != "John" || result.Email != "john@example.com" {
			t.Errorf("Expected embedded struct fields to be copied, got %+v", result)
		}
	})
}

// TestRegisterAutoMap tests the automatic mapping registration functionality
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

		if !Has[Source, Dest](mapper) {
			t.Error("Expected automap to be registered")
		}

		src := Source{Name: "John", Age: 30}
		result, err := Map[Source, Dest](mapper, src)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.Name != "John" || result.Age != 30 {
			t.Errorf("Expected {John 30}, got %+v", result)
		}
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
			City string // No matching field
		}

		mapper := New()
		RegisterAutoMap[Source, Dest](mapper)

		src := Source{Name: "John", Age: 30, Email: "john@example.com"}
		result, err := Map[Source, Dest](mapper, src)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.Name != "John" || result.Age != 30 {
			t.Errorf("Expected matching fields to be copied, got %+v", result)
		}
		if result.City != "" {
			t.Errorf("Expected City to be zero value, got %s", result.City)
		}
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

		// Test Person mapping
		person := Person{Name: "John", Age: 30}
		personResult, err := Map[Person, PersonDTO](mapper, person)
		if err != nil {
			t.Errorf("Unexpected error for Person mapping: %v", err)
		}
		if personResult.Name != "John" || personResult.Age != 30 {
			t.Errorf("Expected Person automap to work, got %+v", personResult)
		}

		// Test Address mapping
		addr := Address{Street: "123 Main St", City: "NYC"}
		addrResult, err := Map[Address, AddressDTO](mapper, addr)
		if err != nil {
			t.Errorf("Unexpected error for Address mapping: %v", err)
		}
		if addrResult.Street != "123 Main St" || addrResult.City != "NYC" {
			t.Errorf("Expected Address automap to work, got %+v", addrResult)
		}
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

		// Register manual mapping first
		Register(mapper, func(s Source) Dest {
			return Dest{Name: "Manual: " + s.Name, Age: s.Age + 10}
		})

		src := Source{Name: "John", Age: 30}
		result1, _ := Map[Source, Dest](mapper, src)

		// Register automap (should overwrite)
		RegisterAutoMap[Source, Dest](mapper)
		result2, _ := Map[Source, Dest](mapper, src)

		if result1.Name == result2.Name {
			t.Error("Expected automap to overwrite manual mapping")
		}
		if result2.Name != "John" || result2.Age != 30 {
			t.Errorf("Expected automap result, got %+v", result2)
		}
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

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.Name != "John" || result.Age != 30 {
			t.Errorf("Expected basic fields to be copied, got %+v", result)
		}
		if result.Address.Street != "123 Main St" || result.Address.City != "NYC" {
			t.Errorf("Expected Address to be copied, got %+v", result.Address)
		}
		if result.Contact.Email != "john@example.com" || result.Contact.Phone != "555-1234" {
			t.Errorf("Expected Contact to be copied, got %+v", result.Contact)
		}
		if len(result.Tags) != 2 || result.Tags[0] != "developer" {
			t.Errorf("Expected Tags to be copied, got %+v", result.Tags)
		}
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

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.Name == nil || *result.Name != "John" {
			t.Errorf("Expected Name pointer to be copied, got %v", result.Name)
		}
		if result.Age == nil || *result.Age != 30 {
			t.Errorf("Expected Age pointer to be copied, got %v", result.Age)
		}
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

		// Test A -> B
		personA := PersonA{Name: "John", Age: 30}
		resultB, err := Map[PersonA, PersonB](mapper, personA)
		if err != nil {
			t.Errorf("Unexpected error A->B: %v", err)
		}
		if resultB.Name != "John" || resultB.Age != 30 {
			t.Errorf("Expected A->B mapping to work, got %+v", resultB)
		}

		// Test B -> A
		personB := PersonB{Name: "Jane", Age: 25}
		resultA, err := Map[PersonB, PersonA](mapper, personB)
		if err != nil {
			t.Errorf("Unexpected error B->A: %v", err)
		}
		if resultA.Name != "Jane" || resultA.Age != 25 {
			t.Errorf("Expected B->A mapping to work, got %+v", resultA)
		}
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
		if err != nil {
			t.Fatalf("Unexpected error mapping Foo->Bar: %v", err)
		}
		if bar.X != 42 || bar.Y != "hello" {
			t.Errorf("Expected Bar{42, hello}, got %+v", bar)
		}

		foo2, err := Map[Bar, Foo](mapper, bar)
		if err != nil {
			t.Fatalf("Unexpected error mapping Bar->Foo: %v", err)
		}
		if foo2.X != 42 || foo2.Y != "hello" {
			t.Errorf("Expected Foo{42, hello}, got %+v", foo2)
		}
	})
}
