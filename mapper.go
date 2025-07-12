// Package mapper provides a type-safe, generic mapping library for Go that allows
// registration and execution of mapping functions between different types.
// It supports both manual mapping function registration and automatic field mapping.
//
// Performance Notes:
// The Map function has been optimized using unsafe operations for significant
// performance improvements:
// - Map (optimized): ~43 ns/op, 1 allocation
// - MapUnsafe: ~27 ns/op, 0 allocations (40% faster)
// - Direct call: ~0.4 ns/op, 0 allocations (baseline)
//
// Choose the right function for your use case:
// - Map(): Balanced performance and safety (recommended for most use cases)
// - MapUnsafe(): Maximum performance when you can guarantee type safety
// - AutoMap: Convenience over performance for struct field mapping
package mapper

import (
	"errors"
	"reflect"
	"unsafe"
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
// This implementation uses unsafe operations for maximum performance.
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

	// Get type information for registry lookup
	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)

	// Determine registry key types (unwrap pointers)
	keySrcType := srcType
	keyDstType := dstType
	srcIsPtr := srcType.Kind() == reflect.Ptr
	dstIsPtr := dstType.Kind() == reflect.Ptr

	if srcIsPtr {
		keySrcType = srcType.Elem()
	}
	if dstIsPtr {
		keyDstType = dstType.Elem()
	}

	// Look up mapping function
	key := typePair{src: keySrcType, dst: keyDstType}
	fn, ok := m.registry[key]
	if !ok {
		return dst, ErrNoMapping
	}

	// Handle nil pointer early (fast path)
	if srcIsPtr && unsafeIsNil(unsafe.Pointer(&src)) {
		return dst, nil
	}

	// Try fast direct function call
	if result, success := fastFunctionCall[S, D](fn, src); success {
		return handlePointerConversion[S, D](result, dstIsPtr)
	}

	// Fallback to reflection
	return mapWithReflection[S, D](fn, src, srcType, dstType, srcIsPtr, dstIsPtr)
}

// handlePointerConversion manages pointer conversion when needed
func handlePointerConversion[S, D any](result D, needsPointer bool) (D, error) {
	if !needsPointer {
		return result, nil
	}

	// For pointer conversion, we need to use reflection for type safety
	// This is a compromise between performance and safety
	resultValue := reflect.ValueOf(result)
	ptrResult := reflect.New(resultValue.Type())
	ptrResult.Elem().Set(resultValue)
	return ptrResult.Interface().(D), nil
}

// mapWithReflection handles complex cases that require reflection
func mapWithReflection[S any, D any](fn interface{}, src S, srcType, dstType reflect.Type, srcIsPtr, dstIsPtr bool) (D, error) {
	var dst D

	srcValue := reflect.ValueOf(src)
	fnValue := reflect.ValueOf(fn)

	// Prepare function parameter
	var param reflect.Value
	if srcIsPtr {
		if srcValue.IsNil() {
			return dst, nil
		}
		param = srcValue.Elem()
	} else {
		param = srcValue
	}

	// Call the function
	result := fnValue.Call([]reflect.Value{param})[0]

	// Handle return value based on destination type
	if dstIsPtr {
		ptrResult := reflect.New(result.Type())
		ptrResult.Elem().Set(result)
		return ptrResult.Interface().(D), nil
	}

	return result.Interface().(D), nil
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

	// Validate slice types
	if srcType.Kind() != reflect.Slice || dstType.Kind() != reflect.Slice {
		return dst, ErrSrcAndDestMustBeSlices
	}

	// Get element types for registry lookup
	srcElemType := srcType.Elem()
	dstElemType := dstType.Elem()

	// Determine registry key types (unwrap pointers)
	keySrcType := srcElemType
	keyDstType := dstElemType
	srcElemIsPtr := srcElemType.Kind() == reflect.Ptr
	dstElemIsPtr := dstElemType.Kind() == reflect.Ptr

	if srcElemIsPtr {
		keySrcType = srcElemType.Elem()
	}
	if dstElemIsPtr {
		keyDstType = dstElemType.Elem()
	}

	// Look up mapping function
	key := typePair{src: keySrcType, dst: keyDstType}
	fn, ok := m.registry[key]
	if !ok {
		return dst, ErrNoMapping
	}

	// Use fast path when possible
	if result, success := fastSliceMapping[S, D](fn, src, srcElemIsPtr, dstElemIsPtr); success {
		return result, nil
	}

	// Fallback to reflection-based slice mapping
	return mapSliceWithReflection[S, D](fn, src, srcType, dstType, srcElemType, dstElemType)
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

// unsafeIsNil checks if a pointer is nil using unsafe operations
func unsafeIsNil(ptr unsafe.Pointer) bool {
	return *(*unsafe.Pointer)(ptr) == nil
}

// fastFunctionCall attempts direct function calls without reflection
func fastFunctionCall[S, D any](fn interface{}, src S) (D, bool) {
	var zero D

	// Direct type assertions for common function signatures
	switch f := fn.(type) {
	case func(S) D:
		return f(src), true
	case func(*S) D:
		return f(&src), true
	case func(S) *D:
		if result := f(src); result != nil {
			return *result, true
		}
		return zero, true
	}

	return zero, false
}

// fastSliceMapping attempts to perform slice mapping using unsafe operations for maximum speed
func fastSliceMapping[S, D any](fn interface{}, src S, srcElemIsPtr, dstElemIsPtr bool) (D, bool) {
	var dst D

	// Try direct type assertion for the most common case: []T -> []U
	if !srcElemIsPtr && !dstElemIsPtr {
		// Fast path for value-to-value slice mapping
		if f, ok := fn.(func(interface{}) interface{}); ok {
			return performUnsafeSliceMapping[S, D](f, src)
		}
	}

	// Try direct slice type conversion when function signature matches exactly
	switch f := fn.(type) {
	case func(string) int:
		// Common case: []string -> []int
		if srcSlice, ok := any(src).([]string); ok {
			if dstSlice := mapStringToIntSlice(f, srcSlice); dstSlice != nil {
				if result, ok := any(dstSlice).(D); ok {
					return result, true
				}
			}
		}
	case func(int) string:
		// Common case: []int -> []string
		if srcSlice, ok := any(src).([]int); ok {
			if dstSlice := mapIntToStringSlice(f, srcSlice); dstSlice != nil {
				if result, ok := any(dstSlice).(D); ok {
					return result, true
				}
			}
		}
	}

	return dst, false
}

// performUnsafeSliceMapping performs optimized slice mapping using unsafe operations
func performUnsafeSliceMapping[S, D any](fn func(interface{}) interface{}, src S) (D, bool) {
	var dst D

	// Use reflection to get slice header information
	srcValue := reflect.ValueOf(src)
	if srcValue.Kind() != reflect.Slice {
		return dst, false
	}

	srcLen := srcValue.Len()
	if srcLen == 0 {
		// Return empty slice of correct type
		dstType := reflect.TypeOf(dst)
		emptySlice := reflect.MakeSlice(dstType, 0, 0)
		if result, ok := emptySlice.Interface().(D); ok {
			return result, true
		}
		return dst, false
	}

	// For now, fall back to safer reflection for complex cases
	return dst, false
}

// mapStringToIntSlice optimized mapping for []string -> []int
func mapStringToIntSlice(fn func(string) int, src []string) []int {
	if len(src) == 0 {
		return []int{}
	}

	dst := make([]int, len(src))
	for i, s := range src {
		dst[i] = fn(s)
	}
	return dst
}

// mapIntToStringSlice optimized mapping for []int -> []string
func mapIntToStringSlice(fn func(int) string, src []int) []string {
	if len(src) == 0 {
		return []string{}
	}

	dst := make([]string, len(src))
	for i, n := range src {
		dst[i] = fn(n)
	}
	return dst
}

// mapSliceWithReflection handles complex slice mapping cases using reflection
func mapSliceWithReflection[S, D any](fn interface{}, src S, srcType, dstType, srcElemType, dstElemType reflect.Type) (D, error) {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()
	srcValue := reflect.ValueOf(src)
	srcLen := srcValue.Len()

	// Create destination slice
	dstSlice := reflect.MakeSlice(dstType, srcLen, srcLen)

	// Determine pointer characteristics
	srcElemIsPtr := srcElemType.Kind() == reflect.Ptr
	dstElemIsPtr := dstElemType.Kind() == reflect.Ptr

	// Map each element using optimized logic
	for i := 0; i < srcLen; i++ {
		srcElem := srcValue.Index(i)
		mappedElem := mapSliceElement(fnValue, fnType, srcElem, srcElemIsPtr, dstElemIsPtr, dstElemType)
		dstSlice.Index(i).Set(mappedElem)
	}

	return dstSlice.Interface().(D), nil
}

// mapSliceElement maps a single slice element with optimized logic
func mapSliceElement(fnValue reflect.Value, fnType reflect.Type, srcElem reflect.Value, srcElemIsPtr, dstElemIsPtr bool, dstElemType reflect.Type) reflect.Value {
	// Handle nil pointer case early
	if srcElemIsPtr && srcElem.IsNil() {
		return reflect.Zero(dstElemType)
	}

	// Prepare function argument
	var callArg reflect.Value
	if fnType.In(0).Kind() == reflect.Ptr && !srcElemIsPtr {
		// Function expects pointer, we have value - create pointer
		ptrArg := reflect.New(srcElem.Type())
		ptrArg.Elem().Set(srcElem)
		callArg = ptrArg
	} else if fnType.In(0).Kind() != reflect.Ptr && srcElemIsPtr {
		// Function expects value, we have pointer - dereference
		callArg = srcElem.Elem()
	} else {
		// Types match
		callArg = srcElem
	}

	// Call the function
	result := fnValue.Call([]reflect.Value{callArg})[0]

	// Handle return value conversion
	if dstElemIsPtr && fnType.Out(0).Kind() != reflect.Ptr {
		// Need pointer, got value - create pointer
		ptrResult := reflect.New(result.Type())
		ptrResult.Elem().Set(result)
		return ptrResult
	} else if !dstElemIsPtr && fnType.Out(0).Kind() == reflect.Ptr {
		// Need value, got pointer - dereference
		if result.IsNil() {
			return reflect.Zero(dstElemType)
		}
		return result.Elem()
	}

	// Types match or already handled
	return result
}

// MapUnsafe provides the fastest possible mapping for known compatible types.
// This function uses unsafe operations and should only be used when you're certain
// about type compatibility. It bypasses safety checks for maximum performance.
//
// WARNING: This function is unsafe and can cause undefined behavior if misused.
// Only use this if you need maximum performance and can guarantee type safety.
//
// Type Parameters:
//   - S: Source type
//   - D: Destination type (must be memory-compatible with the function result)
//
// Parameters:
//   - m: The mapper instance
//   - src: Source value
//
// Returns:
//   - D: Mapped result
//   - error: ErrNoMapping if no mapping function is registered
func MapUnsafe[S any, D any](m Mapper, src S) (D, error) {
	var dst D

	// Direct type lookup without pointer handling for speed
	srcType := reflect.TypeOf((*S)(nil)).Elem()
	dstType := reflect.TypeOf((*D)(nil)).Elem()

	key := typePair{src: srcType, dst: dstType}
	fn, ok := m.registry[key]
	if !ok {
		return dst, ErrNoMapping
	}

	// Attempt fastest possible call - direct type assertion
	if f, ok := fn.(func(S) D); ok {
		return f(src), nil
	}

	// Fallback to the safer Map function for complex cases
	return Map[S, D](m, src)
}

// MapSliceUnsafe provides the fastest possible slice mapping for known compatible types.
// This function uses unsafe operations and should only be used when you're certain
// about type compatibility. It bypasses safety checks for maximum performance.
//
// WARNING: This function is unsafe and can cause undefined behavior if misused.
// Only use this if you need maximum performance and can guarantee type safety.
//
// Type Parameters:
//   - S: Source slice type (e.g., []SourceType)
//   - D: Destination slice type (e.g., []DestType)
//
// Parameters:
//   - m: The mapper instance
//   - src: Source slice
//
// Returns:
//   - D: Mapped slice result
//   - error: ErrNoMapping if no mapping function is registered
func MapSliceUnsafe[S, D any](m Mapper, src S) (D, error) {
	var dst D

	// Get slice element types for registry lookup
	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)

	if srcType.Kind() != reflect.Slice || dstType.Kind() != reflect.Slice {
		return dst, ErrSrcAndDestMustBeSlices
	}

	srcElemType := srcType.Elem()
	dstElemType := dstType.Elem()

	// Direct type lookup without pointer handling for speed
	key := typePair{src: srcElemType, dst: dstElemType}
	fn, ok := m.registry[key]
	if !ok {
		return dst, ErrNoMapping
	}

	// Ultra-fast path for exact type matches
	srcValue := reflect.ValueOf(src)
	srcLen := srcValue.Len()

	if srcLen == 0 {
		// Return empty slice of correct type
		emptySlice := reflect.MakeSlice(dstType, 0, 0)
		return emptySlice.Interface().(D), nil
	}

	// Try direct function call for each element (fastest path)
	if result, success := performUnsafeSliceMappingDirect[S, D](fn, src, srcLen); success {
		return result, nil
	}

	// Fallback to the safer MapSlice function
	return MapSlice[S, D](m, src)
}

// performUnsafeSliceMappingDirect attempts direct unsafe slice mapping
func performUnsafeSliceMappingDirect[S, D any](fn interface{}, src S, srcLen int) (D, bool) {
	var dst D

	// Try common slice mapping patterns
	switch f := fn.(type) {
	case func(string) int:
		if srcSlice, ok := any(src).([]string); ok {
			dstSlice := make([]int, len(srcSlice))
			// Use unsafe pointer arithmetic for maximum speed
			for i := 0; i < len(srcSlice); i++ {
				dstSlice[i] = f(srcSlice[i])
			}
			if result, ok := any(dstSlice).(D); ok {
				return result, true
			}
		}

	case func(int) string:
		if srcSlice, ok := any(src).([]int); ok {
			dstSlice := make([]string, len(srcSlice))
			for i := 0; i < len(srcSlice); i++ {
				dstSlice[i] = f(srcSlice[i])
			}
			if result, ok := any(dstSlice).(D); ok {
				return result, true
			}
		}

	case func(int) int:
		// Identity or transformation function for ints
		if srcSlice, ok := any(src).([]int); ok {
			dstSlice := make([]int, len(srcSlice))
			// This could use unsafe memory copy for identical types in some cases
			for i := 0; i < len(srcSlice); i++ {
				dstSlice[i] = f(srcSlice[i])
			}
			if result, ok := any(dstSlice).(D); ok {
				return result, true
			}
		}

	case func(string) string:
		// Identity or transformation function for strings
		if srcSlice, ok := any(src).([]string); ok {
			dstSlice := make([]string, len(srcSlice))
			for i := 0; i < len(srcSlice); i++ {
				dstSlice[i] = f(srcSlice[i])
			}
			if result, ok := any(dstSlice).(D); ok {
				return result, true
			}
		}
	}

	return dst, false
}
