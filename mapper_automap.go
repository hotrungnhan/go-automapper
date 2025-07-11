// Package mapper provides automatic mapping functionality using reflection and the jinzhu/copier library
package mapper

import (
	"github.com/jinzhu/copier"
	"reflect"
)

// autoMap performs automatic mapping between source and destination types using reflection.
// It uses the jinzhu/copier library to copy matching fields between structs.
// If the source and destination are of the same type, it performs a direct assignment for optimization.
//
// This function is used internally by RegisterAutoMap and is not part of the public API.
//
// Type Parameters:
//   - S: Source type to map from
//   - D: Destination type to map to
//
// Parameters:
//   - src: The source value to map from
//
// Returns:
//   - D: The mapped destination value with copied fields
func autoMap[S any, D any](src S) D {
	var dst D

	// Fast path: if src and dst are the same type, just assign
	if reflect.TypeOf(src) == reflect.TypeOf(dst) {
		anyDst := any(&dst)
		anySrc := any(&src)
		reflect.ValueOf(anyDst).Elem().Set(reflect.ValueOf(anySrc).Elem())
		return dst
	}

	// Avoid unnecessary pointer conversions
	srcPtr := any(&src)
	dstPtr := any(&dst)
	_ = copier.Copy(dstPtr, srcPtr)
	return dst
}

// RegisterAutoMap registers bidirectional automatic mapping functions for types S and D.
// This function creates mapping functions that use reflection to automatically copy
// matching fields between structs. Both S->D and D->S mappings are registered.
//
// The automatic mapping uses the jinzhu/copier library, which copies fields with matching
// names and compatible types. This is convenient for mapping between similar structs
// but comes with a performance cost compared to manually registered functions.
//
// Performance Note: AutoMap functions are approximately 50x slower than manually
// registered mapping functions (~1600ns vs ~25ns per operation) due to reflection overhead.
//
// Type Parameters:
//   - S: Source type for bidirectional mapping
//   - D: Destination type for bidirectional mapping
//
// Parameters:
//   - m: The mapper instance to register the automatic mapping functions with
//
// Example:
//
//	type User struct {
//	    ID   int    `json:"id"`
//	    Name string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	type UserDTO struct {
//	    ID   int    `json:"id"`
//	    Name string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	mapper := New()
//	RegisterAutoMap[User, UserDTO](mapper)
//
//	user := User{ID: 1, Name: "John", Email: "john@example.com"}
//	dto, err := Map[User, UserDTO](mapper, user)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Mapped DTO: %+v\n", dto)
//
//	// Reverse mapping is also available
//	backToUser, err := Map[UserDTO, User](mapper, dto)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Mapped back: %+v\n", backToUser)
func RegisterAutoMap[S any, D any](m Mapper) {
	key := typePair{
		src: reflect.TypeOf((*S)(nil)).Elem(),
		dst: reflect.TypeOf((*D)(nil)).Elem(),
	}
	m.registry[key] = autoMap[S, D]

	// reverse mapping
	key = typePair{
		src: reflect.TypeOf((*D)(nil)).Elem(),
		dst: reflect.TypeOf((*S)(nil)).Elem(),
	}
	m.registry[key] = autoMap[D, S]
}
