package mapper

import (
	"reflect"
	"testing"
)

type myStruct struct {
	i int64
}

type Inccer interface {
	Inc()
}

func (s *myStruct) Inc() {
	s.i++
}

func BenchmarkReflectMethodByNameInterface(b *testing.B) {
	i := new(myStruct)
	b.ResetTimer()
	s := reflect.ValueOf(i)

	for n := 0; n < b.N; n++ {
		f := s.MethodByName("Inc").Interface().(func())
		f()
	}
}

func BenchmarkReflectMethodByName(b *testing.B) {
	i := new(myStruct)
	b.ResetTimer()
	s := reflect.ValueOf(i)

	for n := 0; n < b.N; n++ {
		m := s.MethodByName("Inc")
		m.Call(nil)
	}
}

func BenchmarkReflectMethodCall(b *testing.B) {
	i := new(myStruct)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		reflect.ValueOf(i.Inc).Call(nil)
	}
}

func BenchmarkReflectOnceMethodCall(b *testing.B) {
	i := new(myStruct)
	fn := reflect.ValueOf(i.Inc)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		fn.Call(nil)
	}
}

func BenchmarkReflectCallInterface(b *testing.B) {
	i := new(myStruct)
	fn := reflect.ValueOf(i.Inc).Interface().(func())
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		fn()
	}
}

func BenchmarkStructMethodCall(b *testing.B) {
	i := new(myStruct)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		i.Inc()
	}
}

func BenchmarkInterfaceMethodCall(b *testing.B) {
	var s Inccer = new(myStruct)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s.Inc()
	}
}

func BenchmarkTypeSwitchMethodCall(b *testing.B) {
	var s Inccer = new(myStruct)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		switch v := s.(type) {
		case *myStruct:
			v.Inc()
		}
	}
}

func BenchmarkTypeAssertionMethodCall(b *testing.B) {
	var s Inccer = new(myStruct)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if ms, ok := s.(*myStruct); ok {
			ms.Inc()
		}
	}
}

func BenchmarkReflectElemMethodCall(b *testing.B) {
	i := new(myStruct)
	ptrValue := reflect.ValueOf(i)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		elem := ptrValue.Elem()
		_ = elem
	}
}
func BenchmarkReflectNewAndSet(b *testing.B) {
	type Data struct {
		Value int
	}
	result := reflect.ValueOf(Data{Value: 42})

	for n := 0; n < b.N; n++ {
		ptrResult := reflect.New(result.Type())
		ptrResult.Elem().Set(result)

		_ = ptrResult.Interface().(*Data)
	}
}

func BenchmarkReflectTypeAssertion(b *testing.B) {
	type D struct {
		Value int
	}
	result := reflect.ValueOf(D{Value: 42})

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = result.Interface().(D)
	}
}
