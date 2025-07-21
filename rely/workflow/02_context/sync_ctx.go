package context

import (
	"context"
	"sync"

	iface "github.com/arknights-w/go-utils/rely/workflow/01_def_iface"
)

type syncCtx struct {
	kv *sync.Map
	iface.Context
}

func (c *syncCtx) Value(key any) any {
	if v, ok := c.kv.Load(key); ok {
		return v
	}
	return c.Context.Value(key)
}

func (c *syncCtx) Set(key, value any) {
	c.kv.Store(key, value)
}

func (c *syncCtx) Child() iface.Context {
	return &syncCtx{kv: &sync.Map{}, Context: c}
}

func (c *syncCtx) Err() error {
	if val, ok := c.kv.Load(KEY_ERROR); !ok {
		return nil
	} else {
		return val.(error)
	}
}

type rowSyncCtx struct {
	kv *sync.Map
	context.Context
}

func (c *rowSyncCtx) Value(key any) any {
	if v, ok := c.kv.Load(key); ok {
		return v
	}
	return c.Context.Value(key)
}

func (c *rowSyncCtx) Set(key, value any) {
	c.kv.Store(key, value)
}

func (c *rowSyncCtx) Child() iface.Context {
	return &syncCtx{kv: &sync.Map{}, Context: c}
}

func (c *rowSyncCtx) Err() error {
	if val, ok := c.kv.Load(KEY_ERROR); !ok {
		return nil
	} else {
		return val.(error)
	}
}

func NewSyncContext(parent context.Context) iface.Context {
	return &rowSyncCtx{
		kv:      &sync.Map{},
		Context: parent,
	}
}
