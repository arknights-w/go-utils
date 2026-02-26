package mlfq

import "time"

// Stats 是调度器对外提供的状态快照（用于观测/调试）。
type Stats struct {
	// Now 是快照生成时间（由内部时钟函数产生）。
	Now time.Time
	// Levels 是总层级数。
	Levels int
	// TotalLen 是所有层级的总任务数（不包含已发放但未反馈的任务）。
	TotalLen int
	// ByLevel 是每个 level 的队列长度快照，下标对应 level。
	ByLevel []int

	// Submitted 是累计提交任务数。
	Submitted uint64
	// Dequeued 是累计 Next 成功取出任务数。
	Dequeued uint64
	// Finished 是累计完成并移除的任务数。
	Finished uint64
	// Requeued 是累计反馈后重新入队的次数。
	Requeued uint64
	// Promoted 是累计反馈导致的升级次数（newLevel < oldLevel）。
	Promoted uint64
	// Demoted 是累计反馈导致的降级次数（newLevel > oldLevel）。
	Demoted uint64
	// AgingPromoted 是累计 Tick 老化提升次数。
	AgingPromoted uint64

	// BitMapWords 是当前多级队列位图的 uint64 words 快照（用于调试）。
	BitMapWords []uint64
}
