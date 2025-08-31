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
	var (
		tmp             any
		err             error
		pre, current    any = o, o
		idx, lastIdx        = 0, len(parts) - 1
		lastKey, preKey     = parts[len(parts)-1], parts[len(parts)-2]
	)
	// 获取前一个元素
	for i, part := range parts[:lastIdx] {
		switch currTyped := current.(type) {
		case map[string]any:
			if tmp = currTyped[part]; tmp == nil {
				currTyped[part] = make(map[string]any)
			}
			current, pre = currTyped[part], current
		case Obj:
			if tmp = currTyped[part]; tmp == nil {
				currTyped[part] = make(map[string]any)
			}
			current, pre = currTyped[part], current
		case []any:
			if idx, err = strconv.Atoi(part); err != nil || idx >= len(currTyped) {
				return fmt.Errorf("set invalid index: %s on array %v ", lastKey, strings.Join(parts[:i], "."))
			} else if tmp = currTyped[idx]; tmp == nil {
				currTyped[idx] = make(map[string]any)
			}
			current, pre = currTyped[idx], current
		default:
			return fmt.Errorf("set on %v[unsupported type]: %T",
				strings.Join(parts[:i], "."), currTyped,
			)
		}
	}
	// 赋值
	switch currTyped := current.(type) {
	case map[string]any:
		currTyped[lastKey] = val
	case Obj:
		currTyped[lastKey] = val
	case []any:
		if idx, err = strconv.Atoi(lastKey); err != nil || idx < 0 || idx > len(currTyped) {
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

func (o Obj) GetByPath(parts []string) (any, bool) {
	current := o[parts[0]]
	for _, part := range parts[1:] {
		switch currTyped := current.(type) {
		case map[string]any:
			current = currTyped[part]
		case Obj:
			current = currTyped[part]
		case []any:
			if idx, err := strconv.Atoi(part); err == nil && idx < len(currTyped) && idx >= 0 {
				current = currTyped[idx]
			} else {
				return nil, false
			}
		case string, bool,
			float64, float32,
			int, int32, int64,
			uint, uint32, uint64:
			current = currTyped
		case nil:
			return nil, true
		default:
			return nil, false
		}
	}
	return current, true
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
	return o.GetByPath(parts)
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
