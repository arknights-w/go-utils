package multiqueue

import (
	"strconv"
	"testing"
)

func BenchmarkMultiQueue_PushPop(b *testing.B) {
	for _, levels := range []int{8, 64, 512} {
		b.Run("levels="+strconv.Itoa(levels), func(b *testing.B) {
			mq := New[int](levels)
			for i := 0; i < levels; i++ {
				mq.Push(i, i)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l := i % levels
				mq.Push(l, i)
				_, _ = mq.Pop(l)
			}
		})
	}
}

func BenchmarkMultiQueue_MinNonEmpty(b *testing.B) {
	for _, levels := range []int{8, 64, 512, 4096} {
		b.Run("levels="+strconv.Itoa(levels), func(b *testing.B) {
			mq := New[int](levels)
			mq.Push(levels-1, 1)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = mq.MinNonEmpty()
			}
		})
	}
}
