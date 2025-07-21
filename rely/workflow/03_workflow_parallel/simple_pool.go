package parallel

import (
	"sync"

	pool "github.com/arknights-w/go-utils/go_pool"
)

type Pool struct {
	mu      *sync.Mutex
	workers []*pool.Worker[int]
	now     int
}

func NewPool() *Pool {
	return &Pool{
		mu:      &sync.Mutex{},
		workers: nil,
	}
}

func (s *Pool) AddTask(task pool.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	loopNum := len(s.workers)
	cont := 0
	for cont < loopNum {
		s.now = (s.now + 1) % loopNum
		if s.workers[s.now].Count() != 0 {
			cont++
			continue
		}
		// 尝试添加任务，如果添加成功则返回
		if s.workers[s.now].TryAddTask(task) {
			return
		}
		break
	}
	// 如果所有的worker都满了，则添加一个worker
	s.addWorker()
	s.now = loopNum
	s.workers[s.now].AddTask(task)
}

func (s *Pool) addWorker() {
	s.workers = append(
		s.workers,
		pool.NewWorker(len(s.workers)),
	)
}

func (s *Pool) Group() *pool.TaskGroup {
	return pool.NewTaskGroup(s.AddTask)
}

func (s *Pool) Close() {
	for _, worker := range s.workers {
		worker.Close()
	}
	for _, worker := range s.workers {
		worker.Wait()
	}
}
