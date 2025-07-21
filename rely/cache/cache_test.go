package cache_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/arknights-w/go-utils/rely/cache"
)

func TestXxx(t *testing.T) {
	var a sync.Map
	old, ok := a.LoadOrStore("key", "value")
	fmt.Println("old:", old, "ok:", ok)
	old, ok = a.LoadOrStore("key", "new_value")
	fmt.Println("old:", old, "ok:", ok)
	old, ok = a.LoadOrStore("key", "new_value2")
	fmt.Println("old:", old, "ok:", ok)
}

func TestXxx2(t *testing.T) {
	cache := cache.NewCache()
	cache.Set("key", "value")
	fmt.Printf("key: %v\n", cache.GetStringX("key"))

	cache.SetExpire("key1", 1, 1)
	fmt.Printf("key1: %v\n", cache.GetIntX("key1"))
	time.Sleep(1003 * time.Millisecond)
	fmt.Printf("key1: %v\n", cache.GetIntX("key1"))
}

func TestParallelCache(t *testing.T) {
	cache := cache.NewCache()
	var wg sync.WaitGroup
	for i := range 1000 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if i%3 == 0 {
				cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
				fmt.Printf("Set key%d\n", i)
			}
			if i%3 == 1 {
				time.Sleep(100 * time.Millisecond)
				i = i - 1
				val := cache.GetStringX(fmt.Sprintf("key%d", i))
				fmt.Printf("Get key%d: %s\n", i, val)
			}
			if i%3 == 2 {
				time.Sleep(200 * time.Millisecond)
				i = i - 2
				cache.Del(fmt.Sprintf("key%d", i))
				fmt.Printf("Delete key%d\n", i)
			}
		}()
	}
	wg.Wait()
}

func TestMap(t *testing.T) {
	var syncMap1 = &sync.Map{}
	syncMap1.Store("key1", "value1")
	go func() {
		time.Sleep(1 * time.Second)
		syncMap1.Range(func(key, value any) bool {
			fmt.Printf("syncMap1 key: %v, value: %v\n", key, value)
			return true
		})
	}()
	imap := syncMap1
	if ok := atomic.CompareAndSwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&syncMap1)),
		unsafe.Pointer(imap),
		unsafe.Pointer(&sync.Map{}),
	); !ok {
		imap.Clear()
		return
	}
	imap.Range(func(key, value any) bool {
		fmt.Printf("old key: %v, value: %v\n", key, value)
		return true
	})
	syncMap1.Store("key2", "value2")
	syncMap1.Range(func(key, value any) bool {
		fmt.Printf("new key: %v, value: %v\n", key, value)
		return true
	})

	time.Sleep(2 * time.Second)
}
