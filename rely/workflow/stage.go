package workflow

import (
	iface "github.com/arknights-w/go-utils/rely/workflow/01_def_iface"
)

type stage[nt iface.Name] struct {
	name nt
	deps []nt
	desc string
	run  func(ctx iface.Context) error
}

func (s *stage[nt]) Name() nt {
	return s.name
}

func (s *stage[nt]) DependOn() []nt {
	return s.deps
}

func (s *stage[nt]) Desc() string {
	return s.desc
}

func (s *stage[nt]) Run(ctx iface.Context) error {
	return s.run(ctx)
}

func NewStage[nt iface.Name](
	name nt,
	depend []nt,
	desc string,
	run func(ctx iface.Context) error,
) iface.Stage[nt] {
	return &stage[nt]{
		name: name,
		deps: depend,
		desc: desc,
		run:  run,
	}
}
