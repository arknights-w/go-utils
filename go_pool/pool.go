package go_pool

import (
	"sync"
	"time"
)

type Mode int8

const (
	MODE_HUNGRY Mode = iota
	MODE_BALANCE
)

type Pool struct {
	mu           *sync.Mutex
	cancel       chan struct{}
	workers      []*Worker[int64]
	maxWorkerNum int
	maxTaskNum   int
	now          int
	mode         Mode
}

func NewPool(workerNum, taskQueSize int) *Pool {
	if workerNum <= 0 {
		workerNum = DEFAULT_MAX_WORKER_NUM
	}
	if taskQueSize <= 0 {
		taskQueSize = DEFAULT_TASK_CHAN_SIZE
	}
	pool := &Pool{
		mu:           &sync.Mutex{},
		workers:      []*Worker[int64]{NewWorker(time.Now().UnixNano(), taskQueSize)},
		maxWorkerNum: workerNum,
		maxTaskNum:   taskQueSize,
		mode:         MODE_HUNGRY,
		cancel:       make(chan struct{}),
	}
	go pool.demon()
	return pool
}

func (s *Pool) AddTask(task Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() { s.now = (s.now + 1) % len(s.workers) }()

	switch s.mode {
	case MODE_HUNGRY:
		// 饥饿模式，优先新增worker
		for i := 0; i <= len(s.workers); i++ {
			idx := (s.now + i) % len(s.workers)
			if s.workers[idx].Count() <= 0 && s.workers[idx].TryAddTask(task) {
				s.now = idx
				return
			}
		}
	case MODE_BALANCE:
		// 平衡模式，直接轮询
		for i := 0; i <= len(s.workers); i++ {
			idx := (s.now + i) % len(s.workers)
			if s.workers[idx].TryAddTask(task) {
				s.now = idx
				return
			}
		}
	}
	s.schedule()
	s.workers[s.now].AddTask(task)
}

func (s *Pool) Group() *TaskGroup {
	return NewTaskGroup(s.AddTask)
}

func (s *Pool) Close() {
	close(s.cancel)
	for _, worker := range s.workers {
		worker.Close()
	}
	for _, worker := range s.workers {
		worker.Wait()
	}
}

// region schedule about

func (s *Pool) demon() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			s.schedule()
			s.mu.Unlock()
		case <-s.cancel:
			return
		}
	}
}

func (s *Pool) schedule() {
	infoV3 := s.infoCollect()
	az := s.infoAnalyze(infoV3)
	s.scheduleAction(az)
}

type info struct {
	minIdx       int     // 最小任务数 worker 下标
	minTaskNum   int     // 最小任务数
	maxIdx       int     // 最大任务数 worker 下标
	maxTaskNum   int     // 最大任务数
	totalTasks   int     // 总任务数
	runningCount int     // 正在运行的worker数量
	workerCount  int     // worker数量
	avgTaskNum   float64 // 平均任务数
}

func (s *Pool) infoCollect() *info {
	workerLen := len(s.workers)
	infoV3 := &info{
		minIdx:       0,
		minTaskNum:   s.workers[0].Count(),
		maxIdx:       0,
		maxTaskNum:   s.workers[0].Count(),
		totalTasks:   0,
		runningCount: 0,
		workerCount:  workerLen,
	}

	for i := range workerLen {
		count := s.workers[i].Count()
		infoV3.totalTasks += count
		if count > 0 {
			infoV3.runningCount++
		}

		if count < infoV3.minTaskNum {
			infoV3.minTaskNum = count
			infoV3.minIdx = i
		}

		if count > infoV3.maxTaskNum {
			infoV3.maxTaskNum = count
			infoV3.maxIdx = i
		}
	}

	infoV3.avgTaskNum = float64(infoV3.totalTasks) / float64(workerLen)
	return infoV3
}

type analysis struct {
	add       bool
	del       int
	stealFrom int
	stealTo   int
	minIdx    int
	mode      Mode
}

func (s *Pool) infoAnalyze(info *info) *analysis {
	res := &analysis{del: -1, stealFrom: -1, stealTo: -1, minIdx: info.minIdx}

	// 扩容策略：负载超过80%且未达上限
	if len(s.workers) < s.maxWorkerNum*8/10 || (info.avgTaskNum > float64(s.maxTaskNum)/3 && len(s.workers) < s.maxWorkerNum) {
		res.add = true
	}

	// 缩容策略：存在空闲Worker且数量大于1
	if info.minTaskNum == 0 && len(s.workers) > 1 {
		res.del = info.minIdx
	}

	// 负载均衡：最大负载超过平均1.5倍且存在低负载Worker
	if info.maxTaskNum > int(1.5*info.avgTaskNum) && info.minTaskNum < int(0.5*info.avgTaskNum) {
		res.stealFrom = info.maxIdx
		res.stealTo = info.minIdx
		return res
	}

	// 模式判断
	if info.totalTasks < s.maxWorkerNum {
		res.mode = MODE_HUNGRY
	} else {
		res.mode = MODE_BALANCE
	}

	return res
}

func (s *Pool) scheduleAction(az *analysis) {
	s.mode = az.mode
	s.now = az.minIdx
	if az.del != -1 {
		if az.del >= 0 && az.del < len(s.workers) {
			s.workers[az.del].Close()
			s.workers = append(s.workers[:az.del], s.workers[az.del+1:]...)
			if s.now >= az.del && s.now > 0 {
				s.now = (s.now - 1) % len(s.workers)
			}
		}
	}
	if az.stealFrom != -1 && az.stealTo != -1 {
		Steal(s.workers[az.stealFrom], s.workers[az.stealTo])
	}
	if az.add {
		s.addWorker()
		s.now = len(s.workers) - 1
	}
}

func (s *Pool) addWorker() {
	s.workers = append(
		s.workers,
		NewWorker(time.Now().UnixNano(), s.maxTaskNum),
	)
}

// endregion
