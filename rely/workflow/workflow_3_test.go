package workflow_test

import (
	"context"
	"testing"

	wf "github.com/arknights-w/go-utils/rely/workflow"
	iface "github.com/arknights-w/go-utils/rely/workflow/01_def_iface"
)

func TestCustomStage(t *testing.T) {
	builder, err := wf.NewBuilder(
		&StagePrepare{},
		&StageDiversion{},
		&StageBackups{},
		&StageCustom{},
		&StageBon2Json{},
		&StageNormalize{},
		&StagePreprocess{},
		&StageRetry{},
		&StageState{},
		&StageInsert{},
	)
	if err != nil {
		t.Fatalf("NewBuilder err: %v\n", err)
	}
	workflow, err := builder.Build()
	if err != nil {
		t.Fatalf("Build err: %v\n", err)
	}
	workflow.Work(context.Background())
	workflow.Print("", "test_custom_stage.html")
}

// region 准备工作

type StagePrepare struct{}

func (s *StagePrepare) Name() string {
	return "prepare"
}
func (s *StagePrepare) Desc() string {
	return "准备工作"
}
func (s *StagePrepare) Run(ctx iface.Context) error {
	println("[准备工作]biz_id校验... pass")
	println("[准备工作]初始化strMsgTraceId... pass")
	println("[准备工作]数据标识位处理... pass")
	return nil
}
func (s *StagePrepare) DependOn() []string {
	return nil
}

// endregion

// region 批次分流

type StageDiversion struct{}

func (s *StageDiversion) Name() string {
	return "diversion"
}
func (s *StageDiversion) Desc() string {
	return "批次分流"
}
func (s *StageDiversion) Run(ctx iface.Context) error {
	println("[批次分流]批次分流... pass")
	return nil
}
func (s *StageDiversion) DependOn() []string {
	return []string{"prepare"}
}

// endregion

// region 关键路径备份

type StageBackups struct{}

func (s *StageBackups) Name() string {
	return "backups"
}
func (s *StageBackups) Desc() string {
	return "关键路径备份，并生成TraceID"
}
func (s *StageBackups) Run(ctx iface.Context) error {
	println("[backups]关键路径备份... pass")
	println("[backups]创建gid, trace_id... pass")
	return nil
}
func (s *StageBackups) DependOn() []string {
	return []string{"diversion"}
}

// endregion

// region 个性化

type StageCustom struct{}

func (s *StageCustom) Name() string {
	return "custom"
}

func (s *StageCustom) Desc() string {
	return "入库前个性化"
}

func (s *StageCustom) Run(ctx iface.Context) error {
	println("[个性化]入库前个性化... pass")
	return nil
}

func (s *StageCustom) DependOn() []string {
	return []string{"diversion", "backups"}
}

// endregion

// region bon2json

type StageBon2Json struct{}

func (s *StageBon2Json) Name() string {
	return "bon2json"
}
func (s *StageBon2Json) Desc() string {
	return "bon2json"
}
func (s *StageBon2Json) Run(ctx iface.Context) error {
	println("[bon2json]bon转json... pass")
	return nil
}
func (s *StageBon2Json) DependOn() []string {
	return []string{"custom"}
}

// endregion

// region 排序归一化
type StageNormalize struct{}

func (s *StageNormalize) Name() string {
	return "normalize"
}
func (s *StageNormalize) Desc() string {
	return "排序归一化"
}
func (s *StageNormalize) Run(ctx iface.Context) error {
	println("[normalize]排序归一化... pass")
	return nil
}
func (s *StageNormalize) DependOn() []string {
	return []string{"bon2json"}
}

// endregion

// region 数据预处理

type StagePreprocess struct{}

func (s *StagePreprocess) Name() string {
	return "preprocess"
}
func (s *StagePreprocess) Desc() string {
	return "数据预处理"
}
func (s *StagePreprocess) Run(ctx iface.Context) error {
	println("[preprocess]数据预处理... pass")
	return nil
}
func (s *StagePreprocess) DependOn() []string {
	return []string{"bon2json"}
}

// endregion

// region 服务重试

type StageRetry struct{}

func (s *StageRetry) Name() string {
	return "retry"
}
func (s *StageRetry) Desc() string {
	return "服务重试"
}
func (s *StageRetry) Run(ctx iface.Context) error {
	println("[retry]服务重试... pass")
	return nil
}
func (s *StageRetry) DependOn() []string {
	return []string{"preprocess", "normalize"}
}

// endregion

// region 状态处理

type StageState struct{}

func (s *StageState) Name() string {
	return "state"
}
func (s *StageState) Desc() string {
	return "状态处理"
}
func (s *StageState) Run(ctx iface.Context) error {
	println("[state]状态处理... pass")
	return nil
}
func (s *StageState) DependOn() []string {
	return []string{"retry"}
}

// endregion

// region 数据入库

type StageInsert struct{}

func (s *StageInsert) Name() string {
	return "insert"
}
func (s *StageInsert) Desc() string {
	return "数据入库"
}
func (s *StageInsert) Run(ctx iface.Context) error {
	println("[insert]穿透写db... pass")
	println("[insert]入库... pass")
	println("[insert]上报... pass")
	return nil
}
func (s *StageInsert) DependOn() []string {
	return []string{"state"}
}

// endregion
