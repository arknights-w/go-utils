package mlfq

import (
	"context"
	"testing"
	"time"
)

func TestScheduler_SubmitNextFeedbackFinished(t *testing.T) {
	ctx := context.Background()
	s, err := NewDefault[string](4, WithClock(func() time.Time { return time.Unix(0, 0) }))
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	defer s.Close()
	_, _ = s.Submit(ctx, "a")
	lease, ok := s.Next(ctx)
	if !ok {
		t.Fatalf("expected lease")
	}
	if lease.Task != "a" {
		t.Fatalf("expected task a got %q", lease.Task)
	}
	if err := s.FeedBack(ctx, lease.Token, Feedback{RanFor: lease.Quantum, Finished: true}); err != nil {
		t.Fatalf("feedback: %v", err)
	}
	if _, ok := s.Next(ctx); ok {
		t.Fatalf("expected empty")
	}
}

func TestScheduler_FeedBack_NotLeased(t *testing.T) {
	ctx := context.Background()
	s, _ := NewDefault[string](4)
	defer s.Close()
	tok, _ := s.Submit(ctx, "a")
	if err := s.FeedBack(ctx, tok, Feedback{Finished: true}); err != ErrNotLeased {
		t.Fatalf("expected ErrNotLeased got %v", err)
	}
}

func TestScheduler_TickAgingPromotesHead(t *testing.T) {
	now := time.Unix(10, 0)
	clock := now
	s, _ := New[string](NewDefaultPolicy[string](3, DefaultPolicyConfig{
		BaseQuantum:    10 * time.Millisecond,
		MaxQuantum:     100 * time.Millisecond,
		AgingThreshold: 5 * time.Second,
	}), WithClock(func() time.Time { return clock }))
	defer s.Close()

	ctx := context.Background()
	// low priority tasks (level 2) enqueued at t=10
	_, _ = s.Submit(ctx, "a", WithAttributes(Attributes{Urgency: 0, Importance: 0}))
	_, _ = s.Submit(ctx, "b", WithAttributes(Attributes{Urgency: 0, Importance: 0}))

	// Fast-forward: now=20, waited 10s >= threshold => promote from level 2 to 1
	clock = time.Unix(20, 0)
	s.Tick(ctx, clock)

	st := s.Stats(ctx)
	if st.AgingPromoted == 0 {
		t.Fatalf("expected aging promoted > 0")
	}
	if st.ByLevel[1] != 1 || st.ByLevel[2] != 1 {
		t.Fatalf("expected one promoted into level 1 and one remaining in level 2, got %+v", st.ByLevel)
	}
}
