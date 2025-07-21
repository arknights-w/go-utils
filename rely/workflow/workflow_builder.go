package workflow

import (
	"fmt"

	errs "github.com/arknights-w/go-utils/rely/workflow/01_def_errs"
	iface "github.com/arknights-w/go-utils/rely/workflow/01_def_iface"
	parallel "github.com/arknights-w/go-utils/rely/workflow/03_workflow_parallel"
	pool "github.com/arknights-w/go-utils/rely/workflow/03_workflow_pool"
	serial "github.com/arknights-w/go-utils/rely/workflow/03_workflow_serial"
	topo "github.com/arknights-w/go-utils/topology"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

type WorkflowBuilder[name comparable] struct {
	stages map[name]iface.Stage[name]
}

func NewBuilder[name comparable](stages ...iface.Stage[name]) (*WorkflowBuilder[name], error) {
	// 构建工作流
	builder := &WorkflowBuilder[name]{
		stages: make(map[name]iface.Stage[name]),
	}
	// stage add
	for _, stage := range stages {
		if _, ok := builder.stages[stage.Name()]; ok {
			return nil, errs.ErrDupStage.WithDesc(fmt.Sprintf("duplicate stage: %v", stage.Name()))
		}
		builder.stages[stage.Name()] = stage
	}
	return builder, nil
}

func (builder *WorkflowBuilder[name]) AddStage(stage iface.Stage[name]) error {
	if _, ok := builder.stages[stage.Name()]; ok {
		return errs.ErrDupStage.WithDesc(fmt.Sprintf("duplicate stage: %v", stage.Name()))
	}
	builder.stages[stage.Name()] = stage
	return nil
}

func (builder *WorkflowBuilder[name]) Build(opts ...buildOpt) (iface.Workflow[name], error) {
	// 1. 构建 printer
	prtGraph := builder.buildPrinter()

	// 2. 构建出度边表 outDegreeEdges
	outDegEdge := make(map[name][]name)
	for _, stage := range builder.stages {
		if _, ok := outDegEdge[stage.Name()]; !ok {
			outDegEdge[stage.Name()] = nil
		}
		for _, dep := range stage.DependOn() {
			// 检查依赖的阶段是否存在
			if _, ok := builder.stages[dep]; !ok {
				return nil, errs.ErrNoStage.WithDesc(fmt.Sprintf("no stage: %v", dep))
			}
			outDegEdge[dep] = append(outDegEdge[dep], stage.Name())
		}
	}

	// 3. 拓扑排序
	sorted, cycle := topo.TopologicalSort(outDegEdge)

	// 4. 循环检测
	if len(cycle) != 0 {
		return serial.NewWorkflow[name](nil, prtGraph), errs.ErrCircDep.WithDesc(fmt.Sprintf("cycle: %v", cycle))
	}

	// 5. 构建工作流
	sortedStages := make([][]iface.Stage[name], len(sorted))
	for idx, pieceSorted := range sorted {
		sortedStages[idx] = make([]iface.Stage[name], 0, len(pieceSorted))
		for _, name := range pieceSorted {
			sortedStages[idx] = append(sortedStages[idx], builder.stages[name])
		}
	}

	// 6. 构建 workflow
	opt := defaultBuildOption()
	for _, optFn := range opts {
		optFn(opt)
	}
	switch opt.runType {
	case runSerial:
		return serial.NewWorkflow(sortedStages, prtGraph), nil
	case runParallel:
		return parallel.NewWorkflow(sortedStages, prtGraph), nil
	case runPool:
		return pool.NewWorkflow(sortedStages, prtGraph, opt.threadNum, opt.waitQueNum), nil
	default:
		return nil, errs.ErrNoWorkflowType.WithDesc(fmt.Sprintf("no workflow type: %d", opt.runType))
	}
}

func (builder *WorkflowBuilder[name]) buildPrinter() *charts.Graph {
	var (
		prtEdge   = make([]opts.GraphLink, 0, len(builder.stages)+1)
		prtNode   = make([]opts.GraphNode, 0, len(builder.stages)+1)
		graph     = charts.NewGraph()
		degreeMap = make(map[string]int)
	)
	// 1. build node and edge
	for _, stage := range builder.stages {
		stageName := fmt.Sprint(stage.Name())
		degreeMap[stageName] = len(stage.DependOn())
		prtNode = append(prtNode, opts.GraphNode{
			Name:       stageName,
			SymbolSize: 50,
			Tooltip:    &opts.Tooltip{Formatter: types.FuncStr(stage.Desc())},
		})
		for _, dep := range stage.DependOn() {
			prtEdge = append(prtEdge, opts.GraphLink{
				Source: fmt.Sprint(dep),
				Target: stageName,
			})
		}
	}
	// 2. build start node
	prtNode = append(prtNode, opts.GraphNode{
		Name:       "Σ graph start",
		Tooltip:    &opts.Tooltip{Formatter: "avoid same name"},
		SymbolSize: 50,
		Fixed:      opts.Bool(true),
		X:          200,
		Y:          200,
	})
	for name, degree := range degreeMap {
		if degree == 0 {
			prtEdge = append(prtEdge, opts.GraphLink{
				Source: "Σ graph start",
				Target: name,
			})
		}
	}
	// 3. build graph
	graph.AddSeries("", prtNode, prtEdge,
		charts.WithGraphChartOpts(opts.GraphChart{
			EdgeSymbol: []string{"circle", "arrow"},
			Force:      &opts.GraphForce{Repulsion: 1000, EdgeLength: 100},
			Draggable:  opts.Bool(true),
		}),
		charts.WithLabelOpts(opts.Label{
			Show: opts.Bool(true),
		}),
	).SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:     "90vw",
			Height:    "90vh",
			PageTitle: "workflow dependency graph",
		}),
	)
	return graph
}
