# 设计说明（MLFQ）

本模块实现一个 MLFQ（Multi-Level Feedback Queue）调度数据结构，核心组件：

1. **BitMap**：独立位图结构（`mlfq/bitmap`），内部用 `[]uint64` 表示任意 N 个 bit；用于快速定位最小/最大置位（对应最小/最大非空队列）。
2. **ringQueue**：单队列的环形数组实现（`mlfq/ringqueue`）；`size < cap` 时不扩容，满时按 2 倍扩容并保持逻辑顺序搬移一次数据。
3. **MultiQueue**：多级队列（`mlfq/multiqueue`）：`BitMap + []ringQueue`，用于维护每个 level 的 FIFO 队列与非空索引。
4. **Policy**：策略接口：决定 Submit 初始 level、Next 取哪个 level、每个 level 的时间片（quantum）、反馈后升/降级，以及 Tick 老化提升。
5. **Scheduler**：对外的 `MLFQ[T]` 实现：线程安全（mutex），Next 返回 `Lease{Token,...}`，FeedBack 用 Token 定位任务并调整。

默认约定：
- Level 编号：`0` 最高优先级；Next 默认取最小非空 level。
- FeedBack 通过 `Token` 关联“刚刚 Next 出来的任务”。

## 目录结构

```
mlfq/
  bitmap/       位图数据结构
  ringqueue/    环形队列
  multiqueue/   多级队列（位图 + 多队列）
  policy.go     策略接口 + 默认策略
  scheduler.go  MLFQ 实现（对外返回接口）
  mlfq.go       对外接口与类型
```
