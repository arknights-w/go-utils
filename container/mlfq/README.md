# mlfq

一个可复用的 Go 多级反馈队列（MLFQ）调度数据结构实现：负责任务入队、选择下一个可运行任务、执行后反馈调整优先级、以及 Tick 老化提升与统计快照。

## 快速开始

```go
package main

import (
	"context"
	"fmt"
	"time"

	"mlfq"
)

func main() {
	ctx := context.Background()

	s, _ := mlfq.NewDefault[string](8, mlfq.WithAutoTick(200*time.Millisecond))
	defer s.Close()

	_, _ = s.Submit(ctx, "a", mlfq.WithAttributes(mlfq.Attributes{Urgency: 80, Importance: 50}))
	_, _ = s.Submit(ctx, "b")

	lease, ok := s.Next(ctx)
	if !ok {
		return
	}

	fmt.Println("run:", lease.Task, "level:", lease.Level, "q:", lease.Quantum)

	_ = s.FeedBack(ctx, lease.Token, mlfq.Feedback{
		RanFor:   lease.Quantum,
		Finished: true,
	})

	// Tick 也可以外部手动调用（即使开启了 auto tick）：
	s.Tick(ctx, time.Now())
	fmt.Printf("%+v\n", s.Stats(ctx))
}
```

## 用函数/闭包作为任务（带切片执行）

MLFQ 是泛型的，你可以把任务定义成一个函数类型。下面是一个常见写法：任务每次被调度只做“一小段工作”，直到返回 `isEnd=true` 才算完成。

示例：把 `1..n` 的加法拆成多次执行（每次最多跑 `Lease.Quantum` 预算），返回 `(result, isEnd)`；`Feedback.RanFor` 使用真实执行时间，`Finished=isEnd`。

```go
package main

import (
	"context"
	"fmt"
	"time"

	"mlfq"
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
			if time.Since(start) >= budget {
				return sum, false
			}
		}
		return sum, true
	}
}

func main() {
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
```

## 文档

- 设计说明：`mlfq/DESIGN.md`
- 性能说明：`mlfq/PERF.md`

## 子包

- `mlfq/bitmap`：位图
- `mlfq/ringqueue`：环形队列
- `mlfq/multiqueue`：多级队列
