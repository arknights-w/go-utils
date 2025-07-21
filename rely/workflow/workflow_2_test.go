package workflow_test

import (
	"context"
	"fmt"
	"testing"

	wf "github.com/arknights-w/go-utils/rely/workflow"
	iface "github.com/arknights-w/go-utils/rely/workflow/01_def_iface"
)

func init() {
	builder, err := wf.NewBuilder(
		wf.NewStage("prepare", []string{}, "", runPrepare),
		wf.NewStage("diversion", []string{"prepare"}, "", runDiversion),
		wf.NewStage("backups", []string{"diversion"}, "", runBackups),
		wf.NewStage("custom", []string{"diversion", "backups"}, "", runCustom),
		wf.NewStage("bon2json", []string{"custom"}, "", runBon2Json),
		wf.NewStage("normalize", []string{"bon2json"}, "", runNormalize),
		wf.NewStage("preprocess", []string{"bon2json"}, "", runPreprocess),
		wf.NewStage("retry", []string{"preprocess", "normalize"}, "", runRetry),
		wf.NewStage("state", []string{"retry"}, "", runState),
		wf.NewStage("insert", []string{"state"}, "", runInsert),
	)
	if err != nil {
		panic(fmt.Sprintf("NewBuilder err: %v\n", err))
	}
	workflow, err = builder.Build()
	if err != nil {
		panic(fmt.Sprintf("Build err: %v\n", err))
	}
}

var workflow iface.Workflow[string]

func TestCreateProcess(t *testing.T) {
	workflow.Work(context.Background())
	workflow.Print("", "test_create_process.html")
}

// region 准备工作
func runPrepare(ctx iface.Context) error {
	println("[准备工作]biz_id校验... pass")
	println("[准备工作]初始化strMsgTraceId... pass")
	println("[准备工作]数据标识位处理... pass")
	return nil
}

// endregion

// region 批次分流
func runDiversion(ctx iface.Context) error {
	println("[批次分流]批次分流... pass")
	return nil
}

// endregion

// region 关键路径备份
func runBackups(ctx iface.Context) error {
	println("[backups]关键路径备份... pass")
	println("[backups]创建gid, trace_id... pass")
	return nil
}

// endregion

// region 个性化
func runCustom(ctx iface.Context) error {
	println("[个性化]入库前个性化... pass")
	return nil
}

// endregion

// region bon2json
func runBon2Json(ctx iface.Context) error {
	println("[bon2json]bon转json... pass")
	return nil
}

// endregion

// region 排序归一化
func runNormalize(ctx iface.Context) error {
	println("[normalize]排序归一化... pass")
	return nil
}

// endregion

// region 数据预处理
func runPreprocess(ctx iface.Context) error {
	println("[preprocess]数据预处理... pass")
	return nil
}

// endregion

// region 服务重试
func runRetry(ctx iface.Context) error {
	println("[retry]服务重试... pass")
	return nil
}

// endregion

// region 状态处理
func runState(ctx iface.Context) error {
	println("[state]状态处理... pass")
	return nil
}

// endregion

// region 数据入库
func runInsert(ctx iface.Context) error {
	println("[insert]穿透写db... pass")
	println("[insert]入库... pass")
	println("[insert]上报... pass")
	return nil
}

// endregion
