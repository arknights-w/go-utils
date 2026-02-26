package mlfq

import "time"

// Policy 定义 MLFQ 的可插拔策略：决定初始 level、如何选择 Next、时间片长度、反馈升降级、以及老化提升。
type Policy[T any] interface {
	Levels() int

	OnSubmit(now time.Time, task T, opts SubmitOptions) int
	PickNext(now time.Time, q QueueView) (level int, ok bool)
	Quantum(now time.Time, level int, task T) time.Duration

	OnFeedback(now time.Time, level int, task T, fb Feedback) (newLevel int, requeue bool)
	OnAging(now time.Time, level int, task T, enqueuedAt time.Time) (promote bool, newLevel int)
}

// DefaultPolicyConfig 是默认策略的可调参数。
type DefaultPolicyConfig struct {
	// BaseQuantum 是 level=0 的基础时间片。
	BaseQuantum time.Duration
	// MaxQuantum 是时间片封顶上限（Quantum 永远不会超过该值）。
	MaxQuantum time.Duration
	// AgingThreshold 是老化提升阈值：任务在某 level 等待超过该时长后可被提升（由 OnAging 决定）。
	AgingThreshold time.Duration
}

// DefaultPolicy 是一个可用的基线策略：
//   - 0 为最高优先级，level 越大优先级越低
//   - 时间片随 level 指数增长并以 MaxQuantum 封顶
//   - 用满时间片则倾向降级；高紧急/重要任务可适度升级
type DefaultPolicy[T any] struct {
	levels int
	cfg    DefaultPolicyConfig
}

func NewDefaultPolicy[T any](levels int, cfg DefaultPolicyConfig) *DefaultPolicy[T] {
	if cfg.BaseQuantum <= 0 {
		cfg.BaseQuantum = 10 * time.Millisecond
	}
	if cfg.MaxQuantum <= 0 {
		cfg.MaxQuantum = 1 * time.Second
	}
	if cfg.AgingThreshold <= 0 {
		cfg.AgingThreshold = 1 * time.Second
	}
	return &DefaultPolicy[T]{levels: levels, cfg: cfg}
}

func (p *DefaultPolicy[T]) Levels() int { return p.levels }

func (p *DefaultPolicy[T]) OnSubmit(_ time.Time, _ T, opts SubmitOptions) int {
	score := min(max(int(opts.Attrs.Urgency)+int(opts.Attrs.Importance), 0), 200)
	rank := score * p.levels / 201 // 0..levels-1
	level := max((p.levels-1)-rank, 0)
	if level >= p.levels {
		level = p.levels - 1
	}
	return level
}

func (p *DefaultPolicy[T]) PickNext(_ time.Time, q QueueView) (int, bool) {
	return q.MinNonEmpty()
}

// Quantum 计算某个 level 的建议时间片：BaseQuantum * 2^level，并以 MaxQuantum 封顶。
func (p *DefaultPolicy[T]) Quantum(_ time.Time, level int, _ T) time.Duration {
	if level < 0 {
		level = 0
	}
	q := p.cfg.BaseQuantum
	for i := 0; i < level; i++ {
		if q >= p.cfg.MaxQuantum {
			return p.cfg.MaxQuantum
		}
		q *= 2
	}
	if q > p.cfg.MaxQuantum {
		q = p.cfg.MaxQuantum
	}
	return q
}

func (p *DefaultPolicy[T]) OnFeedback(_ time.Time, level int, _ T, fb Feedback) (int, bool) {
	if fb.Finished {
		return level, false
	}
	newLevel := level
	if fb.UsedFullQuantum {
		if newLevel < p.levels-1 {
			newLevel++
		}
		return newLevel, true
	}
	score := int(fb.Attrs.Urgency) + int(fb.Attrs.Importance)
	if score >= 150 && newLevel > 0 {
		newLevel--
	}
	return newLevel, true
}

func (p *DefaultPolicy[T]) OnAging(now time.Time, level int, _ T, enqueuedAt time.Time) (bool, int) {
	if level <= 0 {
		return false, level
	}
	if now.Sub(enqueuedAt) >= p.cfg.AgingThreshold {
		return true, level - 1
	}
	return false, level
}
