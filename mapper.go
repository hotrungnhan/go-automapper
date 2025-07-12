// Package mapper provides a type-safe, generic mapping library for Go that allows
// registration and execution of mapping functions between different types.
// It supports both manual mapping function registration and automatic field mapping.
package mapper

import (
	"errors"
	"reflect"
)

// Global is a default mapper instance that can be used for convenience.
// It provides a shared mapping registry for applications that don't need multiple mapper instances.
var Global = New()

// typePair represents a mapping relationship between source and destination types.
// It serves as the key in the mapper registry to identify registered mapping functions.
type typePair struct {
	src reflect.Type
	dst reflect.Type
}

// Mapper is the main mapping registry that stores mapping functions between type pairs.
// Each mapper instance maintains its own independent registry of mapping functions.
type Mapper struct {
	registry map[typePair]interface{}
}

// ErrNoMapping is returned when attempting to map between types that don't have
// a registered mapping function.
var ErrNoMapping = errors.New("no mapping function registered for this type pair")

// ErrSrcAndDestMustBeSlices is returned when a function expects both the source and destination
// parameters to be slices, but one or both are not. This error helps enforce type safety
// when performing operations that require slice types.
var ErrSrcAndDestMustBeSlices = errors.New("both source and destination must be slices")

// New creates a new Mapper instance with an empty registry.
// Each mapper maintains its own independent registry of mapping functions.
//
// Returns:
//   - Mapper: A new mapper instance with an initialized empty registry
//
// Example:
//
//	mapper := New()
//	Register(mapper, func(s string) int { return len(s) })
//	result, err := Map[string, int](mapper, "hello")
func New() Mapper {
	return Mapper{
		registry: make(map[typePair]interface{}),
	}
}

// Register registers a mapping function for converting from type S to type D.
// The function will be stored in the mapper's registry and can be used by Map and MapSlice.
// If a mapping for the same type pair already exists, it will be overwritten.
//
// Type Parameters:
//   - S: Source type (input type for the mapping function)
//   - D: Destination type (output type for the mapping function)
//
// Parameters:
//   - m: The mapper instance to register the function with
//   - fn: The mapping function that converts from S to D
//
// Example:
//
//	mapper := New()
//	Register(mapper, func(s string) int { return len(s) })
//	Register(mapper, func(p Person) PersonDTO {
//	    return PersonDTO{Name: p.FirstName + " " + p.LastName}
//	})
func Register[S any, D any](m Mapper, fn func(S) D) {
	key := typePair{
		src: reflect.TypeOf((*S)(nil)).Elem(),
		dst: reflect.TypeOf((*D)(nil)).Elem(),
	}
	m.registry[key] = fn
}

// Map executes a registered mapping function to convert a value from type S to type D.
// It supports mapping between values, pointers, and mixed value/pointer combinations.
// The function automatically handles pointer dereferencing and creation as needed.
//
// Type Parameters:
//   - S: Source type (must match a registered mapping function's input type)
//   - D: Destination type (must match a registered mapping function's output type)
//
// Parameters:
//   - m: The mapper instance containing the registered mapping functions
//   - src: The source value to be mapped
//
// Returns:
//   - D: The mapped result of type D
//   - error: ErrNoMapping if no mapping function is registered for the type pair
//
// Supported mapping combinations:
//   - Value to Value: T -> U
//   - Pointer to Pointer: *T -> *U (nil input returns nil output)
//   - Value to Pointer: T -> *U
//   - Pointer to Value: *T -> U (nil input returns zero value)
//
// Example:
//
//	mapper := New()
//	Register(mapper, func(s string) int { return len(s) })
//
//	result, err := Map[string, int](mapper, "hello")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result) // Output: 5
//
//	// Pointer mapping
//	name := "world"
//	ptrResult, err := Map[*string, *int](mapper, &name)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(*ptrResult) // Output: 5
func Map[S any, D any](m Mapper, src S) (D, error) {
	var dst D

	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)

	// Get the underlying types for the registry key
	keySrcType := srcType
	keyDstType := dstType

	if keySrcType.Kind() == reflect.Ptr {
		keySrcType = keySrcType.Elem()
	}
	if keyDstType.Kind() == reflect.Ptr {
		keyDstType = keyDstType.Elem()
	}

	key := typePair{
		src: keySrcType,
		dst: keyDstType,
	}

	fn, ok := m.registry[key]
	if !ok {
		return dst, ErrNoMapping
	}

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// Handle the four cases based on pointer combinations
	srcValue := reflect.ValueOf(src)

	// Case 1: src is pointer, dst is pointer
	if srcType.Kind() == reflect.Ptr && dstType.Kind() == reflect.Ptr {
		if srcValue.IsNil() {
			return dst, nil // return zero value (nil pointer)
		}

		var callArg reflect.Value
		if fnType.In(0).Kind() == reflect.Ptr {
			// Function expects pointer, we have pointer
			callArg = srcValue
		} else {
			// Function expects value, we have pointer - dereference
			callArg = srcValue.Elem()
		}

		result := fnValue.Call([]reflect.Value{callArg})[0]

		if fnType.Out(0).Kind() == reflect.Ptr {
			// Function returns pointer, we need pointer
			return result.Interface().(D), nil
		} else {
			// Function returns value, we need pointer - create pointer
			ptrResult := reflect.New(result.Type())
			ptrResult.Elem().Set(result)
			return ptrResult.Interface().(D), nil
		}
	}

	// Case 2: src is pointer, dst is value
	if srcType.Kind() == reflect.Ptr && dstType.Kind() != reflect.Ptr {
		if srcValue.IsNil() {
			return dst, nil // return zero value
		}

		var callArg reflect.Value
		if fnType.In(0).Kind() == reflect.Ptr {
			// Function expects pointer, we have pointer
			callArg = srcValue
		} else {
			// Function expects value, we have pointer - dereference
			callArg = srcValue.Elem()
		}

		result := fnValue.Call([]reflect.Value{callArg})[0]

		if fnType.Out(0).Kind() == reflect.Ptr {
			// Function returns pointer, we need value - dereference
			if result.IsNil() {
				return dst, nil
			}
			return result.Elem().Interface().(D), nil
		} else {
			// Function returns value, we need value
			return result.Interface().(D), nil
		}
	}

	// Case 3: src is value, dst is pointer
	if srcType.Kind() != reflect.Ptr && dstType.Kind() == reflect.Ptr {
		var callArg reflect.Value
		if fnType.In(0).Kind() == reflect.Ptr {
			// Function expects pointer, we have value - create pointer
			ptrArg := reflect.New(srcValue.Type())
			ptrArg.Elem().Set(srcValue)
			callArg = ptrArg
		} else {
			// Function expects value, we have value
			callArg = srcValue
		}

		result := fnValue.Call([]reflect.Value{callArg})[0]

		if fnType.Out(0).Kind() == reflect.Ptr {
			// Function returns pointer, we need pointer
			return result.Interface().(D), nil
		} else {
			// Function returns value, we need pointer - create pointer
			ptrResult := reflect.New(result.Type())
			ptrResult.Elem().Set(result)
			return ptrResult.Interface().(D), nil
		}
	}

	// Case 4: src is value, dst is value
	if srcType.Kind() != reflect.Ptr && dstType.Kind() != reflect.Ptr {
		var callArg reflect.Value
		if fnType.In(0).Kind() == reflect.Ptr {
			// Function expects pointer, we have value - create pointer
			ptrArg := reflect.New(srcValue.Type())
			ptrArg.Elem().Set(srcValue)
			callArg = ptrArg
		} else {
			// Function expects value, we have value
			callArg = srcValue
		}

		result := fnValue.Call([]reflect.Value{callArg})[0]

		if fnType.Out(0).Kind() == reflect.Ptr {
			// Function returns pointer, we need value - dereference
			if result.IsNil() {
				return dst, nil
			}
			return result.Elem().Interface().(D), nil
		} else {
			// Function returns value, we need value
			return result.Interface().(D), nil
		}
	}

	return dst, nil
}

// MustMap is like Map but panics if mapping fails.
// It is useful for cases where mapping failure is considered a programmer error.
//
// Type Parameters:
//   - S: Source type
//   - D: Destination type
//
// Parameters:
//   - m: The mapper instance
//   - src: The source value to be mapped
//
// Returns:
//   - D: The mapped result of type D
//
// Panics:
//   - If no mapping function is registered for the type pair, or if mapping fails
//
// Example:
//
//	mapper := New()
//	Register(mapper, func(s string) int { return len(s) })
//	result := MustMap[string, int](mapper, "hello")
//	fmt.Println(result) // Output: 5
func MustMap[S any, D any](m Mapper, src S) D {
	result, err := Map[S, D](m, src)
	if err != nil {
		panic(err)
	}
	return result
}

// MapSlice applies a registered mapping function to each element of a slice,
// returning a new slice with the mapped elements. It supports various combinations
// of slice element types including values and pointers.
//
// Type Parameters:
//   - S: Source slice type (e.g., []SourceType, []*SourceType)
//   - D: Destination slice type (e.g., []DestType, []*DestType)
//
// Parameters:
//   - m: The mapper instance containing the registered mapping functions
//   - src: The source slice to be mapped
//
// Returns:
//   - D: A new slice containing the mapped elements
//   - error: ErrNoMapping if no mapping function is registered for the element types,
//     or an error if source/destination are not slices
//
// Supported slice mapping combinations:
//   - []T -> []U: Value elements to value elements
//   - []*T -> []*U: Pointer elements to pointer elements (nil elements remain nil)
//   - []T -> []*U: Value elements to pointer elements
//   - []*T -> []U: Pointer elements to value elements (nil elements become zero values)
//
// Example:
//
//	mapper := New()
//	Register(mapper, func(s string) int { return len(s) })
//
//	words := []string{"hello", "world", "go"}
//	lengths, err := MapSlice[[]string, []int](mapper, words)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(lengths) // Output: [5 5 2]
//
//	// Pointer slice mapping
//	wordPtrs := []*string{&words[0], nil, &words[2]}
//	lengthPtrs, err := MapSlice[[]*string, []*int](mapper, wordPtrs)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// lengthPtrs will be [*5, nil, *2]
func MapSlice[S any, D any](m Mapper, src S) (D, error) {
	var dst D

	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)

	// Ensure we're working with slices
	if srcType.Kind() != reflect.Slice || dstType.Kind() != reflect.Slice {
		return dst, ErrSrcAndDestMustBeSlices
	}

	// Get element types for registry lookup
	srcElemType := srcType.Elem()
	dstElemType := dstType.Elem()

	// Remove pointer indirection for the key
	keySrcType := srcElemType
	keyDstType := dstElemType

	if keySrcType.Kind() == reflect.Ptr {
		keySrcType = keySrcType.Elem()
	}
	if keyDstType.Kind() == reflect.Ptr {
		keyDstType = keyDstType.Elem()
	}

	key := typePair{
		src: keySrcType,
		dst: keyDstType,
	}

	fn, ok := m.registry[key]
	if !ok {
		return dst, ErrNoMapping
	}

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	srcValue := reflect.ValueOf(src)
	srcLen := srcValue.Len()

	// Create destination slice
	dstSlice := reflect.MakeSlice(dstType, srcLen, srcLen)

	// Map each element
	for i := 0; i < srcLen; i++ {
		srcElem := srcValue.Index(i)

		// Handle the four cases based on pointer combinations
		var mappedElem reflect.Value

		// Case 1: []A -> []B (value to value)
		if srcElemType.Kind() != reflect.Ptr && dstElemType.Kind() != reflect.Ptr {
			var callArg reflect.Value
			if fnType.In(0).Kind() == reflect.Ptr {
				// Function expects pointer, we have value - create pointer
				ptrArg := reflect.New(srcElem.Type())
				ptrArg.Elem().Set(srcElem)
				callArg = ptrArg
			} else {
				// Function expects value, we have value
				callArg = srcElem
			}

			result := fnValue.Call([]reflect.Value{callArg})[0]

			if fnType.Out(0).Kind() == reflect.Ptr {
				// Function returns pointer, we need value - dereference
				if result.IsNil() {
					mappedElem = reflect.Zero(dstElemType)
				} else {
					mappedElem = result.Elem()
				}
			} else {
				// Function returns value, we need value
				mappedElem = result
			}
		}

		// Case 2: []*A -> []*B (pointer to pointer)
		if srcElemType.Kind() == reflect.Ptr && dstElemType.Kind() == reflect.Ptr {
			if srcElem.IsNil() {
				mappedElem = reflect.Zero(dstElemType) // nil pointer
			} else {
				var callArg reflect.Value
				if fnType.In(0).Kind() == reflect.Ptr {
					// Function expects pointer, we have pointer
					callArg = srcElem
				} else {
					// Function expects value, we have pointer - dereference
					callArg = srcElem.Elem()
				}

				result := fnValue.Call([]reflect.Value{callArg})[0]

				if fnType.Out(0).Kind() == reflect.Ptr {
					// Function returns pointer, we need pointer
					mappedElem = result
				} else {
					// Function returns value, we need pointer - create pointer
					ptrResult := reflect.New(result.Type())
					ptrResult.Elem().Set(result)
					mappedElem = ptrResult
				}
			}
		}

		// Case 3: []*A -> []B (pointer to value)
		if srcElemType.Kind() == reflect.Ptr && dstElemType.Kind() != reflect.Ptr {
			if srcElem.IsNil() {
				mappedElem = reflect.Zero(dstElemType)
			} else {
				var callArg reflect.Value
				if fnType.In(0).Kind() == reflect.Ptr {
					// Function expects pointer, we have pointer
					callArg = srcElem
				} else {
					// Function expects value, we have pointer - dereference
					callArg = srcElem.Elem()
				}

				result := fnValue.Call([]reflect.Value{callArg})[0]

				if fnType.Out(0).Kind() == reflect.Ptr {
					// Function returns pointer, we need value - dereference
					if result.IsNil() {
						mappedElem = reflect.Zero(dstElemType)
					} else {
						mappedElem = result.Elem()
					}
				} else {
					// Function returns value, we need value
					mappedElem = result
				}
			}
		}

		// Case 4: []A -> []*B (value to pointer)
		if srcElemType.Kind() != reflect.Ptr && dstElemType.Kind() == reflect.Ptr {
			var callArg reflect.Value
			if fnType.In(0).Kind() == reflect.Ptr {
				// Function expects pointer, we have value - create pointer
				ptrArg := reflect.New(srcElem.Type())
				ptrArg.Elem().Set(srcElem)
				callArg = ptrArg
			} else {
				// Function expects value, we have value
				callArg = srcElem
			}

			result := fnValue.Call([]reflect.Value{callArg})[0]

			if fnType.Out(0).Kind() == reflect.Ptr {
				// Function returns pointer, we need pointer
				mappedElem = result
			} else {
				// Function returns value, we need pointer - create pointer
				ptrResult := reflect.New(result.Type())
				ptrResult.Elem().Set(result)
				mappedElem = ptrResult
			}
		}

		dstSlice.Index(i).Set(mappedElem)
	}

	return dstSlice.Interface().(D), nil
}

// MustMapSlice is like MapSlice but panics if mapping fails.
// It is useful for cases where mapping failure is considered a programmer error.
//
// Type Parameters:
//   - S: Source slice type (e.g., []SourceType, []*SourceType)
//   - D: Destination slice type (e.g., []DestType, []*DestType)
//
// Parameters:
//   - m: The mapper instance
//   - src: The source slice to be mapped
//
// Returns:
//   - D: The mapped result slice of type D
//
// Panics:
//   - If no mapping function is registered for the element types, or if mapping fails
//
// Example:
//
//	mapper := New()
//	Register(mapper, func(s string) int { return len(s) })
//	result := MustMapSlice[[]string, []int](mapper, []string{"a", "bb"})
//	fmt.Println(result) // Output: [1 2]
func MustMapSlice[S any, D any](m Mapper, src S) D {
	result, err := MapSlice[S, D](m, src)
	if err != nil {
		panic(err)
	}
	return result
}

// Has checks if a mapping function is registered for the specified type pair.
// This is useful for conditional mapping or validation before attempting to map.
//
// Type Parameters:
//   - S: Source type to check for mapping
//   - D: Destination type to check for mapping
//
// Parameters:
//   - m: The mapper instance to check
//
// Returns:
//   - bool: true if a mapping function is registered for S -> D, false otherwise
//
// Example:
//
//	mapper := New()
//	Register(mapper, func(s string) int { return len(s) })
//
//	if Has[string, int](mapper) {
//	    result, _ := Map[string, int](mapper, "hello")
//	    fmt.Println("Mapped:", result)
//	} else {
//	    fmt.Println("No mapping available")
//	}
//
//	fmt.Println(Has[int, string](mapper)) // Output: false
func Has[S any, D any](m Mapper) bool {
	key := typePair{
		src: reflect.TypeOf((*S)(nil)).Elem(),
		dst: reflect.TypeOf((*D)(nil)).Elem(),
	}
	_, ok := m.registry[key]
	return ok
}

// Remove unregisters a mapping function for the specified type pair.
// After removal, attempting to map between these types will return ErrNoMapping.
// This operation is safe to call even if no mapping exists for the type pair.
//
// Type Parameters:
//   - S: Source type to remove mapping for
//   - D: Destination type to remove mapping for
//
// Parameters:
//   - m: The mapper instance to remove the mapping from
//
// Example:
//
//	mapper := New()
//	Register(mapper, func(s string) int { return len(s) })
//
//	fmt.Println(Has[string, int](mapper)) // Output: true
//
//	Remove[string, int](mapper)
//	fmt.Println(Has[string, int](mapper)) // Output: false
//
//	_, err := Map[string, int](mapper, "hello")
//	fmt.Println(err) // Output: no mapping function registered for this type pair
func Remove[S any, D any](m Mapper) {
	key := typePair{
		src: reflect.TypeOf((*S)(nil)).Elem(),
		dst: reflect.TypeOf((*D)(nil)).Elem(),
	}
	delete(m.registry, key)
}

// List returns a slice of strings representing all registered mapping type pairs.
// Each string is formatted as "SourceType-DestinationType" and can be used for
// debugging, logging, or displaying available mappings to users.
//
// Parameters:
//   - m: The mapper instance to list mappings from
//
// Returns:
//   - []string: A slice of strings representing all registered type pair mappings
//
// Example:
//
//	mapper := New()
//	Register(mapper, func(s string) int { return len(s) })
//	Register(mapper, func(i int) string { return fmt.Sprintf("%d", i) })
//	Register(mapper, func(p Person) PersonDTO { return PersonDTO{Name: p.Name} })
//
//	mappings := List(mapper)
//	for _, mapping := range mappings {
//	    fmt.Println("Available mapping:", mapping)
//	}
//	// Output:
//	// Available mapping: string-int
//	// Available mapping: int-string
//	// Available mapping: main.Person-main.PersonDTO
func List(m Mapper) []string {
	keys := make([]string, 0, len(m.registry))
	for k := range m.registry {
		keys = append(keys, k.src.String()+"-"+k.dst.String())
	}
	return keys
}
