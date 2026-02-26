package multiqueue

import (
	"github.com/arknights-w/go-utils/container/bitmap"
	"github.com/arknights-w/go-utils/container/ringqueue"
)

// View 是 MultiQueue 的只读视图接口（供策略选择下一个 level）。
type View interface {
	Levels() int
	TotalLen() int
	Len(level int) int
	MinNonEmpty() (int, bool)
	MaxNonEmpty() (int, bool)
}

// MultiQueue 是多级队列：每个 level 一个 FIFO 队列，另配套位图用于快速定位非空 level。
type MultiQueue[T any] struct {
	levels int
	bm     bitmap.BitMap
	qs     []ringqueue.Queue[T]
	total  int
}

// New 创建一个 levels 层的多级队列。
func New[T any](levels int) *MultiQueue[T] {
	if levels <= 0 {
		panic("mlfq: levels must be > 0")
	}
	return &MultiQueue[T]{
		levels: levels,
		bm:     bitmap.New(levels),
		qs:     make([]ringqueue.Queue[T], levels),
	}
}

func (m *MultiQueue[T]) Levels() int { return m.levels }

func (m *MultiQueue[T]) TotalLen() int { return m.total }

func (m *MultiQueue[T]) Len(level int) int {
	m.mustLevel(level)
	return m.qs[level].Len()
}

func (m *MultiQueue[T]) MinNonEmpty() (int, bool) { return m.bm.Min() }

func (m *MultiQueue[T]) MaxNonEmpty() (int, bool) { return m.bm.Max() }

func (m *MultiQueue[T]) BitMapWords() []uint64 { return m.bm.Words() }

func (m *MultiQueue[T]) Push(level int, v T) {
	m.mustLevel(level)
	if m.qs[level].Len() == 0 {
		m.bm.Set(level)
	}
	m.qs[level].PushBack(v)
	m.total++
}

func (m *MultiQueue[T]) Peek(level int) (T, bool) {
	m.mustLevel(level)
	return m.qs[level].PeekFront()
}

func (m *MultiQueue[T]) Pop(level int) (T, bool) {
	m.mustLevel(level)
	v, ok := m.qs[level].PopFront()
	if !ok {
		var zero T
		return zero, false
	}
	m.total--
	if m.qs[level].Len() == 0 {
		m.bm.Clear(level)
	}
	return v, true
}

func (m *MultiQueue[T]) mustLevel(level int) {
	if level < 0 || level >= m.levels {
		panic("mlfq: level out of range")
	}
}
