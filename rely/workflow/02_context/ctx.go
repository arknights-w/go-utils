package context

import (
	"context"

	iface "github.com/arknights-w/go-utils/rely/workflow/01_def_iface"
)

type ctx struct {
	kv map[any]any
	iface.Context
}

func (c *ctx) Value(key any) any {
	if v, ok := c.kv[key]; ok {
		return v
	}
	return c.Context.Value(key)
}

func (c *ctx) Set(key, value any) {
	c.kv[key] = value
}

func (c *ctx) Child() iface.Context {
	return &ctx{kv: make(map[any]any), Context: c}
}

func (c *ctx) Err() error {
	if val, ok := c.kv[KEY_ERROR]; !ok {
		return nil
	} else {
		return val.(error)
	}
}

type rowCtx struct {
	kv map[any]any
	context.Context
}

func (c *rowCtx) Value(key any) any {
	if v, ok := c.kv[key]; ok {
		return v
	}
	return c.Context.Value(key)
}

func (c *rowCtx) Set(key, value any) {
	c.kv[key] = value
}

func (c *rowCtx) Child() iface.Context {
	return &ctx{kv: make(map[any]any), Context: c}
}

func (c *rowCtx) Err() error {
	if val, ok := c.kv[KEY_ERROR]; !ok {
		return nil
	} else {
		return val.(error)
	}
}

func NewContext(parent context.Context) iface.Context {
	return &rowCtx{
		kv:      make(map[any]any),
		Context: parent,
	}
}
