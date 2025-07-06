# Go Type Mapper 

> **⚠️ Notice:** This library is under active development. The API may change until version 1.0.0 is released.
A high-performance, type-safe mapping library for Go that allows you to register and execute type conversion functions using generics and reflection.

**Summary:**  
Use manual mapping for the absolute fastest scenarios. For most real-world applications, this library offers an ideal balance of speed, safety, and developer productivity. Avoid slower automapper libraries in performance-sensitive code.

## 🚀 Features

- **Type-Safe**: Uses Go generics for compile-time type safety
- **High Performance**: ~25ns per mapping operation with zero allocations
- **Flexible**: Support for any type conversion (primitives, structs, pointers, interfaces)
- **Simple API**: Clean and intuitive interface
- **Zero Dependencies**: No external dependencies beyond Go standard library
- **Thread-Safe Registry**: Safe concurrent access to mapping registry
- **Comprehensive Testing**: 100% test coverage with extensive benchmarks

## 📊 Performance

```text
BenchmarkMap-10                 95016578      25.36 ns/op      0 B/op      0 allocs/op
BenchmarkMapStruct-10           95307589      25.12 ns/op      0 B/op      0 allocs/op
BenchmarkRegister-10            75473577      32.30 ns/op      0 B/op      0 allocs/op
BenchmarkHasMapping-10          86743591      27.77 ns/op      0 B/op      0 allocs/op
```

**~40 million mapping operations per second** - Production ready performance!

## 🛠 Installation

```bash
go get github.com/hotrungnhan/go-automapper
```

## 📖 Quick Start

```go
package main

import (
    "fmt"
    "github.com/hotrungnhan/go-automapper"
)

type Person struct {
    Name string
    Age  int
}

type PersonDTO struct {
    FullName string
    Years    int
}

func main() {
    // Create a new mapper
    m := mapper.NewMapper()
    
    // Register a mapping function
    mapper.Register(m, func(p Person) PersonDTO {
        return PersonDTO{
            FullName: p.Name,
            Years:    p.Age,
        }
    })
    
    // Use the mapping
    person := Person{Name: "John Doe", Age: 30}
    dto, err := mapper.Map[Person, PersonDTO](m, person)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("%+v\n", dto) // {FullName:John Doe Years:30}
}
```

## 📚 API Documentation

### Creating a Mapper

```go
m := mapper.NewMapper()
```

Creates a new mapper instance with an empty registry.

### Registering Mappings

```go
mapper.Register[SourceType, DestType](m, mappingFunction)
```

Registers a mapping function from `SourceType` to `DestType`.

**Example:**

```go
// Simple type conversion
mapper.Register(m, func(s string) int {
    return len(s)
})

// Struct conversion
mapper.Register(m, func(p Person) PersonDTO {
    return PersonDTO{FullName: p.Name, Years: p.Age}
})

// Pointer conversion
mapper.Register(m, func(s *string) *int {
    if s == nil { return nil }
    length := len(*s)
    return &length
})
```

### Executing Mappings

```go
result, err := mapper.Map[SourceType, DestType](m, sourceValue)
```

Executes a registered mapping function.

**Example:**

```go
// Map string to int
length, err := mapper.Map[string, int](m, "hello")
// Result: length = 5, err = nil

// Map struct
dto, err := mapper.Map[Person, PersonDTO](m, person)
```

### Checking Mapping Existence

```go
exists := mapper.HasMapping[SourceType, DestType](m)
```

Checks if a mapping from `SourceType` to `DestType` is registered.

**Example:**

```go
if mapper.HasMapping[string, int](m) {
    fmt.Println("String to int mapping exists")
}
```

### Removing Mappings

```go
mapper.RemoveMapping[SourceType, DestType](m)
```

Removes a registered mapping.

**Example:**

```go
mapper.RemoveMapping[string, int](m)
```

### Listing All Mappings

```go
mappings := m.ListMappings()
```

Returns a slice of all registered mapping type pairs as strings.

**Example:**

```go
mappings := m.ListMappings()
// Result: ["string-int", "main.Person-main.PersonDTO"]
```

## 🎯 Use Cases

### 1. API Layer Transformations

```go
// Convert domain models to API responses
mapper.Register(m, func(user User) UserResponse {
    return UserResponse{
        ID:       user.ID,
        Name:     user.FullName(),
        Email:    user.Email,
        JoinDate: user.CreatedAt.Format("2006-01-02"),
    }
})

users, _ := mapper.Map[[]User, []UserResponse](m, domainUsers)
```

### 2. Database Layer Mappings

```go
// Convert database rows to domain objects
mapper.Register(m, func(row UserRow) User {
    return User{
        ID:        row.ID,
        FirstName: row.FirstName,
        LastName:  row.LastName,
        Email:     row.Email,
        CreatedAt: row.CreatedAt,
    }
})
```

### 3. Configuration Transformations

```go
// Convert configuration formats
mapper.Register(m, func(yaml YAMLConfig) JSONConfig {
    return JSONConfig{
        Host: yaml.Server.Host,
        Port: yaml.Server.Port,
        DB:   yaml.Database.URL,
    }
})
```

### 4. Event Processing

```go
// Transform events between formats
mapper.Register(m, func(event DomainEvent) MessageBusEvent {
    return MessageBusEvent{
        Type:      event.GetType(),
        Payload:   event.Serialize(),
        Timestamp: event.OccurredAt(),
    }
})
```

## 🔧 Advanced Usage

### Bidirectional Mappings

```go
// Register both directions
mapper.Register(m, func(p Person) PersonDTO {
    return PersonDTO{FullName: p.Name, Years: p.Age}
})

mapper.Register(m, func(dto PersonDTO) Person {
    return Person{Name: dto.FullName, Age: dto.Years}
})
```

### Error Handling

```go
result, err := mapper.Map[Source, Dest](m, source)
if err != nil {
    // Handle mapping not found
    fmt.Printf("No mapping registered: %v", err)
}
```

### Complex Type Mappings

```go
// Generic types
type Container[T any] struct { Value T }

mapper.Register(m, func(c Container[string]) Container[int] {
    return Container[int]{Value: len(c.Value)}
})

// Interface mappings
mapper.Register(m, func(r io.Reader) []byte {
    data, _ := io.ReadAll(r)
    return data
})
```

### Registry Management

```go
// Check before registering
if !mapper.HasMapping[Source, Dest](m) {
    mapper.Register(m, mappingFunc)
}

// List all mappings for debugging
mappings := m.ListMappings()
for _, mapping := range mappings {
    fmt.Println("Registered:", mapping)
}

// Clean up mappings
mapper.RemoveMapping[OldSource, OldDest](m)
```

## ⚡ Performance Tips

1. **Reuse Mapper Instances**: Create one mapper per application lifecycle
2. **Register Once**: Register mappings at startup, not per operation
3. **Avoid Complex Logic**: Keep mapping functions simple and fast
4. **Batch Operations**: Process collections when possible

```go
// Good: Simple, fast mapping
mapper.Register(m, func(s string) int { return len(s) })

// Avoid: Complex operations in mapping
mapper.Register(m, func(url string) Data {
    // Don't do HTTP calls or heavy computation here
    return fetchDataFromAPI(url) // ❌
})
```

### 🏎️ AutoMapper vs Manual Mapping Benchmarks

Below are comparative benchmarks between this library, manual mapping, and other automapper libraries:

```text
BenchmarkMapperVsManual/Mapper-10           47,657,895    25.45 ns/op    0 B/op    0 allocs/op
BenchmarkMapperVsManual/Manual-10        1,000,000,000     0.42 ns/op    0 B/op    0 allocs/op

BenchmarkManualMappingVsAutoMap/Manual-10     37,053,046    31.87 ns/op    0 B/op    0 allocs/op
BenchmarkManualMappingVsAutoMap/AutoMap-10        719,054  1624 ns/op    496 B/op   15 allocs/op
BenchmarkManualMappingVsAutoMap/DirectAutoMap-10   751,530  1547 ns/op    496 B/op   15 allocs/op
```

### 📈 Recommendation by Use Case

- **Manual Mapping**:  
  - *Best for*: Ultra-high-performance, hot code paths, or extremely simple mappings.
  - *Why*: Manual code is always fastest (sub-nanosecond), with zero overhead.

- **With Mapped Function**:  
  - *Best for*: Most application code, especially when you want type safety, maintainability, and flexibility.
  - *Why*: Only ~25ns/op, zero allocations, and much easier to maintain than manual mapping for many types.

- **With Automapper**:  
  - *Best for*: When you need advanced features not present here, and performance is less critical.
  - *Why*: Typically 50–60x slower (1500+ ns/op) and introduce allocations.

## 🧪 Testing

The library includes comprehensive tests covering:

- ✅ All API methods
- ✅ Edge cases (nil, empty values, zero values)
- ✅ Error conditions
- ✅ Type safety
- ✅ Memory efficiency
- ✅ Performance benchmarks

Run tests:

```bash
go test -v                    # Run all tests
go test -cover               # Run with coverage
go test -bench=.             # Run benchmarks
```

## 🔍 Error Types

### Mapping Not Found

```go
_, err := mapper.Map[string, int](m, "test")
// Error: "no mapping function registered for string to int"
```

### Type Assertion Errors

Internal type assertions are safe and will not panic. Invalid registrations are caught at compile-time due to generic constraints.

## 🏗 Architecture

The mapper uses:

- **Reflection-based keys**: `reflect.Type` for efficient map lookups
- **Generic constraints**: Compile-time type safety
- **Zero-allocation hot path**: No memory allocations for successful mappings
- **Simple registry**: Map-based storage for O(1) lookups

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Run benchmarks (`go test -bench=.`)
6. Commit your changes (`git commit -am 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## 📝 Semantic Commit Emoji Guide

Use the following semantic emoji in your commit messages to indicate the type of change:

| Type     | Emoji | Code                    | Description                                  |
| -------- | ----- | ----------------------- | -------------------------------------------- |
| feat     | ✨     | `:sparkles:`            | A new feature                                |
| fix      | 🐛     | `:bug:`                 | A bug fix                                    |
| docs     | 📚     | `:books:`               | Documentation only changes                   |
| style    | 💎     | `:gem:`                 | Code style changes (formatting, etc.)        |
| refactor | 🔨     | `:hammer:`              | Code refactoring (no feature or fix)         |
| perf     | 🚀     | `:rocket:`              | Performance improvements                     |
| test     | 🚨     | `:rotating_light:`      | Adding or updating tests                     |
| build    | 📦     | `:package:`             | Build system or external dependency changes  |
| ci       | 👷     | `:construction_worker:` | CI/CD configuration changes                  |
| chore    | 🔧     | `:wrench:`              | Other changes that don't modify src or tests |

**Example commit message:**

```
✨ feat: Add support for custom mapping functions
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🎖 Acknowledgments

- Inspired by AutoMapper (.NET) and similar mapping libraries
- Built with Go's powerful generics and reflection capabilities
- Optimized for high-performance applications

---

Made with ❤️ for the Go community
