package mlfq

import "errors"

var (
	// ErrInvalidLevel 表示策略或调用方提供的 level 超出 [0, Levels)。
	ErrInvalidLevel = errors.New("mlfq: invalid level")
	// ErrUnknownToken 表示 FeedBack 的 token 不存在（已完成/已被移除/从未提交）。
	ErrUnknownToken = errors.New("mlfq: unknown token")
	// ErrNotLeased 表示 token 对应任务当前不处于“已发放 Lease、等待反馈”的状态。
	ErrNotLeased = errors.New("mlfq: token not leased")
	// ErrNilPolicy 表示 New 传入了 nil policy。
	ErrNilPolicy = errors.New("mlfq: nil policy")
	// ErrClosed 表示调度器已 Close。
	ErrClosed = errors.New("mlfq: closed")
)
