package timewheel

import (
	"fmt"
	"time"

	"github.com/arknights-w/go-utils/go_pool"
)

var (
	ErrTickerClosed = fmt.Errorf("ticker is closed")
)

type TimeWheel interface {
	AddDelayedTask(delay int64, fn func()) (int64, error)
	AddScheduledTask(execTime int64, fn func()) (int64, error)
	RemoveTask(id int64)

	Close() error
}

func NewTimeWheel() TimeWheel {
	t := &ticker{
		start:  time.Now().Unix(),
		now:    time.Now().Unix(),
		pool:   go_pool.NewPool(5, 1),
		mgr:    &taskMgr{},
		cancel: make(chan struct{}),
	}
	go t.demon(1) // 1 second time wheel
	return t
}
