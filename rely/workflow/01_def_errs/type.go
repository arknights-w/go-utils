package errs

import "fmt"

var _ error = (*WorkflowErr)(nil)

type WorkflowErr struct {
	code int
	msg  string
	desc string
}

func (e *WorkflowErr) Error() string {
	return fmt.Sprintf("{code: %d, msg: \"%s\", desc: \"%s\"}", e.code, e.msg, e.desc)
}

func (e *WorkflowErr) WithDesc(desc string) *WorkflowErr {
	var newErr = *e
	newErr.desc = desc
	return &newErr
}

func (e *WorkflowErr) Code() int {
	return e.code
}

func (e *WorkflowErr) Msg() string {
	return e.msg
}

func (e *WorkflowErr) Desc() string {
	return e.desc
}
