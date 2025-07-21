package errs

var (
	// workflow build err form 10001 to 20000

	// 阶段名称重复
	ErrDupStage = &WorkflowErr{code: 10001, msg: "duplicate stage"}
	// 循环依赖
	ErrCircDep = &WorkflowErr{code: 10002, msg: "circular dependency"}
	// 阶段不存在
	ErrNoStage = &WorkflowErr{code: 10003, msg: "stage not found"}

	// workflow run err form 20001 to 30000

	// 未知的workflow类型
	ErrNoWorkflowType = &WorkflowErr{code: 20001, msg: "unknown workflow type"}

	// stage err form 30001 to 40000
)
