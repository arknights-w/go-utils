package serial

import (
	"context"
	"os"
	"path"
	"strings"

	iface "github.com/arknights-w/go-utils/rely/workflow/01_def_iface"
	ictx "github.com/arknights-w/go-utils/rely/workflow/02_context"
	"github.com/go-echarts/go-echarts/v2/charts"
)

type workflow[name comparable] struct {
	stages []iface.Stage[name]
	printG *charts.Graph
}

func (w *workflow[nt]) Work(ctx context.Context) (err error) {
	context := ictx.NewContext(ctx)
	for _, stage := range w.stages {
		if err = stage.Run(context); err != nil {
			return
		}
	}
	return
}

func (w *workflow[nt]) GetStage(name nt) iface.Stage[nt] {
	for idx := range w.stages {
		if w.stages[idx].Name() == name {
			return w.stages[idx]
		}
	}
	return nil
}

func (w *workflow[nt]) Print(_path string, name string) error {
	fp := path.Join(_path, name)
	if !strings.HasSuffix(fp, ".html") {
		fp = fp + ".html"
	}
	return os.WriteFile(
		fp,
		[]byte(w.printG.RenderContent()),
		0755,
	)
}

func (w *workflow[nt]) Close() error {
	return nil
}

func NewWorkflow[nt comparable](
	stages [][]iface.Stage[nt],
	printer *charts.Graph,
) iface.Workflow[nt] {
	length := 0
	for idx := range stages {
		length += len(stages[idx])
	}
	stageLi := make([]iface.Stage[nt], 0, length)
	for idx := range stages {
		stageLi = append(stageLi, stages[idx]...)
	}
	return &workflow[nt]{
		stages: stageLi,
		printG: printer,
	}
}
