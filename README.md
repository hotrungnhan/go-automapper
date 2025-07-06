# Go AutoMapper

> **⚠️ Notice:** This library is under active development. The API may change until version 1.0.0 is released.
A high-performance, type-safe mapping library for Go that allows you to register and execute type conversion functions using generics and reflection.

**Summary:**
Use manual mapping for the absolute fastest scenarios. For most real-world applications, this library offers an ideal balance of speed, safety, and developer productivity. Avoid slower automapper libraries in performance-sensitive code.

## 🚀 Features

- **Type-Safe**: Uses Go generics for compile-time type safety
- **High Performance**: ~25ns per mapping operation with zero allocations
- **Flexible**: Support for any type conversion (primitives, structs, pointers, interfaces)
- **Simple API**: Clean and intuitive interface
- **Comprehensive Testing**: 100% test coverage with extensive benchmarks

## 🛠 Installation

```bash
go get github.com/hotrungnhan/go-automapper
```

## 📖 Quick Start

There are 3 ways to map data between types in Go:

### 1. Manual Mapping (No Library)

```go

type Person struct {
    Name string
    Age  int
}

// Manual conversion between types
func PersonToDTO(p Person) PersonDTO {
    return PersonDTO{
        FullName: p.Name,
        Years:    p.Age,
    }
}

person := Person{Name: "Alice", Age: 28}
dto := PersonToDTO(person)
```

**Pros:** Fastest possible, zero overhead.
**Cons:** Hard to manage, not generic, lots of boilerplate.

---

### 2. Using This Library with Your Mapping Function

```go
import "github.com/hotrungnhan/go-automapper"

func main(){
    m := mapper.NewMapper()

    // Single Direction mapping
    mapper.Register(m, func(p Person) PersonDTO {
        return PersonDTO{
            FullName: p.Name,
            Years:    p.Age,
        }
    })

    person := Person{Name: "Alice", Age: 28}
    dto, err := mapper.Map[Person, PersonDTO](m, person)
    if err != nil {
        panic(err)
    }
}

```

**Pros:** Type-safe, maintainable, generic, and almost as fast as manual mapping.
**Cons:** You still write mapping functions, but only once per type pair.

### 3. Using Auto Mapping with Direct Auto Map

```go

import "github.com/hotrungnhan/go-automapper"

func main(){
    m := mapper.NewMapper()

    // bidirection
    mapper.RegisterAutoMap[Person, PersonDTO](m)
    person := Person{Name: "Alice", Age: 28}
    dto, err := mapper.Map[Person, PersonDTO](m, person)
    if err != nil {
        panic(err)
    }
}


```

**Pros:** No mapping function needed, works for most structs with similar fields.  
**Cons:** Dramatically slower (50x+), allocates memory, not for hot paths.

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

### Registry Management

```go
// List all mappings for debugging
mappings := mapper.List(m)
for _, mapping := range mappings {
    fmt.Println("Registered:", mapping)
}

// Clean up mappings
mapper.RemoveMapping[OldSource, OldDest](m)**
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

Below are comparative benchmarks between multiple way to using this library:

| Test Name                                        |    Iterations | Time (ns/op) | Memory (B/op) | Allocs (op) |
| ------------------------------------------------ | ------------: | -----------: | ------------: | ----------: |
| BenchmarkMapperVsManual/Mapper-10                |    47,657,895 |        25.45 |             0 |           0 |
| BenchmarkMapperVsManual/Manual-10                | 1,000,000,000 |         0.42 |             0 |           0 |
| BenchmarkManualMappingVsAutoMap/Manual-10        |    37,053,046 |        31.87 |             0 |           0 |
| BenchmarkManualMappingVsAutoMap/AutoMap-10       |       719,054 |         1624 |           496 |          15 |
| BenchmarkManualMappingVsAutoMap/DirectAutoMap-10 |       751,530 |         1547 |           496 |          15 |

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

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Run benchmarks (`go test -bench=.`)
6. Commit your changes (`git commit -am 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### 📝 Semantic Commit Emoji Guide

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

```text
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
