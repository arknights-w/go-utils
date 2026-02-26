package mlfq

import (
	"context"
	"time"
)

type Token uint64

// Lease 表示一次从调度器获取的“可执行任务租约”。
//
// 调用方应当：
//  1. 依据 Quantum 控制单次执行预算（例如运行到超时就暂停/切片）
//  2. 用 Token 调用 FeedBack 回传本次真实执行情况
type Lease[T any] struct {
	// Token 用于关联本次 Next 发放的任务与后续 FeedBack。
	Token Token
	// Task 是调用方提交的任务对象。
	Task T
	// Level 是本次出队时所在的队列层级（0 为最高优先级）。
	Level int
	// Quantum 是策略给出的本次建议时间片预算（上限），用于调用方控制单次执行切片长度。
	Quantum time.Duration
	// DequeuedAt 是本次 Next 取出任务的时间戳（由内部时钟函数产生）。
	DequeuedAt time.Time
}

// Attributes 用于描述任务属性（紧急程度、重要程度等），供策略计算初始/反馈后的优先级参考。
type Attributes struct {
	// Urgency 表示紧急程度，建议范围 0..100（越大越紧急）。
	Urgency int8
	// Importance 表示重要程度，建议范围 0..100（越大越重要）。
	Importance int8
}

// Feedback 由调用方在一次执行切片后回传。
//
// RanFor 是真实执行耗时；调度器会将其与上次 Next 返回的 Lease.Quantum 进行对比，用于推断是否“用满时间片”。
type Feedback struct {
	// RanFor 是本次执行切片的真实耗时。
	RanFor time.Duration
	// Finished 表示任务已完成；为 true 时调度器会移除该任务，不再入队。
	Finished bool
	// UsedFullQuantum 表示本次是否用满时间片。
	// 若调用方不填写（false），调度器会使用 RanFor >= Lease.Quantum 的规则进行推断。
	UsedFullQuantum bool
	// Attrs 允许调用方在反馈时更新任务属性（例如紧急/重要程度的动态变化）。
	// 若为零值，调度器默认沿用 Submit 时的属性。
	Attrs Attributes
}

type SubmitOptions struct {
	// Attrs 是任务的初始属性（供策略计算初始 level）。
	Attrs Attributes
}

type SubmitOption func(*SubmitOptions)

func WithAttributes(attrs Attributes) SubmitOption {
	return func(o *SubmitOptions) {
		o.Attrs = attrs
	}
}

// MLFQ 是对外调度接口（线程安全）。
//
// 约定：
//   - Submit 入队返回 Token
//   - Next 出队返回 Lease（含 Token）
//   - FeedBack 必须在 Next 后调用，且每个 Token 在被再次发放前只能反馈一次
//   - Tick 用于老化提升/维护（可手动调用，也可 WithAutoTick 自动调用）
//   - Close 用于停止后台 auto-tick（若启用）并将调度器标记为关闭
type MLFQ[T any] interface {
	Submit(ctx context.Context, task T, opts ...SubmitOption) (Token, error)
	Next(ctx context.Context) (Lease[T], bool)
	FeedBack(ctx context.Context, token Token, fb Feedback) error
	Tick(ctx context.Context, now time.Time)
	Stats(ctx context.Context) Stats
	Close() error
}
