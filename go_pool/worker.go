package go_pool

type Worker[id comparable] struct {
	id     id
	recv   chan Task
	cancel chan struct{}
	status workerStatus
}

func NewWorker[id comparable](_id id, chanSize ...int) *Worker[id] {
	var worker = &Worker[id]{id: _id, cancel: make(chan struct{})}
	if len(chanSize) > 0 {
		worker.recv = make(chan Task, chanSize[0])
	} else {
		worker.recv = make(chan Task, DEFAULT_TASK_CHAN_SIZE)
	}
	go worker.demon()
	return worker
}

func (w *Worker[id]) demon() {
	w.status = WORKER_STATUS_PENDING
	for task := range w.recv {
		if task == nil {
			continue
		}
		w.safeRun(task)
	}
	close(w.cancel)
	w.status = WORKER_STATUS_STOPPED
}

func (w *Worker[id]) safeRun(task Task) {
	defer func() {
		w.status = WORKER_STATUS_PENDING
		if r := recover(); r != nil {
			// Handle panic
		}
	}()
	w.status = WORKER_STATUS_RUNNING
	task()
}

func (w *Worker[id]) AddTask(task Task) {
	defer func() {
		if r := recover(); r != nil {
			// Handle panic
		}
	}()
	w.recv <- task // panic if channel is closed
}

func (w *Worker[id]) TryAddTask(task Task) bool {
	select {
	case w.recv <- task:
		return true
	default:
		return false
	}
}

func (w *Worker[id]) TryPopTask() (Task, bool) {
	select {
	case task, ok := <-w.recv:
		return task, ok
	default:
		return nil, false
	}
}

func (w *Worker[id]) Count() int {
	len := len(w.recv)
	if w.status == WORKER_STATUS_RUNNING {
		len += 1
	}
	return len
}

func (w *Worker[id]) Id() id {
	return w.id
}

func (w *Worker[id]) Close() {
	close(w.recv)
}

func (w *Worker[id]) Wait() {
	<-w.cancel
}

func (w *Worker[id]) Cap() int {
	return cap(w.recv)
}

func Steal[id comparable](from, to *Worker[id]) {
	var (
		task Task
		ok   bool
	)
	// 从 from 中取出一个任务放到 to 中
	if task, ok = from.TryPopTask(); !ok {
		return
	}
	origins := make([]Task, 0, to.Count())
	for {
		if task, ok := to.TryPopTask(); ok {
			origins = append(origins, task)
			continue
		}
		break
	}
	to.AddTask(task)
	for _, task = range origins {
		to.AddTask(task)
	}
}

func StealMany[id comparable](from, to *Worker[id], num int) {
	tasks := make([]Task, 0, num)
	// 从 from 中取出一个任务放到 to 中
	for i := 0; i < num; i++ {
		if task, ok := from.TryPopTask(); !ok {
			tasks = append(tasks, task)
			continue
		}
		break
	}
	origins := make([]Task, 0, to.Count())
	for {
		if task, ok := to.TryPopTask(); ok {
			origins = append(origins, task)
			continue
		}
		break
	}
	for _, task := range tasks {
		to.AddTask(task)
	}
	for _, task := range origins {
		to.AddTask(task)
	}
}
