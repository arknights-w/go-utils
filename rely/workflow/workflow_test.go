package workflow_test

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	wf "github.com/arknights-w/go-utils/rely/workflow"
	iface "github.com/arknights-w/go-utils/rely/workflow/01_def_iface"
)

type WorkType string

const (
	Init   WorkType = "init"
	Create WorkType = "create"
	Update WorkType = "update"
	Delete WorkType = "delete"
	None   WorkType = "none"
)

func TestSuccess(t *testing.T) {
	stages := []iface.Stage[WorkType]{
		wf.NewStage(Init,
			[]WorkType{}, "",
			func(ctx iface.Context) error {
				println("this is Init")
				ctx.Set(Init, "success")
				return nil
			},
		), wf.NewStage(Create, []WorkType{Init}, "",
			func(ctx iface.Context) error {
				fmt.Printf("this is Create, Init stage is %v, Update stage still is %v\n", ctx.Value(Init), ctx.Value(Update))
				ctx.Set(Create, "success")
				return nil
			},
		), wf.NewStage(Update, []WorkType{Create}, "",
			func(ctx iface.Context) error {
				fmt.Printf("this is Update, Create stage is %v\n", ctx.Value(Create))
				ctx.Child().Set(Update, "success")
				return nil
			},
		), wf.NewStage(Delete, []WorkType{Create, Update}, "",
			func(ctx iface.Context) error {
				fmt.Printf("this is Delete, can not get Update stage: %v\n", ctx.Value(Update))
				return nil
			},
		),
	}
	builder, err := wf.NewBuilder(stages...)
	if err != nil {
		t.Fatalf("NewBuilder err: %v\n", err)
	}
	workflow, err := builder.Build()
	if err != nil {
		t.Fatalf("Build err: %v\n", err)
	}
	workflow.Work(context.Background())
}

func TestDuplicate(t *testing.T) {
	var (
		init = wf.NewStage(
			Init, nil, "",
			func(ctx iface.Context) error {
				println("this is Init")
				return nil
			},
		)
		init2 = wf.NewStage(
			Init, nil, "",
			func(ctx iface.Context) error {
				println("this is Init2")
				return nil
			},
		)
	)
	builder, err := wf.NewBuilder(init, init2)
	if err != nil {
		t.Fatalf("NewBuilder err: %v\n", err)
	}
	workflow, err := builder.Build()
	if err != nil {
		t.Fatalf("Build err: %v\n", err)
	}
	workflow.Work(context.Background())
}

func TestNoStage(t *testing.T) {
	var (
		init = wf.NewStage(
			Init, nil, "",
			func(ctx iface.Context) error {
				println("this is Init")
				return nil
			},
		)
		init2 = wf.NewStage(
			Create, []WorkType{Init, None}, "",
			func(ctx iface.Context) error {
				println("this is Init2")
				return nil
			},
		)
	)
	builder, err := wf.NewBuilder(init, init2)
	if err != nil {
		t.Fatalf("NewBuilder err: %v\n", err)
	}
	workflow, err := builder.Build()
	if err != nil {
		t.Fatalf("Build err: %v\n", err)
	}
	workflow.Work(context.Background())
}

func TestCircular(t *testing.T) {
	var (
		init = wf.NewStage(Init, nil, "", func(ctx iface.Context) error {
			println("this is Init")
			return nil
		})
		create = wf.NewStage(
			Create, []WorkType{Init}, "",
			func(ctx iface.Context) error {
				println("this is Create")
				return nil
			},
		)
		update = wf.NewStage(
			Update, []WorkType{Create, Delete}, "",
			func(ctx iface.Context) error {
				println("this is Update")
				return nil
			},
		)
		delete = wf.NewStage(
			Delete, []WorkType{Update}, "",
			func(ctx iface.Context) error {
				println("this is Delete")
				return nil
			},
		)
	)
	builder, err := wf.NewBuilder(init, create, delete, update)
	if err != nil {
		t.Fatalf("NewBuilder err: %v\n", err)
	}
	workflow, err := builder.Build()
	if err != nil {
		t.Fatalf("Build err: %v\n", err)
	}
	workflow.Work(context.Background())
}

func TestPrint(t *testing.T) {
	stages := []iface.Stage[WorkType]{}
	for i := 1; i < 10; i++ {
		str_i := strconv.Itoa(i)
		stages = append(stages, wf.NewStage(
			WorkType("init_"+str_i),
			nil,
			"初始化 "+str_i,
			func(ctx iface.Context) error {
				println("this is Init", str_i)
				return nil
			},
		))
	}
	for i := 1; i < 10; i++ {
		str_i := strconv.Itoa(i)
		str_sub_i := strconv.Itoa(i - 1)
		if i == 1 {
			stages = append(stages, wf.NewStage(
				WorkType("stage "+str_i), nil, "",
				func(ctx iface.Context) error {
					println("this is stage", str_i)
					return nil
				},
			))
		} else {
			stages = append(stages, wf.NewStage(
				WorkType("stage "+str_i),
				[]WorkType{WorkType("stage " + str_sub_i)}, "",
				func(ctx iface.Context) error {
					println("this is stage", str_i)
					return nil
				},
			))
		}
	}
	for _, stage := range stages {
		fmt.Printf("stage %v, desc: %v, dependOn: %v\n", stage.Name(), stage.Desc(), stage.DependOn())
	}

	builder, err := wf.NewBuilder(stages...)
	if err != nil {
		t.Fatalf("NewBuilder err: %v\n", err)
	}
	workflow, err := builder.Build(wf.WithRunParallel())
	if err != nil {
		t.Fatalf("Build err: %v\n", err)
	}
	workflow.Work(context.Background())
	workflow.Print("", "test.html")
}

func TestType(t *testing.T) {
	stages := buildStage()
	builder, err := wf.NewBuilder(stages...)
	if err != nil {
		t.Fatalf("NewBuilder err: %v\n", err)
	}
	// workflow, err := builder.Build()
	// workflow, err := builder.Build(wf.WithRunParallel())
	workflow, err := builder.Build(wf.WithRunPool(100, 3))
	if err != nil {
		t.Fatalf("Build err: %v\n", err)
	}
	// workflow.Work(context.Background())
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			workflow.Work(context.Background())
			wg.Done()
		}()
	}
	wg.Wait()
	workflow.Print("", "test.html")
}

func buildStage() []iface.Stage[WorkType] {
	stages := []iface.Stage[WorkType]{}
	for i := 0; i < 20; i++ {
		if i == 0 {
			for j := 0; j < 10; j++ {
				str_j := strconv.Itoa(i*10 + j)
				stages = append(stages, wf.NewStage(
					WorkType("stage "+str_j), nil, "",
					func(ctx iface.Context) error {
						println("this is stage", str_j)
						return nil
					},
				))
			}
		} else {
			for j := 0; j < 10; j++ {
				str_j := strconv.Itoa(i*10 + j)
				str_sub_j := strconv.Itoa((i-1)*10 + j)
				stages = append(stages, wf.NewStage(
					WorkType("stage "+str_j),
					[]WorkType{WorkType("stage " + str_sub_j)},
					"",
					func(ctx iface.Context) error {
						time.Sleep(time.Second / 10)
						println("this is stage", str_j, ", dependOn: ", str_sub_j)
						return nil
					},
				))
			}
		}
	}
	return stages
}

func TestXxx(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Wait()
}
