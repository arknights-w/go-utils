package timewheel

import (
	"time"

	"github.com/arknights-w/go-utils/go_pool"
)

type ticker struct {
	start  int64
	now    int64
	ticker *time.Ticker

	pool *go_pool.Pool
	mgr  *taskMgr

	cancel chan struct{}
}

func (t *ticker) demon(spanTime int64) {
	t.start = time.Now().Unix()
	t.now = t.start
	t.ticker = time.NewTicker(time.Duration(spanTime) * time.Second)
	for {
		select {
		case <-t.cancel:
			t.ticker.Stop()
			return
		case <-t.ticker.C:
			t.now += spanTime
			tasks := t.mgr.GetRunableTasks(t.now)
			for _, task := range tasks {
				t.pool.AddTask(task.fn)
			}
		}
	}
}

func (t *ticker) AddDelayedTask(delay int64, fn func()) (int64, error) {
	select {
	case <-t.cancel:
		return 0, ErrTickerClosed
	default:
		return t.mgr.AddTask(t.now+delay, fn), nil
	}
}

func (t *ticker) AddScheduledTask(execTime int64, fn func()) (int64, error) {
	select {
	case <-t.cancel:
		return 0, ErrTickerClosed
	default:
		return t.mgr.AddTask(execTime, fn), nil
	}
}

func (t *ticker) RemoveTask(id int64) {
	t.mgr.RemoveTask(id)
}

func (t *ticker) Close() error {
	close(t.cancel)
	t.pool.Close()
	return nil
}
