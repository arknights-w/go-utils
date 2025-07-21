package pool

import (
	"context"
	"os"
	"path"
	"strings"

	pool "github.com/arknights-w/go-utils/go_pool"
	iface "github.com/arknights-w/go-utils/rely/workflow/01_def_iface"
	ictx "github.com/arknights-w/go-utils/rely/workflow/02_context"
	"github.com/go-echarts/go-echarts/v2/charts"
)

type workflow[name comparable] struct {
	scheduler *pool.Pool
	stages    [][]iface.Stage[name]
	printG    *charts.Graph
}

func (w *workflow[nt]) Work(ctx context.Context) (err error) {
	context := ictx.NewSyncContext(ctx)
	tg := w.scheduler.Group()
	for _, stages := range w.stages {
		for _, stage := range stages {
			tg.AddTask(func() {
				if err := stage.Run(context); err != nil {
					context.Set(ictx.KEY_ERROR, err)
				}
			})
		}
		tg.Wait()
		if err := context.Err(); err != nil {
			return err
		}
	}
	return
}

func (w *workflow[nt]) GetStage(name nt) iface.Stage[nt] {
	for idx := range w.stages {
		for jdx := range w.stages[idx] {
			if w.stages[idx][jdx].Name() == name {
				return w.stages[idx][jdx]
			}
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
	w.scheduler.Close()
	return nil
}

func NewWorkflow[nt comparable](
	stages [][]iface.Stage[nt],
	printer *charts.Graph,
	threadNum int,
	waitQueNum int,
) iface.Workflow[nt] {
	return &workflow[nt]{
		scheduler: pool.NewPool(threadNum, waitQueNum),
		stages:    stages,
		printG:    printer,
	}
}
