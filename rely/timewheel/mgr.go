package timewheel

import (
	"container/heap"
	"sync"
)

type taskMgr struct {
	id int64
	li taskLi
	mu sync.Mutex
}

func (mgr *taskMgr) genId() int64 {
	mgr.id++
	return mgr.id
}

func (mgr *taskMgr) AddTask(execTime int64, fn func()) int64 {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	heap.Push(&mgr.li, task{
		execTime: execTime,
		id:       mgr.genId(),
		fn:       fn,
	})
	return mgr.id
}

func (mgr *taskMgr) RemoveTask(id int64) {
	if mgr.li.Len() == 0 {
		return
	}
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	for idx, task := range mgr.li {
		if task.id == id {
			heap.Remove(&mgr.li, idx)
			return
		}
	}
}

func (mgr *taskMgr) GetRunableTasks(time int64) []task {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	var tasks []task
	for mgr.li.Len() > 0 && mgr.li.Peek().execTime <= time {
		tasks = append(tasks, heap.Pop(&mgr.li).(task))
	}
	return tasks
}

type taskLi []task

func (li taskLi) Len() int           { return len(li) }
func (li taskLi) Less(i, j int) bool { return li[i].execTime < li[j].execTime }
func (li taskLi) Swap(i, j int)      { li[i], li[j] = li[j], li[i] }
func (li taskLi) Peek() task         { return li[0] }
func (li *taskLi) Push(x any)        { *li = append(*li, x.(task)) }
func (li *taskLi) Pop() any {
	old := *li
	n := len(old)
	x := old[n-1]
	*li = old[0 : n-1]
	return x
}

type task struct {
	execTime int64
	id       int64
	fn       func()
}
