package mlfq

import (
	"context"
	"sync"
	"time"
)

// Option 用于配置调度器（如时钟注入、自动 Tick 等）。
type Option func(*config)

type config struct {
	now func() time.Time

	autoTickInterval time.Duration
}

// WithClock 注入时钟函数，便于测试可控。
func WithClock(now func() time.Time) Option {
	return func(c *config) {
		c.now = now
	}
}

// WithAutoTick 启用后台自动 Tick；interval<=0 表示禁用。
//
// 开启后请在退出时调用 Close()，以停止后台 goroutine。
func WithAutoTick(interval time.Duration) Option {
	return func(c *config) {
		c.autoTickInterval = interval
	}
}

type scheduler[T any] struct {
	mu     sync.Mutex
	closed bool

	policy Policy[T]
	mq     *MultiQueue[*taskState[T]]
	states map[Token]*taskState[T]

	nextToken uint64

	cfg config

	autoTickCancel context.CancelFunc
	autoTickWG     sync.WaitGroup

	submitted     uint64
	dequeued      uint64
	finished      uint64
	requeued      uint64
	promoted      uint64
	demoted       uint64
	agingPromoted uint64
}

type taskState[T any] struct {
	token      Token
	task       T
	level      int
	attrs      Attributes
	enqueuedAt time.Time

	// leased 表示当前已被 Next 发放 Lease，等待 FeedBack。
	leased       bool
	lastDequeued time.Time
	lastQuantum  time.Duration
}

// New 创建一个 MLFQ 调度器实例，并对外以接口 MLFQ[T] 暴露（隐藏内部实现细节）。
func New[T any](policy Policy[T], opts ...Option) (MLFQ[T], error) {
	if policy == nil {
		return nil, ErrNilPolicy
	}

	cfg := config{
		now:              time.Now,
		autoTickInterval: 0,
	}
	for _, o := range opts {
		o(&cfg)
	}

	levels := policy.Levels()
	if levels <= 0 {
		return nil, ErrInvalidLevel
	}

	s := &scheduler[T]{
		policy: policy,
		mq:     NewMultiQueue[*taskState[T]](levels),
		states: make(map[Token]*taskState[T]),
		cfg:    cfg,
	}

	if cfg.autoTickInterval > 0 {
		s.startAutoTick(cfg.autoTickInterval)
	}

	return s, nil
}

// NewDefault 使用 DefaultPolicy 创建调度器。
func NewDefault[T any](levels int, opts ...Option) (MLFQ[T], error) {
	return New[T](NewDefaultPolicy[T](levels, DefaultPolicyConfig{}), opts...)
}

// Close 关闭调度器并停止后台 auto-tick（若启用）。
//
// Close 幂等：允许重复调用。
func (s *scheduler[T]) Close() error {
	s.mu.Lock()
	s.closed = true
	cancel := s.autoTickCancel
	s.autoTickCancel = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	s.autoTickWG.Wait()
	return nil
}

func (s *scheduler[T]) startAutoTick(interval time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	s.autoTickCancel = cancel

	ticker := time.NewTicker(interval)
	s.autoTickWG.Add(1)
	go func() {
		defer s.autoTickWG.Done()
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				// auto-tick 不依赖外部 ctx；Tick 内部会自行检查 closed。
				s.Tick(context.Background(), now)
			}
		}
	}()
}

func (s *scheduler[T]) Submit(ctx context.Context, task T, opts ...SubmitOption) (Token, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	var so SubmitOptions
	for _, o := range opts {
		o(&so)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return 0, ErrClosed
	}

	now := s.cfg.now()
	level := s.policy.OnSubmit(now, task, so)
	if level < 0 || level >= s.mq.Levels() {
		return 0, ErrInvalidLevel
	}

	s.nextToken++
	tok := Token(s.nextToken)

	st := &taskState[T]{
		token:      tok,
		task:       task,
		level:      level,
		attrs:      so.Attrs,
		enqueuedAt: now,
	}
	s.states[tok] = st
	s.mq.Push(level, st)
	s.submitted++
	return tok, nil
}

func (s *scheduler[T]) Next(ctx context.Context) (Lease[T], bool) {
	var zero Lease[T]
	if ctx.Err() != nil {
		return zero, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return zero, false
	}

	if s.mq.TotalLen() == 0 {
		return zero, false
	}

	now := s.cfg.now()
	level, ok := s.policy.PickNext(now, s.mq)
	if !ok {
		return zero, false
	}
	if level < 0 || level >= s.mq.Levels() {
		return zero, false
	}

	st, ok := s.mq.Pop(level)
	if !ok || st == nil {
		return zero, false
	}

	q := s.policy.Quantum(now, level, st.task)
	st.level = level
	st.leased = true
	st.lastDequeued = now
	st.lastQuantum = q
	s.dequeued++

	return Lease[T]{
		Token:      st.token,
		Task:       st.task,
		Level:      level,
		Quantum:    q,
		DequeuedAt: now,
	}, true
}

func (s *scheduler[T]) FeedBack(ctx context.Context, token Token, fb Feedback) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrClosed
	}

	st, ok := s.states[token]
	if !ok {
		return ErrUnknownToken
	}
	if !st.leased {
		return ErrNotLeased
	}

	now := s.cfg.now()

	if !fb.UsedFullQuantum && st.lastQuantum > 0 && fb.RanFor >= st.lastQuantum {
		fb.UsedFullQuantum = true
	}
	// 默认使用提交时的 attrs；若调用方希望动态调整，可在 Feedback.Attrs 中提供新值。
	if fb.Attrs == (Attributes{}) {
		fb.Attrs = st.attrs
	}

	st.leased = false

	if fb.Finished {
		delete(s.states, token)
		s.finished++
		return nil
	}

	oldLevel := st.level
	newLevel, requeue := s.policy.OnFeedback(now, oldLevel, st.task, fb)
	if !requeue {
		delete(s.states, token)
		s.finished++
		return nil
	}
	if newLevel < 0 || newLevel >= s.mq.Levels() {
		return ErrInvalidLevel
	}

	if newLevel < oldLevel {
		s.promoted++
	}
	if newLevel > oldLevel {
		s.demoted++
	}

	st.level = newLevel
	st.attrs = fb.Attrs
	st.enqueuedAt = now
	s.mq.Push(newLevel, st)
	s.requeued++
	return nil
}

func (s *scheduler[T]) Tick(ctx context.Context, now time.Time) {
	if ctx.Err() != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}

	levels := s.mq.Levels()
	for level := levels - 1; level >= 1; level-- {
		st, ok := s.mq.Peek(level)
		if !ok || st == nil {
			continue
		}
		promote, newLevel := s.policy.OnAging(now, level, st.task, st.enqueuedAt)
		if !promote {
			continue
		}
		if newLevel < 0 || newLevel >= levels {
			continue
		}
		if newLevel == level {
			continue
		}

		_, _ = s.mq.Pop(level)
		st.level = newLevel
		st.enqueuedAt = now
		s.mq.Push(newLevel, st)
		s.agingPromoted++
	}
}

func (s *scheduler[T]) Stats(ctx context.Context) Stats {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	by := make([]int, s.mq.Levels())
	for i := range by {
		by[i] = s.mq.Len(i)
	}

	return Stats{
		Now:      s.cfg.now(),
		Levels:   s.mq.Levels(),
		TotalLen: s.mq.TotalLen(),
		ByLevel:  by,

		Submitted:     s.submitted,
		Dequeued:      s.dequeued,
		Finished:      s.finished,
		Requeued:      s.requeued,
		Promoted:      s.promoted,
		Demoted:       s.demoted,
		AgingPromoted: s.agingPromoted,

		BitMapWords: s.mq.BitMapWords(),
	}
}
