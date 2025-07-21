package cache

import (
	"sync"

	"github.com/arknights-w/go-utils/rely/timewheel"
)

type Cache struct {
	imap *sync.Map
	tw   timewheel.TimeWheel
}

type data struct {
	id  int64
	val any
}

func NewCache() *Cache {
	return &Cache{
		imap: &sync.Map{},
		tw:   timewheel.NewTimeWheel(),
	}
}

func (c *Cache) Set(key, val any) {
	if old, ok := c.imap.LoadOrStore(key, &data{val: val}); ok {
		old.(*data).val = val
	}
}

func (c *Cache) SetExpireWithFn(key, val any, expire int64, execFn func()) {
	old, ok := c.imap.LoadOrStore(key, &data{val: val})
	data := old.(*data)
	if ok {
		if data.id > 0 {
			c.tw.RemoveTask(data.id)
		}
		data.val = val
	}
	data.id, _ = c.tw.AddDelayedTask(expire, execFn)
}

func (c *Cache) SetExpire(key, val any, expire int64) {
	c.SetExpireWithFn(key, val, expire, func() {
		c.imap.Delete(key)
	})
}

func (c *Cache) SetExpireAtWithFn(key, val any, execTime int64, execFn func()) {
	old, ok := c.imap.LoadOrStore(key, &data{val: val})
	data := old.(*data)
	if ok {
		if data.id > 0 {
			c.tw.RemoveTask(data.id)
		}
		data.val = val
	}
	data.id, _ = c.tw.AddScheduledTask(execTime, execFn)
}

func (c *Cache) SetExpireAt(key, val any, execTime int64) {
	c.SetExpireAtWithFn(key, val, execTime, func() {
		c.imap.Delete(key)
	})
}

func (c *Cache) Del(key any) {
	if old, ok := c.imap.LoadAndDelete(key); ok {
		data := old.(*data)
		if data.id > 0 {
			c.tw.RemoveTask(data.id)
		}
	}
}

func (c *Cache) Clear() {
	c.imap.Range(func(key, value any) bool {
		if data, ok := value.(*data); ok && data.id > 0 {
			c.tw.RemoveTask(data.id)
		}
		return true
	})
	c.imap.Clear()
	// 由于1.24使用了hashtriemap作为sync.Map的底层实现,clear速度比重新New快
	// if ok := atomic.CompareAndSwapPointer(
	// 	(*unsafe.Pointer)(unsafe.Pointer(&c.imap)),
	// 	unsafe.Pointer(c.imap),
	// 	unsafe.Pointer(&sync.Map{}),
	// ); !ok {
	// 	c.imap.Clear()
	// 	return
	// }
}

func (c *Cache) Get(key any) (val any, ok bool) {
	if old, ok := c.imap.Load(key); ok {
		return old.(*data).val, true
	}
	return nil, false
}

func (c *Cache) GetX(key any) any {
	if val, ok := c.Get(key); ok {
		return val
	}
	return nil
}

func (c *Cache) GetString(key any) (string, bool) {
	return Get[string](c, key)
}

func (c *Cache) GetStringX(key any) string {
	return GetX[string](c, key)
}

func (c *Cache) GetBool(key any) (bool, bool) {
	return Get[bool](c, key)
}

func (c *Cache) GetBoolX(key any) bool {
	return GetX[bool](c, key)
}

func (c *Cache) GetF64(key any) (float64, bool) {
	return Get[float64](c, key)
}

func (c *Cache) GetF64X(key any) float64 {
	return GetX[float64](c, key)
}

func (c *Cache) GetF32(key any) (float32, bool) {
	return Get[float32](c, key)
}

func (c *Cache) GetF32X(key any) float32 {
	return GetX[float32](c, key)
}

func (c *Cache) GetInt(key any) (int, bool) {
	return Get[int](c, key)
}

func (c *Cache) GetIntX(key any) int {
	return GetX[int](c, key)
}

func (c *Cache) GetI64(key any) (int64, bool) {
	return Get[int64](c, key)
}

func (c *Cache) GetI64X(key any) int64 {
	return GetX[int64](c, key)
}

func (c *Cache) GetI32(key any) (int32, bool) {
	return Get[int32](c, key)
}

func (c *Cache) GetI32X(key any) int32 {
	return GetX[int32](c, key)
}

func (c *Cache) GetI16(key any) (int16, bool) {
	return Get[int16](c, key)
}

func (c *Cache) GetI16X(key any) int16 {
	return GetX[int16](c, key)
}

func (c *Cache) GetI8(key any) (int8, bool) {
	return Get[int8](c, key)
}

func (c *Cache) GetI8X(key any) int8 {
	return GetX[int8](c, key)
}

func (c *Cache) GetUInt(key any) (uint, bool) {
	return Get[uint](c, key)
}

func (c *Cache) GetUIntX(key any) uint {
	return GetX[uint](c, key)
}

func (c *Cache) GetU64(key any) (uint64, bool) {
	return Get[uint64](c, key)
}

func (c *Cache) GetU64X(key any) uint64 {
	return GetX[uint64](c, key)
}

func (c *Cache) GetU32(key any) (uint32, bool) {
	return Get[uint32](c, key)
}

func (c *Cache) GetU32X(key any) uint32 {
	return GetX[uint32](c, key)
}

func (c *Cache) GetU16(key any) (uint16, bool) {
	return Get[uint16](c, key)
}

func (c *Cache) GetU16X(key any) uint16 {
	return GetX[uint16](c, key)
}

func (c *Cache) GetU8(key any) (uint8, bool) {
	return Get[uint8](c, key)
}

func (c *Cache) GetU8X(key any) uint8 {
	return GetX[uint8](c, key)
}

func Get[itype any](c *Cache, key any) (itype, bool) {
	var zero itype
	if val, ok := c.Get(key); !ok {
		return zero, false
	} else if v, ok := val.(itype); ok {
		return v, true
	}
	return zero, false
}

func GetX[itype any](c *Cache, key any) itype {
	if val, ok := Get[itype](c, key); ok {
		return val
	}
	var zero itype
	return zero
}
