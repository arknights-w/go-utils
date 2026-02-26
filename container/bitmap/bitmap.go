package bitmap

import "math/bits"

// BitMap 是一个支持任意 N 的位图，内部以 []uint64 存储。
//
// 约定：
//   - 只允许设置 [0, N) 范围内的 bit，越界会 panic（用于暴露调用方 bug）。
//   - Min/Max 返回最小/最大置位索引；若全空则 ok=false。
type BitMap struct {
	n            int
	words        []uint64
	nonZeroWords int
}

// New 创建一个包含 n 个 bit 的位图。
func New(n int) BitMap {
	if n <= 0 {
		return BitMap{n: 0, words: nil}
	}
	return BitMap{
		n:     n,
		words: make([]uint64, (n+63)/64),
	}
}

func (b *BitMap) N() int { return b.n }

func (b *BitMap) Any() bool { return b.nonZeroWords > 0 }

// Reset 清空所有 bit。
func (b *BitMap) Reset() {
	for i := range b.words {
		b.words[i] = 0
	}
	b.nonZeroWords = 0
}

func (b *BitMap) Words() []uint64 {
	if len(b.words) == 0 {
		return nil
	}
	out := make([]uint64, len(b.words))
	copy(out, b.words)
	return out
}

func (b *BitMap) IsSet(k int) bool {
	wi, mask := b.idxMask(k)
	return (b.words[wi] & mask) != 0
}

// Set 将第 k 位设置为 1。
func (b *BitMap) Set(k int) {
	wi, mask := b.idxMask(k)
	before := b.words[wi]
	after := before | mask
	if before == after {
		return
	}
	b.words[wi] = after
	if before == 0 {
		b.nonZeroWords++
	}
}

// Clear 将第 k 位清零。
func (b *BitMap) Clear(k int) {
	wi, mask := b.idxMask(k)
	before := b.words[wi]
	after := before &^ mask
	if before == after {
		return
	}
	b.words[wi] = after
	if after == 0 {
		b.nonZeroWords--
	}
}

// Min 返回最小置位的索引。
func (b *BitMap) Min() (int, bool) {
	if b.nonZeroWords == 0 {
		return 0, false
	}
	for wi, w := range b.words {
		if w == 0 {
			continue
		}
		off := bits.TrailingZeros64(w)
		return wi*64 + off, true
	}
	return 0, false
}

// Max 返回最大置位的索引。
func (b *BitMap) Max() (int, bool) {
	if b.nonZeroWords == 0 {
		return 0, false
	}
	for wi := len(b.words) - 1; wi >= 0; wi-- {
		w := b.words[wi]
		if w == 0 {
			continue
		}
		off := bits.Len64(w) - 1
		return wi*64 + off, true
	}
	return 0, false
}

func (b *BitMap) idxMask(k int) (int, uint64) {
	if k < 0 || k >= b.n {
		panic("mlfq: bitmap index out of range")
	}
	wi := k >> 6
	bi := k & 63
	return wi, uint64(1) << uint(bi)
}
