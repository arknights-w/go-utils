package workflow

type runType int

const (
	runSerial runType = iota
	runParallel
	runPool
)

type buildOption struct {
	runType
	threadNum  int
	waitQueNum int
}

type buildOpt func(*buildOption)

func defaultBuildOption() *buildOption {
	return &buildOption{
		runType:    runSerial,
		threadNum:  10,
		waitQueNum: 10,
	}
}

func WithRunSerial() buildOpt {
	return func(bo *buildOption) {
		bo.runType = runSerial
	}
}

func WithRunParallel() buildOpt {
	return func(bo *buildOption) {
		bo.runType = runParallel
	}
}

func WithRunPool(threadNum, waitQueNum int) buildOpt {
	return func(bo *buildOption) {
		bo.runType = runPool
		bo.threadNum = threadNum
		bo.waitQueNum = waitQueNum
	}
}
