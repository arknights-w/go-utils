// Package multiqueue 提供一个多级队列：BitMap + 多个 ringqueue.Queue。
//
// 它用于 MLFQ 调度器快速定位最小/最大非空 level，并在各 level 内维持 FIFO。
package multiqueue
