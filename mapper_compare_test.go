package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Source struct {
	Name  string
	Age   int
	Email string
}
type Dest struct {
	Name  string
	Age   int
	Email string
}

func mapFn(s Source) Dest {
	return Dest{
		Name:  s.Name,
		Age:   s.Age,
		Email: s.Email,
	}
}

// Benchmark comparison: NoLibrary vs DirectMap vs AutoMap
func BenchmarkMapStructCompare(b *testing.B) {
	src := Source{Name: "John Doe", Age: 30, Email: "john@example.com"}
	b.Run("Manual", func(b *testing.B) {
		var dst Dest

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dst = mapFn(src)
		}
		b.StopTimer()
		assert.Equal(b, "John Doe", dst.Name)
		assert.Equal(b, 30, dst.Age)
		assert.Equal(b, "john@example.com", dst.Email)
	})

	b.Run("DirectMap", func(b *testing.B) {
		mapper := New()
		var dst Dest
		b.ResetTimer()
		Register(mapper, mapFn)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dst, _ = Map[Source, Dest](mapper, src)
		}
		b.StopTimer()
		assert.Equal(b, "John Doe", dst.Name)
		assert.Equal(b, 30, dst.Age)
		assert.Equal(b, "john@example.com", dst.Email)
	})

	b.Run("AutoMap", func(b *testing.B) {
		mapper := New()
		var dst Dest
		RegisterAutoMap[Source, Dest](mapper)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dst, _ = Map[Source, Dest](mapper, src)
		}
		b.StopTimer()
		assert.Equal(b, "John Doe", dst.Name)
		assert.Equal(b, 30, dst.Age)
		assert.Equal(b, "john@example.com", dst.Email)
	})
}

func BenchmarkMapSliceBigCompare(b *testing.B) {
	const size = 1000
	srcSlice := make([]Source, size)
	for i := 0; i < size; i++ {
		srcSlice[i] = Source{Name: "John Doe", Age: 30, Email: "john@example.com"}
	}

	b.Run("ManualSliceMap", func(b *testing.B) {
		var destSlice []Dest
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			destSlice = make([]Dest, len(srcSlice))
			for j, s := range srcSlice {
				destSlice[j] = Dest{Name: s.Name, Age: s.Age, Email: s.Email}
			}
		}
		b.StopTimer()
		assert.Equal(b, size, len(destSlice))
		assert.Equal(b, "John Doe", destSlice[0].Name)
		assert.Equal(b, 30, destSlice[0].Age)
		assert.Equal(b, "john@example.com", destSlice[0].Email)
	})

	b.Run("DirectMapSlice", func(b *testing.B) {
		mapper := New()
		Register(mapper, mapFn)
		var destSlice []Dest
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			destSlice = make([]Dest, len(srcSlice))
			for j, s := range srcSlice {
				d, _ := Map[Source, Dest](mapper, s)
				destSlice[j] = d
			}
		}
		b.StopTimer()
		assert.Equal(b, size, len(destSlice))
		assert.Equal(b, "John Doe", destSlice[0].Name)
		assert.Equal(b, 30, destSlice[0].Age)
		assert.Equal(b, "john@example.com", destSlice[0].Email)
	})

	b.Run("AutoMapSlice", func(b *testing.B) {
		mapper := New()
		RegisterAutoMap[Source, Dest](mapper)
		var destSlice []Dest
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			destSlice = make([]Dest, len(srcSlice))
			for j, s := range srcSlice {
				d, _ := Map[Source, Dest](mapper, s)
				destSlice[j] = d
			}
		}
		b.StopTimer()
		assert.Equal(b, size, len(destSlice))
		assert.Equal(b, "John Doe", destSlice[0].Name)
		assert.Equal(b, 30, destSlice[0].Age)
		assert.Equal(b, "john@example.com", destSlice[0].Email)
	})
}
