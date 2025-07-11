package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/hotrungnhan/go-automapper"
)

// Source and destination structs
type Person struct {
	Name  string
	Age   int
	Email string
}

type PersonDTO struct {
	Name  string
	Age   int
	Email string
}

type Employee struct {
	Name     string
	Age      int
	Position string
	Salary   float64
}

type EmployeeDTO struct {
	Name     string
	Age      int
	Position string
	// Note: Salary field is missing - AutoMap will ignore it
}

func main() {
	fmt.Println("🚀 AutoMap Examples")
	fmt.Println("===================")

	// Create a mapper instance
	// m := mapper.New()
	// Can replace this with global
	m := mapper.Global

	// Example 1: Perfect field match
	fmt.Println("\n1. Perfect Field Match:")
	mapper.RegisterAutoMap[Person, PersonDTO](m)

	person := Person{
		Name:  "Alice Johnson",
		Age:   28,
		Email: "alice@example.com",
	}

	personDTO, err := mapper.Map[Person, PersonDTO](m, person)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Original: %+v\n", person)
	fmt.Printf("Mapped:   %+v\n", personDTO)

	// Example 2: Partial field match
	fmt.Println("\n2. Partial Field Match:")
	mapper.RegisterAutoMap[Employee, EmployeeDTO](m)

	employee := Employee{
		Name:     "Bob Smith",
		Age:      35,
		Position: "Senior Developer",
		Salary:   95000.50,
	}

	employeeDTO, err := mapper.Map[Employee, EmployeeDTO](m, employee)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Original: %+v\n", employee)
	fmt.Printf("Mapped:   %+v (Salary field ignored)\n", employeeDTO)

	// Example 3: Using local import for autoMap
	fmt.Println("\n3. Direct AutoMap (using local package):")

	// Get current directory for local import
	_, currentFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(currentFile))
	fmt.Printf("Note: For direct autoMap, import from: %s\n", projectRoot)

	anotherPerson := Person{
		Name:  "Charlie Brown",
		Age:   42,
		Email: "charlie@example.com",
	}

	fmt.Printf("Original: %+v\n", anotherPerson)
	fmt.Println("(Direct autoMap function is not exported, use RegisterAutoMap instead)")

	// Example 4: Bidirectional mapping
	fmt.Println("\n4. Bidirectional Mapping:")

	// AutoMap registers both directions
	originalDTO := PersonDTO{
		Name:  "Diana Prince",
		Age:   30,
		Email: "diana@example.com",
	}

	backToPerson, err := mapper.Map[PersonDTO, Person](m, originalDTO)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("DTO:      %+v\n", originalDTO)
	fmt.Printf("Person:   %+v\n", backToPerson)

	// Example 5: Complex nested structures
	fmt.Println("\n5. Nested Structures:")

	type Address struct {
		Street string
		City   string
		ZIP    string
	}

	type UserWithAddress struct {
		Name    string
		Age     int
		Address Address
	}

	type UserWithAddressDTO struct {
		Name    string
		Age     int
		Address Address
	}

	mapper.RegisterAutoMap[UserWithAddress, UserWithAddressDTO](m)

	userWithAddr := UserWithAddress{
		Name: "Eve Adams",
		Age:  25,
		Address: Address{
			Street: "123 Main St",
			City:   "Springfield",
			ZIP:    "12345",
		},
	}

	userDTO, err := mapper.Map[UserWithAddress, UserWithAddressDTO](m, userWithAddr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Original: %+v\n", userWithAddr)
	fmt.Printf("Mapped:   %+v\n", userDTO)

	// Example 6: Manual mapping vs AutoMap comparison
	fmt.Println("\n6. Manual vs Auto Mapping Comparison:")

	// Manual mapping
	mapper.Register(m, func(p Person) PersonDTO {
		return PersonDTO{
			Name:  "Manual: " + p.Name,
			Age:   p.Age + 1,
			Email: p.Email,
		}
	})

	// This will use the manual mapping (registered last)
	manualResult, _ := mapper.Map[Person, PersonDTO](m, person)
	fmt.Printf("Manual mapping result: %+v\n", manualResult)

	// Example 7: List all registered mappings
	fmt.Println("\n7. Registered Mappings:")
	mappings := mapper.List(m)
	for i, mapping := range mappings {
		fmt.Printf("%d. %s\n", i+1, mapping)
	}

	fmt.Println("\n✅ All AutoMap examples completed successfully!")
}
