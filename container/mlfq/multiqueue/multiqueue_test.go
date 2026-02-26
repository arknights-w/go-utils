package multiqueue

import "testing"

func TestMultiQueue_BitmapSetClear(t *testing.T) {
	m := New[int](5)
	if _, ok := m.MinNonEmpty(); ok {
		t.Fatalf("expected empty")
	}
	m.Push(3, 10)
	if min, _ := m.MinNonEmpty(); min != 3 {
		t.Fatalf("expected min=3 got %d", min)
	}
	_, _ = m.Pop(3)
	if _, ok := m.MinNonEmpty(); ok {
		t.Fatalf("expected empty after pop")
	}
}

func TestMultiQueue_MixedLevels(t *testing.T) {
	m := New[int](4)
	m.Push(2, 1)
	m.Push(0, 2)
	m.Push(3, 3)

	if min, _ := m.MinNonEmpty(); min != 0 {
		t.Fatalf("expected min=0 got %d", min)
	}
	if max, _ := m.MaxNonEmpty(); max != 3 {
		t.Fatalf("expected max=3 got %d", max)
	}

	v, _ := m.Pop(0)
	if v != 2 {
		t.Fatalf("expected pop=2 got %d", v)
	}
	if min, _ := m.MinNonEmpty(); min != 2 {
		t.Fatalf("expected min=2 got %d", min)
	}
}
