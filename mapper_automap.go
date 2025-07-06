package mapper

import (
	"github.com/jinzhu/copier"
	"reflect"
)

func AutoMap[S any, D any](src S) D {
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

// it
func RegisterAutoMap[S any, D any](m Mapper) {
	key := typePair{
		src: reflect.TypeOf((*S)(nil)).Elem(),
		dst: reflect.TypeOf((*D)(nil)).Elem(),
	}
	m.registry[key] = AutoMap[S, D]

	// reverse mapping
	key = typePair{
		src: reflect.TypeOf((*D)(nil)).Elem(),
		dst: reflect.TypeOf((*S)(nil)).Elem(),
	}
	m.registry[key] = AutoMap[D, S]
}
