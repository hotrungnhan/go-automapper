package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Benchmark comparison: NoLibrary vs DirectMap vs AutoMap
func BenchmarkMapStructCompare(b *testing.B) {
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

	src := Source{Name: "John Doe", Age: 30, Email: "john@example.com"}
	b.Run("Manual", func(b *testing.B) {
		var dst Dest

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dst = Dest{
				Name:  src.Name,
				Age:   src.Age,
				Email: src.Email,
			}
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
		Register(mapper, func(s Source) Dest {
			return Dest{
				Name:  s.Name,
				Age:   s.Age,
				Email: s.Email,
			}
		})

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
	type Source struct {
		ID    int
		Value string
	}
	type Dest struct {
		ID    int
		Value string
	}

	const size = 1000
	srcSlice := make([]Source, size)
	for i := 0; i < size; i++ {
		srcSlice[i] = Source{ID: i, Value: "val"}
	}

	b.Run("ManualSliceMap", func(b *testing.B) {
		destSlice := make([]Dest, len(srcSlice))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j, s := range srcSlice {
				destSlice[j] = Dest{ID: s.ID, Value: s.Value}
			}
		}
		b.StopTimer()
		assert.Equal(b, size, len(destSlice))
		assert.Equal(b, "val", destSlice[0].Value)
	})

	b.Run("DirectMapSlice", func(b *testing.B) {
		mapper := New()
		Register(mapper, func(s Source) Dest {
			return Dest{ID: s.ID, Value: s.Value}
		})
		destSlice := make([]Dest, len(srcSlice))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j, s := range srcSlice {
				d, _ := Map[Source, Dest](mapper, s)
				destSlice[j] = d
			}
		}
		b.StopTimer()
		assert.Equal(b, size, len(destSlice))
		assert.Equal(b, "val", destSlice[0].Value)
	})

	b.Run("AutoMapSlice", func(b *testing.B) {
		mapper := New()
		RegisterAutoMap[Source, Dest](mapper)
		destSlice := make([]Dest, len(srcSlice))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j, s := range srcSlice {
				d, _ := Map[Source, Dest](mapper, s)
				destSlice[j] = d
			}
		}
		b.StopTimer()
		assert.Equal(b, size, len(destSlice))
		assert.Equal(b, "val", destSlice[0].Value)
	})
}

func BenchmarkMapSliceSmallCompare(b *testing.B) {
	type Source struct {
		ID    int
		Value string
	}
	type Dest struct {
		ID    int
		Value string
	}

	const size = 1
	srcSlice := make([]Source, size)
	for i := 0; i < size; i++ {
		srcSlice[i] = Source{ID: i, Value: "val"}
	}

	b.Run("ManualSliceMap", func(b *testing.B) {
		var destSlice []Dest
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			destSlice = make([]Dest, len(srcSlice))

			for j, s := range srcSlice {
				destSlice[j] = Dest{ID: s.ID, Value: s.Value}
			}
		}
		b.StopTimer()
		assert.Equal(b, size, len(destSlice))
		assert.Equal(b, "val", destSlice[0].Value)
		assert.Equal(b, 0, destSlice[0].ID)
	})

	b.Run("DirectMapSlice", func(b *testing.B) {
		mapper := New()
		Register(mapper, func(s Source) Dest {
			return Dest{ID: s.ID, Value: s.Value}
		})

		var destSlice []Dest
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			destSlice, _ = MapSlice[[]Source, []Dest](mapper, srcSlice)
		}
		b.StopTimer()
		assert.Equal(b, size, len(destSlice))
		assert.Equal(b, "val", destSlice[0].Value)
		assert.Equal(b, 0, destSlice[0].ID)
	})

	b.Run("AutoMapSlice", func(b *testing.B) {
		mapper := New()
		RegisterAutoMap[Source, Dest](mapper)
		var destSlice []Dest
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			destSlice, _ = MapSlice[[]Source, []Dest](mapper, srcSlice)
		}
		b.StopTimer()
		assert.Equal(b, size, len(destSlice))
		assert.Equal(b, "val", destSlice[0].Value)
		assert.Equal(b, 0, destSlice[0].ID)
	})
}
