package mlfq

import (
	"context"
	"testing"
	"time"
)

func BenchmarkScheduler_Submit(b *testing.B) {
	ctx := context.Background()
	s, _ := NewDefault[int](64, WithClock(func() time.Time { return time.Unix(0, 0) }))
	defer s.Close()

	for i := 0; b.Loop(); i++ {
		_, _ = s.Submit(ctx, i)
	}
}

func BenchmarkScheduler_SteadyState_NextFeedback_Requeue(b *testing.B) {
	ctx := context.Background()
	s, _ := NewDefault[int](64, WithClock(func() time.Time { return time.Unix(0, 0) }))
	defer s.Close()

	// Preload some tasks to keep queues non-empty.
	const preload = 4096
	for i := range preload {
		_, _ = s.Submit(ctx, i)
	}

	for b.Loop() {
		lease, ok := s.Next(ctx)
		if !ok {
			b.Fatalf("unexpected empty")
		}
		_ = s.FeedBack(ctx, lease.Token, Feedback{
			RanFor:          lease.Quantum,
			Finished:        false,
			UsedFullQuantum: true, // force demotion path
		})
	}
}

func BenchmarkScheduler_Tick_OLevels(b *testing.B) {
	ctx := context.Background()
	clock := time.Unix(100, 0)
	s, _ := New(
		NewDefaultPolicy[int](4096, DefaultPolicyConfig{AgingThreshold: 1 * time.Second}),
		WithClock(func() time.Time { return clock }),
	)
	defer s.Close()

	// Ensure each level has at least one task so Tick scans the whole bitmap path.
	for level := range 4096 {
		_, _ = s.Submit(ctx, level, WithAttributes(Attributes{Urgency: 0, Importance: 0}))
		// Force enqueue into lowest-ish levels by feedback demotion.
		lease, ok := s.Next(ctx)
		if !ok {
			b.Fatalf("unexpected empty")
		}
		_ = s.FeedBack(ctx, lease.Token, Feedback{RanFor: lease.Quantum, UsedFullQuantum: true})
	}

	clock = time.Unix(200, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Tick(ctx, clock)
	}
}

func BenchmarkScheduler_Parallel_NextFeedback(b *testing.B) {
	ctx := context.Background()
	s, _ := NewDefault[int](64)
	defer s.Close()

	const preload = 1 << 14
	for i := range preload {
		_, _ = s.Submit(ctx, i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lease, ok := s.Next(ctx)
			if !ok {
				continue
			}
			_ = s.FeedBack(ctx, lease.Token, Feedback{
				RanFor:          lease.Quantum,
				Finished:        false,
				UsedFullQuantum: true,
			})
		}
	})
}
