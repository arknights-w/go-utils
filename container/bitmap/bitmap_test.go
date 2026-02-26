package bitmap

import "testing"

func TestBitMap_MinMax_Empty(t *testing.T) {
	b := New(130)
	if b.Any() {
		t.Fatalf("expected Any=false")
	}
	if _, ok := b.Min(); ok {
		t.Fatalf("expected Min ok=false")
	}
	if _, ok := b.Max(); ok {
		t.Fatalf("expected Max ok=false")
	}
}

func TestBitMap_SetClear_MinMax_Non64Multiple(t *testing.T) {
	cases := []int{1, 63, 65, 130}
	for _, n := range cases {
		b := New(n)
		b.Set(0)
		min, _ := b.Min()
		max, _ := b.Max()
		if min != 0 || max != 0 {
			t.Fatalf("n=%d: expected min=max=0 got min=%d max=%d", n, min, max)
		}

		last := n - 1
		b.Set(last)
		min, _ = b.Min()
		max, _ = b.Max()
		if min != 0 || max != last {
			t.Fatalf("n=%d: expected min=0 max=%d got min=%d max=%d", n, last, min, max)
		}

		b.Clear(0)
		min, _ = b.Min()
		max, _ = b.Max()
		if min != last || max != last {
			t.Fatalf("n=%d: expected min=max=%d got min=%d max=%d", n, last, min, max)
		}
	}
}

func TestBitMap_WordBoundary(t *testing.T) {
	b := New(130)
	b.Set(64)
	if min, _ := b.Min(); min != 64 {
		t.Fatalf("expected min=64 got %d", min)
	}
	if max, _ := b.Max(); max != 64 {
		t.Fatalf("expected max=64 got %d", max)
	}
	b.Set(127)
	if min, _ := b.Min(); min != 64 {
		t.Fatalf("expected min=64 got %d", min)
	}
	if max, _ := b.Max(); max != 127 {
		t.Fatalf("expected max=127 got %d", max)
	}
}

func TestBitMap_nonZeroWords(t *testing.T) {
	b := New(128)
	if b.nonZeroWords != 0 {
		t.Fatalf("expected nonZeroWords=0")
	}
	b.Set(0)
	if b.nonZeroWords != 1 {
		t.Fatalf("expected nonZeroWords=1 got %d", b.nonZeroWords)
	}
	b.Set(1)
	if b.nonZeroWords != 1 {
		t.Fatalf("expected nonZeroWords stay 1 got %d", b.nonZeroWords)
	}
	b.Set(64)
	if b.nonZeroWords != 2 {
		t.Fatalf("expected nonZeroWords=2 got %d", b.nonZeroWords)
	}
	b.Clear(1)
	if b.nonZeroWords != 2 {
		t.Fatalf("expected nonZeroWords stay 2 got %d", b.nonZeroWords)
	}
	b.Clear(0)
	if b.nonZeroWords != 1 {
		t.Fatalf("expected nonZeroWords=1 got %d", b.nonZeroWords)
	}
	b.Clear(64)
	if b.nonZeroWords != 0 {
		t.Fatalf("expected nonZeroWords=0 got %d", b.nonZeroWords)
	}
}
