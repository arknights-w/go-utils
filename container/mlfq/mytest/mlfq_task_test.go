package my_test

import (
	"context"
	"fmt"
	"github.com/arknights-w/go-utils/container/mlfq"
	"testing"
	"time"
)

// StepTask 每次调用执行一小段，返回当前 result 和是否结束。
type StepTask func(budget time.Duration) (result int64, isEnd bool)

func makeSumTask(n int64) StepTask {
	var (
		i   int64 = 1
		sum int64 = 0
	)

	return func(budget time.Duration) (int64, bool) {
		start := time.Now()
		for i <= n {
			sum += i
			i++
			if i%10000 == 0 && time.Since(start) >= budget {
				return sum, false
			}
		}
		fmt.Println("任务", n, "完成")
		return sum, true
	}
}

func TestXxx(t *testing.T) {
	ctx := context.Background()

	q, _ := mlfq.NewDefault[StepTask](8, mlfq.WithAutoTick(200*time.Millisecond))
	defer q.Close()

	// 高优先级：紧急+重要
	_, _ = q.Submit(ctx, makeSumTask(50_000_000), mlfq.WithAttributes(mlfq.Attributes{Urgency: 90, Importance: 90}))
	// 中优先级：一般任务
	_, _ = q.Submit(ctx, makeSumTask(10_000_000), mlfq.WithAttributes(mlfq.Attributes{Urgency: 50, Importance: 50}))
	// 低优先级：不紧急不重要
	_, _ = q.Submit(ctx, makeSumTask(5_000_000), mlfq.WithAttributes(mlfq.Attributes{Urgency: 5, Importance: 5}))

	for {
		lease, ok := q.Next(ctx)
		if !ok {
			break
		}

		start := time.Now()
		result, isEnd := lease.Task(lease.Quantum)
		ranFor := time.Since(start)

		_ = q.FeedBack(ctx, lease.Token, mlfq.Feedback{
			RanFor:   ranFor,
			Finished: isEnd,
		})

		if isEnd {
			fmt.Printf("done(level=%d): result=%d\n", lease.Level, result)
		}
	}
}
