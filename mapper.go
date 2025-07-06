package mapper

import (
	"errors"
	"reflect"
)

var Global = NewMapper()

type typePair struct {
	src reflect.Type
	dst reflect.Type
}

type Mapper struct {
	registry map[typePair]interface{}
}

var ErrNoMapping = errors.New("no mapping function registered for this type pair")

func NewMapper() Mapper {
	return Mapper{
		registry: make(map[typePair]interface{}),
	}
}

func Register[S any, D any](m Mapper, fn func(S) D) {
	key := typePair{
		src: reflect.TypeOf((*S)(nil)).Elem(),
		dst: reflect.TypeOf((*D)(nil)).Elem(),
	}
	m.registry[key] = fn
}

func Map[S any, D any](m Mapper, src S) (D, error) {
	key := typePair{
		src: reflect.TypeOf(src),
		dst: reflect.TypeOf((*D)(nil)).Elem(),
	}
	if fn, ok := m.registry[key]; ok {
		return fn.(func(S) D)(src), nil
	}
	var zero D
	return zero, ErrNoMapping
}

func HasMapping[S any, D any](m Mapper) bool {
	key := typePair{
		src: reflect.TypeOf((*S)(nil)).Elem(),
		dst: reflect.TypeOf((*D)(nil)).Elem(),
	}
	_, ok := m.registry[key]
	return ok
}

func RemoveMapping[S any, D any](m Mapper) {
	key := typePair{
		src: reflect.TypeOf((*S)(nil)).Elem(),
		dst: reflect.TypeOf((*D)(nil)).Elem(),
	}
	delete(m.registry, key)
}

func List(m Mapper) []string {
	keys := make([]string, 0, len(m.registry))
	for k := range m.registry {
		keys = append(keys, k.src.String()+"-"+k.dst.String())
	}
	return keys
}
