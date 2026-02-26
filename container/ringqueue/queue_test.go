package ringqueue

import "testing"

func TestQueue_FIFO(t *testing.T) {
	var q Queue[int]
	for i := range 100 {
		q.PushBack(i)
	}
	for i := range 100 {
		v, ok := q.PopFront()
		if !ok || v != i {
			t.Fatalf("expected %d got %d ok=%v", i, v, ok)
		}
	}
	if q.Len() != 0 {
		t.Fatalf("expected empty")
	}
}

func TestQueue_WrapAndGrowKeepsOrder(t *testing.T) {
	var q Queue[int]
	for i := 0; i < 8; i++ {
		q.PushBack(i)
	}
	for i := 0; i < 6; i++ {
		_, _ = q.PopFront()
	}
	for i := 8; i < 40; i++ {
		q.PushBack(i)
	}

	want := []int{6, 7}
	for i := 8; i < 40; i++ {
		want = append(want, i)
	}

	for i, w := range want {
		v, ok := q.PopFront()
		if !ok || v != w {
			t.Fatalf("idx=%d expected %d got %d ok=%v", i, w, v, ok)
		}
	}
}
