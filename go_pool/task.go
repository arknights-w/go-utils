package go_pool

import "sync"

type Task func()

type TaskGroup struct {
	wg    sync.WaitGroup
	adder func(Task)
}

func (t *TaskGroup) AddTask(task Task) {
	t.wg.Add(1)
	t.adder(func() {
		defer t.wg.Done()
		task()
	})
}

func (t *TaskGroup) Wait() {
	t.wg.Wait()
}

func NewTaskGroup(adder func(Task)) *TaskGroup {
	return &TaskGroup{adder: adder}
}
