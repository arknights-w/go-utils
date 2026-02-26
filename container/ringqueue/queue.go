package ringqueue

// Queue 是一个基于环形数组的 FIFO 队列。
//
// PushBack/PopFront 为摊还 O(1)；满时按 2 倍扩容并搬移一次数据以保持顺序。
type Queue[T any] struct {
	buf  []T
	head int
	size int
}

func (q *Queue[T]) Len() int { return q.size }

// Reset 清空队列。
func (q *Queue[T]) Reset() {
	var zero T
	for i := 0; i < q.size; i++ {
		q.buf[(q.head+i)%cap(q.buf)] = zero
	}
	q.head = 0
	q.size = 0
}

// PushBack 入队（队尾追加）。
func (q *Queue[T]) PushBack(v T) {
	if cap(q.buf) == 0 {
		q.buf = make([]T, 4)
	}
	if q.size == cap(q.buf) {
		q.grow()
	}
	tail := (q.head + q.size) % cap(q.buf)
	q.buf[tail] = v
	q.size++
}

// PeekFront 查看队头但不出队。
func (q *Queue[T]) PeekFront() (T, bool) {
	if q.size == 0 {
		var zero T
		return zero, false
	}
	return q.buf[q.head], true
}

// PopFront 出队（队头移除）。
func (q *Queue[T]) PopFront() (T, bool) {
	if q.size == 0 {
		var zero T
		return zero, false
	}
	v := q.buf[q.head]
	var zero T
	q.buf[q.head] = zero
	q.head++
	if q.head == cap(q.buf) {
		q.head = 0
	}
	q.size--
	if q.size == 0 {
		q.head = 0
	}
	return v, true
}

func (q *Queue[T]) grow() {
	nb := make([]T, cap(q.buf)*2)
	for i := range q.size {
		// for i := 0; i < q.size; i++ {
		nb[i] = q.buf[(q.head+i)%cap(q.buf)]
	}
	q.buf = nb
	q.head = 0
}
