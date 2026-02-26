# 性能与适用量级（Bench + 瓶颈分析）

本文档基于本仓库内置 benchmark 的一次运行结果，用于说明 MLFQ 的适用量级、主要瓶颈与可优化点。

## 如何跑 benchmark

> 说明：本环境需要指定 `GOCACHE` 到可写目录。

```bash
cd mlfq
mkdir -p /tmp/gocache
GOCACHE=/tmp/gocache go test -run=^$ -bench . -benchmem ./...
```

## 基准结果（一次采样）

环境：
- go1.24.5 linux/amd64
- CPU: Intel(R) Core(TM) Ultra 7 155H

核心数据（节选）：
- `Scheduler`
  - `BenchmarkScheduler_Submit`：~422 ns/op，`162 B/op`，`2 allocs/op`
  - `BenchmarkScheduler_SteadyState_NextFeedback_Requeue`：~73 ns/op，`0 allocs/op`
  - `BenchmarkScheduler_Parallel_NextFeedback`：~250 ns/op（主要受 mutex 竞争影响）
  - `BenchmarkScheduler_Tick_OLevels`（levels=4096）：~3.9 µs/op（当前 Tick 为 O(levels)）
- `BitMap`
  - `Min/Max` 在 levels=4096 且仅高位有值（sparse worst-case）时：~26 ns/op
- `MultiQueue`
  - `MinNonEmpty` 在 levels=4096 且仅高位有值（sparse worst-case）时：~22 ns/op

## 适用量级（建议）

### Levels（队列层级数）
- 推荐：`8 ~ 4096`
  - levels=4096 时，bitmap 的最坏查找仍在几十纳秒量级；Tick（扫描所有 level）在微秒量级。
- 不建议：极大 levels（例如 10^5 以上）+ 高频 Tick
  - 现实现 Tick 为 O(levels)，levels 过大时 Tick 成本会线性放大。

### Tasks（同时在队列中的任务数）
主要受内存影响，而不是 Next/FeedBack 的纯 CPU 成本：
- 每个任务至少会产生 1 个 `taskState`（`Submit` 的 alloc 主要来自这一步 + map 增长）。
- 经验建议：
  - `<= 10^5`：常见场景较稳妥
  - `10^6`：需要关注内存（并建议 `T` 使用指针或小对象，避免把大 struct 直接拷贝进队列状态）

## 瓶颈与优化空间（按 bench 反推）

### 1) Submit 的分配与 map 成本（主要瓶颈）
证据：`BenchmarkScheduler_Submit` 有 `2 allocs/op`，`162 B/op`。

可优化方向：
- **taskState 复用**：用 `sync.Pool` 复用 `taskState`，避免每次 Submit 都分配。
- **states 预分配**：提供 option（例如 `WithCapacity(n)`）初始化 `states` map 容量，减少增长带来的额外分配。
- **去 map（可选）**：若 token 单调递增且可接受稀疏数组，可用 `[]*taskState` 代替 `map[Token]*taskState`（以 token 为索引），减少哈希开销与分配。

### 2) 并发吞吐受 mutex 竞争限制
证据：单线程 steady-state ~73ns/op，而并行基准 ~250ns/op。

可优化方向（会增加复杂度）：
- 单消费者模型：推荐把 `Next+FeedBack` 放到一个调度 goroutine 内执行（外部并发只 Submit），可显著降低竞争。
- 分段锁/按 level 分锁：提升并发，但实现复杂度与一致性成本更高。
- 提供 `Unsafe` 版本：在外部保证单线程调用时绕开锁（需要 API/类型额外设计）。

### 3) Tick 的线性扫描（O(levels)）
证据：levels=4096 时 Tick ~3.9µs/op，可接受；但 levels 上去会线性变慢。

可优化方向：
- 基于 bitmap 遍历“非空 level 集合”，把 Tick 从 O(levels) 变为 O(non-empty-levels)（对稀疏队列更友好）。

