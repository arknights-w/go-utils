// Package bitmap 提供一个通用位图结构（任意 N），内部用 []uint64 表示。
//
// 本包主要用于 MultiQueue：记录每个 level 是否非空，并能快速找到最小/最大置位。
package bitmap
