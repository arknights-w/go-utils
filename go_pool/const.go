package go_pool

const (
	DEFAULT_TASK_CHAN_SIZE = 0  // 默认任务通道大小
	DEFAULT_MAX_WORKER_NUM = 10 // 默认最大worker数量
)

type workerStatus uint32

const (
	WORKER_STATUS_RUNNING workerStatus = iota
	WORKER_STATUS_PENDING
	WORKER_STATUS_STOPPED
)
