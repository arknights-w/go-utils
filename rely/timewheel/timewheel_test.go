package timewheel_test

import (
	"testing"
	"time"

	"github.com/arknights-w/go-utils/rely/timewheel"
)

func TestXxx(t *testing.T) {
	tw := timewheel.NewTimeWheel()
	defer tw.Close()
	var ids [10]int64
	var err error
	for i := range 10 {
		ids[i], err = tw.AddDelayedTask(int64(i), func() {
			println("Delayed task executed:", i)
		})
		if err != nil {
			t.Fatalf("Failed to add delayed task %d: %v", i, err)
		}
	}
	for idx, id := range ids {
		if idx%3 == 0 {
			tw.RemoveTask(id)
			println("Removed task with ID:", idx)
		}
	}
	time.Sleep(time.Second * 9) // Wait for tasks to execute
}
