// Package mlfq 提供一个可复用的多级反馈队列（MLFQ）调度数据结构。
//
// 说明：
//   - 本模块只负责“调度”，不负责实际执行任务；因此实际运行耗时需要由调用方测量后通过 FeedBack 回传。
//   - Next 返回 Lease（包含 Token 与本次建议时间片 Quantum）；FeedBack 通过 Token 关联上一次 Next 发放的任务。
//   - 可选开启自动 Tick（WithAutoTick），用于周期性老化提升；使用后务必在退出时 Close 以停止后台 goroutine。
package mlfq
