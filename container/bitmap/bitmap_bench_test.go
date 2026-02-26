package bitmap

import (
	"strconv"
	"testing"
)

func BenchmarkBitMap_SetClear(b *testing.B) {
	for _, n := range []int{64, 256, 4096} {
		b.Run("n="+strconv.Itoa(n), func(b *testing.B) {
			bm := New(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				k := i % n
				bm.Set(k)
				bm.Clear(k)
			}
		})
	}
}

func BenchmarkBitMap_MinMax_Sparse(b *testing.B) {
	for _, n := range []int{64, 256, 4096} {
		b.Run("n="+strconv.Itoa(n), func(b *testing.B) {
			bm := New(n)
			bm.Set(n - 1)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = bm.Min()
				_, _ = bm.Max()
			}
		})
	}
}

func BenchmarkBitMap_MinMax_Dense(b *testing.B) {
	for _, n := range []int{64, 256, 4096} {
		b.Run("n="+strconv.Itoa(n), func(b *testing.B) {
			bm := New(n)
			for i := 0; i < n; i++ {
				bm.Set(i)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = bm.Min()
				_, _ = bm.Max()
			}
		})
	}
}
