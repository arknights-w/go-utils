package obj

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

// RftObj 是一个基于反射实现的通用键值对容器，支持通过点分隔的路径访问嵌套数据
// 例如: obj.GetStr("user.pets.1.name")
//
// 与 Obj 的主要区别:
//
//  1. 实现方式:
//     - Obj: 基于类型断言实现，性能更好但类型支持受限
//     - RftObj: 基于反射实现，性能较低但支持更多类型
//
//  2. 类型兼容性:
//     - Obj: 仅支持 map[string]any, []any 等标准类型
//     - RftObj: 支持任何实现了 map/slice 接口的类型(如 bson.M, bson.A 等)
//
//  3. 使用场景:
//     - Obj: 适用于标准 JSON 反序列化场景
//     - RftObj: 适用于需要处理自定义类型的场景(如 MongoDB BSON)
//
// 支持数据类型:
//   - 基本类型: string, bool, float32/64, int/i32/i64, uint/u32/u64
//   - 复合类型: 任何实现了 map 或 slice 接口的类型
type RftObj map[string]any

var (
	objType     = reflect.TypeOf(RftObj{})
	anyMapType  = reflect.TypeOf(map[string]any{})
	anyListType = reflect.TypeOf([]any{})
	defaultVal  = reflect.ValueOf(nil)
)

func (o RftObj) iget(parts []string, isCreate ...bool) (pre, ret reflect.Value, err error) {
	if len(parts) == 0 {
		return defaultVal, reflect.ValueOf(o), nil
	}
	if len(parts) == 1 {
		return reflect.ValueOf(o), reflect.ValueOf(o[parts[0]]), nil
	}
	var (
		flag           = len(isCreate) > 0 && isCreate[0]
		curVal         = reflect.ValueOf(o)
		preVal, preKey = reflect.Value{}, any(nil)
	)
	// 遍历路径
	for i, key := range parts {
	SWITCH:
		switch curVal.Kind() {
		case reflect.Map:
			tmp := curVal.MapIndex(reflect.ValueOf(key))
			if !tmp.IsValid() {
				if !flag {
					return defaultVal, defaultVal, fmt.Errorf("invalid map key %s at %s",
						key, strings.Join(parts[:i], "."))
				}
				// 创建新的map
				newMap := reflect.MakeMap(anyMapType)
				if err = setMap(curVal, newMap, key); err != nil {
					return defaultVal, defaultVal, err
				}
				curVal = newMap
			} else {
				preVal, preKey, curVal = curVal, key, tmp.Elem()
			}
		case reflect.Slice:
			idx, err := strconv.Atoi(key)
			if err != nil || idx < 0 || idx > curVal.Len() || (idx == curVal.Len() && !flag) {
				return defaultVal, defaultVal, fmt.Errorf("invalid array index %s at %s",
					key, strings.Join(parts[:i], "."))
			} else if idx == curVal.Len() && flag {
				newMap := reflect.MakeMap(anyMapType)
				curVal = reflect.Append(curVal, newMap)
				switch preVal.Kind() {
				case reflect.Map:
					if err = setMap(preVal, curVal, preKey.(string)); err != nil {
						return defaultVal, defaultVal, err
					}
				case reflect.Slice:
					if err = setList(preVal, curVal, preKey.(int)); err != nil {
						return defaultVal, defaultVal, err
					}
				}
			}
			preVal, preKey, curVal = curVal, idx, curVal.Index(idx)
		case reflect.Interface:
			curVal = curVal.Elem()
			goto SWITCH
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
			if i < len(parts)-1 {
				return defaultVal, defaultVal, fmt.Errorf("cannot access %s on %v (type %s)",
					key, strings.Join(parts[:i], "."), curVal.Type(),
				)
			}
		default:
			return defaultVal, defaultVal, fmt.Errorf("cannot access %s on %v (type %s)",
				key, strings.Join(parts[:i], "."), curVal.Type(),
			)
		}
	}
	return preVal, curVal, nil
}

func setMap(imap, val reflect.Value, key string) error {
	if !val.Type().AssignableTo(imap.Type().Elem()) {
		return fmt.Errorf("cannot set val(type %s) key %s on type %s", val.Type(), key, imap.Type())
	}
	imap.SetMapIndex(reflect.ValueOf(key), val)
	return nil
}

func setList(list, val reflect.Value, idx int) error {
	if !val.Type().AssignableTo(list.Type().Elem()) {
		return fmt.Errorf("cannot set val(type %s) index %d on type %s", val.Type(), idx, list.Type())
	}
	list.Index(idx).Set(val)
	return nil
}

// Set 设置指定路径的值.
// key: 支持点分隔的路径，如 "user.address.street" 或 "users.0.name".
// val: 支持的值类型包括:
//   - 基本类型: string, bool, float32/64, int/i32/i64, uint/u32/u64
//   - 复合类型: map[string]any, []any, RftObj
//
// 返回错误:
//   - 当 key 为空时
//   - 当 val 类型不支持时
//   - 当路径中的数组索引无效时
func (o RftObj) Set(key string, val any) error {
	if key == "" {
		return fmt.Errorf("key is empty")
	}
	rftVal := reflect.ValueOf(val)
	switch rftVal.Kind() {
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Array,
		reflect.Chan, reflect.Func, reflect.Pointer, reflect.Struct, reflect.UnsafePointer:
		return fmt.Errorf("unsupported value type: %T", val)
	}
	parts := strings.Split(key, ".")
	if len(parts) <= 1 {
		o[key] = val
		return nil
	}
	lastKey := parts[len(parts)-1]
	pre, _, err := o.iget(parts, true) // 预创建路径
	if err != nil {
		return err
	}
	switch pre.Kind() {
	case reflect.Map:
		pre.SetMapIndex(reflect.ValueOf(lastKey), rftVal)
	case reflect.Slice:
		idx, _ := strconv.Atoi(lastKey)
		pre.Index(idx).Set(rftVal)
	}
	return nil
}

func (o RftObj) GetWithCheck(key string) (any, bool) {
	if key == "" {
		return nil, false
	}
	parts := strings.Split(key, ".")
	if len(parts) == 0 {
		obj, ok := o[key]
		return obj, ok
	}
	_, cur, err := o.iget(parts)
	if err != nil {
		return nil, false
	}
	return cur.Interface(), true
}

// Get 获取指定路径的值，如果不存在则返回 nil
// key: 点分隔的路径
func (o RftObj) Get(key string) any {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return nil
	}
	return v
}

func (o RftObj) GetStr(key string) string {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return ""
	}
	return cast.ToString(v)
}

func (o RftObj) GetBool(key string) bool {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return false
	}
	return cast.ToBool(v)
}

func (o RftObj) GetF64(key string) float64 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToFloat64(v)
}

func (o RftObj) GetF32(key string) float32 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToFloat32(v)
}

func (o RftObj) GetInt(key string) int {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToInt(v)
}

func (o RftObj) GetI32(key string) int32 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToInt32(v)
}

func (o RftObj) GetI64(key string) int64 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToInt64(v)
}

func (o RftObj) GetUint(key string) uint {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToUint(v)
}

func (o RftObj) GetU32(key string) uint32 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToUint32(v)
}

func (o RftObj) GetU64(key string) uint64 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToUint64(v)
}
