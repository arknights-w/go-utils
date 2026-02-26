package ringqueue

import (
	"strconv"
	"testing"
)

func BenchmarkQueue_PushPop(b *testing.B) {
	for _, n := range []int{64, 4096, 65536, 16777216} {
		b.Run("n="+strconv.Itoa(n), func(b *testing.B) {
			var q Queue[int]
			for i := range n {
				q.PushBack(i)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				q.PushBack(i)
				_, _ = q.PopFront()
			}
		})
	}
}
