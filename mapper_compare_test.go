package mapper

import "testing"

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
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Dest{
				Name:  src.Name,
				Age:   src.Age,
				Email: src.Email,
			}
		}
	})
	b.Run("DirectMap", func(b *testing.B) {
		mapper := New()
		Register(mapper, func(s Source) Dest {
			return Dest{
				Name:  s.Name,
				Age:   s.Age,
				Email: s.Email,
			}
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Map[Source, Dest](mapper, src)
		}
	})

	b.Run("AutoMap", func(b *testing.B) {
		mapper := New()
		RegisterAutoMap[Source, Dest](mapper)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Map[Source, Dest](mapper, src)
		}
	})

}

func BenchmarkMapSliceCompare(b *testing.B) {
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
		for i := 0; i < b.N; i++ {
			destSlice := make([]Dest, len(srcSlice))
			for j, s := range srcSlice {
				destSlice[j] = Dest{ID: s.ID, Value: s.Value}
			}
		}
	})

	b.Run("DirectMapSlice", func(b *testing.B) {
		mapper := New()
		Register(mapper, func(s Source) Dest {
			return Dest{ID: s.ID, Value: s.Value}
		})
		for i := 0; i < b.N; i++ {
			destSlice := make([]Dest, len(srcSlice))
			for j, s := range srcSlice {
				d, _ := Map[Source, Dest](mapper, s)
				destSlice[j] = d
			}
		}
	})

	b.Run("AutoMapSlice", func(b *testing.B) {
		mapper := New()
		RegisterAutoMap[Source, Dest](mapper)
		for i := 0; i < b.N; i++ {
			destSlice := make([]Dest, len(srcSlice))
			for j, s := range srcSlice {
				d, _ := Map[Source, Dest](mapper, s)
				destSlice[j] = d
			}
		}
	})
}
