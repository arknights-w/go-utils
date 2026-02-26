// Package ringqueue 提供一个基于环形数组的 FIFO 队列实现。
//
// 特性：
//   - size < cap 时不扩容，通过 (head+size)%cap 追加
//   - 满时按 2 倍扩容，并保持逻辑顺序搬移一次数据
package ringqueue
