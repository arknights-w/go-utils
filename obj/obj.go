package obj

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

// Obj 是一个通用的键值对容器，支持通过点分隔的路径访问嵌套数据
// 例如: obj.GetStr("user.pets.1.name")
//
// 支持数据类型:
//   - 基本类型: string, bool, float32/64, int/i32/i64, uint/u32/u64
//   - 复合类型: map[string]any, []any, Obj
//
// 为什么不是gjson:
// gjson 无法兼顾作为字段进行序列化或者反序列化的场景.
// obj 则底层完全依赖 map 实现, 能够很好的兼容 map[string]any 的场景.
// obj 适用于数据格式过于个性化，并且层级深，需要作为字段进行快速序列化与反序列的场景.
//
// obj 还尚处于早期阶段，完全遵从bytes->map反序列化得到的数据类型，对于更为具体的复合类型支持有限
// (仅str,f64,[]any,map[str]any, 如 []string, []int, map[string]string 等均未支持)
type Obj map[string]any

// var (
// 	objType     = reflect.TypeOf(Obj{})
// 	anyMapType  = reflect.TypeOf(map[string]any{})
// 	anyListType = reflect.TypeOf([]any{})
// 	defaultVal  = reflect.ValueOf(nil)
// )

// func (o Obj) iget(parts []string, isCreate ...bool) (pre, ret reflect.Value, err error) {
// 	if len(parts) == 0 {
// 		return defaultVal, reflect.ValueOf(o), nil
// 	}
// 	if len(parts) == 1 {
// 		return reflect.ValueOf(o), reflect.ValueOf(o[parts[0]]), nil
// 	}
// 	var (
// 		flag           = len(isCreate) > 0 && isCreate[0]
// 		curVal         = reflect.ValueOf(o)
// 		preVal, preKey = reflect.Value{}, any(nil)
// 	)
// 	// 遍历路径
// 	for i, key := range parts {
// 	SWITCH:
// 		switch curVal.Kind() {
// 		case reflect.Map:
// 			tmp := curVal.MapIndex(reflect.ValueOf(key))
// 			if !tmp.IsValid() {
// 				if !flag {
// 					return defaultVal, defaultVal, fmt.Errorf("invalid map key %s at %s",
// 						key, strings.Join(parts[:i], "."))
// 				}
// 				// 创建新的map
// 				newMap := reflect.MakeMap(anyMapType)
// 				curVal.SetMapIndex(reflect.ValueOf(key), newMap)
// 				curVal = newMap
// 			} else {
// 				preVal, preKey, curVal = curVal, key, tmp.Elem()
// 			}
// 		case reflect.Slice:
// 			idx, err := strconv.Atoi(key)
// 			if err != nil || idx < 0 || idx > curVal.Len() || (idx == curVal.Len() && !flag) {
// 				return defaultVal, defaultVal, fmt.Errorf("invalid array index %s at %s",
// 					key, strings.Join(parts[:i], "."))
// 			} else if idx == curVal.Len() && flag {
// 				newMap := reflect.MakeMap(anyMapType)
// 				curVal = reflect.Append(curVal, newMap)
// 				switch preVal.Kind() {
// 				case reflect.Map:
// 					preVal.SetMapIndex(reflect.ValueOf(preKey), curVal)
// 				case reflect.Slice:
// 					preVal.Index(preKey.(int)).Set(curVal)
// 				}
// 			}
// 			preVal, preKey, curVal = curVal, idx, curVal.Index(idx)
// 		case reflect.Interface:
// 			curVal = curVal.Elem()
// 			goto SWITCH
// 		case reflect.Uint, reflect.Uint32, reflect.Uint64,
// 			reflect.Int, reflect.Int32, reflect.Int64,
// 			reflect.Float32, reflect.Float64,
// 			reflect.Bool, reflect.String:
// 			if i < len(parts)-1 {
// 				return defaultVal, defaultVal, fmt.Errorf("cannot access %s on %v (type %s)",
// 					key, strings.Join(parts[:i], "."), curVal.Type(),
// 				)
// 			}
// 		default:
// 			return defaultVal, defaultVal, fmt.Errorf("cannot access %s on %v (type %s)",
// 				key, strings.Join(parts[:i], "."), curVal.Type(),
// 			)
// 		}
// 	}
// 	return preVal, curVal, nil
// }

func (o Obj) get(parts []string, isCreate ...bool) (pre, ret any, err error) {
	if len(parts) == 0 {
		return nil, o, nil
	}
	if len(parts) == 1 {
		return o, o[parts[0]], nil
	}
	var (
		flag               = len(isCreate) > 0 && isCreate[0]
		curVal, idx        = any(o), 0
		preVal, preKey any = o, nil
	)
	for i, key := range parts {
		switch currTyped := curVal.(type) {
		case map[string]any:
			if tmp := currTyped[key]; tmp == nil {
				if !flag {
					return nil, nil, fmt.Errorf("invalid map key %s at %s",
						key, strings.Join(parts[:i], "."))
				}
				currTyped[key] = make(map[string]any)
			}
			preVal, preKey, curVal = curVal, key, currTyped[key]
		case Obj:
			if tmp := currTyped[key]; tmp == nil {
				if !flag {
					return nil, nil, fmt.Errorf("invalid map key %s at %s",
						key, strings.Join(parts[:i], "."))
				}
				currTyped[key] = make(map[string]any)
			}
			preVal, preKey, curVal = curVal, key, currTyped[key]
		case []any:
			if idx, err = strconv.Atoi(key); err != nil ||
				idx > len(currTyped) || (idx == len(currTyped) && !flag) {
				return nil, nil, fmt.Errorf("set invalid index: %s on array %v ", key, strings.Join(parts[:i], "."))
			} else if idx == len(currTyped) && flag {
				currTyped = append(currTyped, make(map[string]any))
				switch preTyped := preVal.(type) {
				case map[string]any:
					preTyped[preKey.(string)] = currTyped
				case Obj:
					preTyped[preKey.(string)] = currTyped
				case []any:
					preTyped[preKey.(int)] = currTyped
				}
			}
			preVal, preKey, curVal = curVal, idx, currTyped[idx]
		case uint, uint32, uint64,
			int, int32, int64,
			float32, float64,
			bool, string:
			if i < len(parts)-1 {
				return nil, nil, fmt.Errorf("cannot access %s on %v (type %T)",
					key, strings.Join(parts[:i], "."), curVal,
				)
			}
		default:
			return nil, nil, fmt.Errorf("cannot access %s on %v (type %T)",
				key, strings.Join(parts[:i], "."), preVal,
			)
		}
	}
	return preVal, curVal, nil
}

// Set 设置指定路径的值.
// key: 支持点分隔的路径，如 "user.address.street" 或 "users.0.name".
// val: 支持的值类型包括:
//   - 基本类型: string, bool, float32/64, int/i32/i64, uint/u32/u64
//   - 复合类型: map[string]any, []any, Obj
//
// 返回错误:
//   - 当 key 为空时
//   - 当 val 类型不支持时
//   - 当路径中的数组索引无效时
func (o Obj) Set(key string, val any) error {
	if key == "" {
		return fmt.Errorf("key is empty")
	}
	switch val.(type) {
	case string, bool,
		float64, float32,
		int, int32, int64,
		uint, uint32, uint64,
		map[string]any, []any, Obj:
		// pass
	default:
		return fmt.Errorf("unsupported value type: %T", val)
	}
	parts := strings.Split(key, ".")
	if len(parts) <= 1 {
		o[key] = val
		return nil
	}
	preKey, lastKey, lastIdx := parts[len(parts)-2], parts[len(parts)-1], len(parts)-1
	pre, cur, err := o.get(parts[:len(parts)-1], true) // 预创建路径
	if err != nil {
		return err
	}
	// 赋值
	switch currTyped := cur.(type) {
	case map[string]any:
		currTyped[lastKey] = val
	case Obj:
		currTyped[lastKey] = val
	case []any:
		if idx, err := strconv.Atoi(lastKey); err != nil || idx < 0 || idx > len(currTyped) {
			return fmt.Errorf("set invalid index: %s on array %v ", lastKey, strings.Join(parts[:lastIdx], "."))
		} else if idx == len(currTyped) {
			currTyped = append(currTyped, val)
			switch preTyped := pre.(type) {
			case map[string]any:
				preTyped[preKey] = currTyped
			case Obj:
				preTyped[preKey] = currTyped
			case []any:
				pidx, _ := strconv.Atoi(preKey)
				preTyped[pidx] = currTyped
			}
		} else {
			currTyped[idx] = val
		}
	default:
		return fmt.Errorf("set on %v[unsupported type]: %T", strings.Join(parts[:lastIdx], "."), currTyped)
	}
	return nil
}

func (o Obj) GetWithCheck(key string) (any, bool) {
	if key == "" {
		return nil, false
	}
	parts := strings.Split(key, ".")
	if len(parts) == 0 {
		obj, ok := o[key]
		return obj, ok
	}
	_, cur, err := o.get(parts)
	if err != nil {
		return nil, false
	}
	return cur, true
}

// Get 获取指定路径的值，如果不存在则返回 nil
// key: 点分隔的路径
func (o Obj) Get(key string) any {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return nil
	}
	return v
}

func (o Obj) GetStr(key string) string {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return ""
	}
	return cast.ToString(v)
}

func (o Obj) GetBool(key string) bool {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return false
	}
	return cast.ToBool(v)
}

func (o Obj) GetF64(key string) float64 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToFloat64(v)
}

func (o Obj) GetF32(key string) float32 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToFloat32(v)
}

func (o Obj) GetInt(key string) int {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToInt(v)
}

func (o Obj) GetI32(key string) int32 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToInt32(v)
}

func (o Obj) GetI64(key string) int64 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToInt64(v)
}

func (o Obj) GetUint(key string) uint {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToUint(v)
}

func (o Obj) GetU32(key string) uint32 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToUint32(v)
}

func (o Obj) GetU64(key string) uint64 {
	v, ok := o.GetWithCheck(key)
	if !ok {
		return 0
	}
	return cast.ToUint64(v)
}
