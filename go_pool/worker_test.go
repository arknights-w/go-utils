package go_pool_test

import (
	"testing"
	"time"

	pool "github.com/arknights-w/go-utils/go_pool"
)

func TestWorker(t *testing.T) {
	worker := pool.NewWorker(1)
	go func() {
		for i := range 100 {
			worker.AddTask(func() {
				time.Sleep(time.Millisecond * 10)
				t.Log("task_", i)
			})
			time.Sleep(time.Millisecond * 1)
		}
		worker.Close()
	}()
	worker.Wait()
}

func TestChan(t *testing.T) {
	ch := make(channel, 1)
	println(ch.send(1))
	println(ch.send(2))
}

type channel chan int

func (ch channel) send(i int) bool {
	select {
	case ch <- i:
		return true
	default:
		return false
	}
}
