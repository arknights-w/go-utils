package go_pool_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pool "github.com/arknights-w/go-utils/go_pool"
)

func TestScheduler(t *testing.T) {
	scheduler := pool.NewPool(0, 0)
	waitGroup := sync.WaitGroup{}
	count := int32(0)
	defer scheduler.Close()
	for i := range 100 {
		waitGroup.Add(1)
		scheduler.AddTask(func() {
			time.Sleep(time.Second)
			atomic.AddInt32(&count, 1)
			defer waitGroup.Done()
			t.Logf("count: %d, task %d", count, i)
		})
	}
	waitGroup.Wait()
}

func TestTaskGroup(t *testing.T) {
	scheduler := pool.NewPool(0, 0)
	defer scheduler.Close()
	group1 := scheduler.Group()
	group2 := scheduler.Group()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := range 100 {
			time.Sleep(time.Second / 5)
			group2.AddTask(func() {
				time.Sleep(time.Second * 4 / 5)
				t.Logf("group2: %d", i)
			})
		}
		wg.Done()
	}()
	go func() {
		for i := range 100 {
			time.Sleep(time.Second / 5)
			group1.AddTask(func() {
				time.Sleep(time.Second * 4 / 5)
				t.Logf("group1: %d", i)
			})
		}
		wg.Done()
	}()
	wg.Wait()
	println("send done")
	group1.Wait()
	group2.Wait()
	println("exec done")
}

func TestPoolV3(t *testing.T) {
	pool := pool.NewPool(10, 3)
	defer pool.Close()
	waitGroup := pool.Group()
	count := int32(0)
	for i := range 1000 {
		waitGroup.AddTask(func() {
			time.Sleep(time.Millisecond * 100)
			atomic.AddInt32(&count, 1)
			t.Logf("count: %d, task %d", count, i)
		})
	}
	waitGroup.Wait()
}
