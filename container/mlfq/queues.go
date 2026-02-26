package mlfq

import (
	"github.com/arknights-w/go-utils/container/bitmap"
	"github.com/arknights-w/go-utils/container/mlfq/multiqueue"
)

// BitMap 是 bitmap.BitMap 的根包别名，便于外部只 import mlfq 也能复用位图。
type BitMap = bitmap.BitMap

// NewBitMap 创建一个支持 n 个 bit 的位图。
func NewBitMap(n int) BitMap { return bitmap.New(n) }

// QueueView 是 multiqueue.View 的根包别名，供策略接口使用。
type QueueView = multiqueue.View

// MultiQueue 是 multiqueue.MultiQueue 的根包别名。
type MultiQueue[T any] = multiqueue.MultiQueue[T]

// NewMultiQueue 创建一个 levels 层的多级队列。
func NewMultiQueue[T any](levels int) *MultiQueue[T] { return multiqueue.New[T](levels) }
