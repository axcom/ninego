package skit

import (
	"fmt"
	"reflect"
	"strings"
)

//整除
func Div(a, b int) int {
	return Trunc(float64(a) / float64(b))
}

//计算X的Y次方（X^Y)
func PowInt(x, y int) int {
	if y <= 0 {
		return 1
	} else {
		if y%2 == 0 {
			sqrt := PowInt(x, y/2)
			return sqrt * sqrt
		} else {
			return PowInt(x, y-1) * x
		}
	}
}

//iif(仿三元表达式)
func If[T any](b bool, t, f T) T {
    if b {
        return t
    }
    return f
}

//删除串中指定长度
func Delete(s *string, start, length int) {
	*s = string((*s)[:start]) + string((*s)[start+length:])
}

//取子串
func SubStr(s string, start, length int) string {
	return string([]byte(s)[start : start+length])
}

//返回(UTF-8)子串sub在字符串str中的起始位置
func Pos(str, sub string) (index int) {
	// 处理空串情况：空sub在任何str中起始位置都是0（符合常见字符串查找惯例）
	if sub == "" {
		return 0
	}
	// 若sub长度大于str，直接返回未找到
	if len(sub) > len(str) {
		return -1
	}

	// 将字符串转为rune切片，确保按Unicode字符遍历（适配UTF-8多字节字符）
	strRunes := []rune(str)
	subRunes := []rune(sub)
	subLen := len(subRunes)
	strLen := len(strRunes)

	// 遍历str的rune切片，查找sub的rune序列
	for i := 0; i <= strLen-subLen; i++ {
		// 比较当前位置开始的rune序列是否与subRunes完全匹配
		match := true
		for j := 0; j < subLen; j++ {
			if strRunes[i+j] != subRunes[j] {
				match = false
				break
			}
		}
		// 找到匹配时，计算当前rune位置对应的字节偏移量
		if match {
			// 将前i个rune转为字符串，其长度即为字节偏移量
			return len(string(strRunes[:i]))
		}
	}

	// 未找到匹配的子串
	return -1
}

//取子串(UTF-8)从0开始，length大于串尾或=0取余下全部；start<0从尾部开始，length<0向前取
func Copy(s string, start, length int) string {
	rs := []rune(s)
	rl := len(rs)
	if start < 0 {
		start = rl - 1 + start
	}
	end := 0
	if length == 0 {
		end = rl
	} else {
		end = start + length
	}
	if start > end {
		start, end = end, start
	}
	if start < 0 {
		start = 0
	} else if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	} else if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

// 非空校验
func isBlank(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

//判断 interface 是否为 nil
func IsNil(x interface{}) bool {
	if x == nil {
		return true
	}
	//问题: 即使接口持有的值为 nil，也不意味着接口本身为 nil。
	rv := reflect.ValueOf(x)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

//是否是整数字符串
func IsInt(s string) bool {
	if Int(s) != 0 {
		return true
	}
	return s != "0"
}

//是否是数字串
func IsNum(s string) bool {
	if Float64(s) != 0.0 {
		return true
	}
	return s != "0"
}

//字符串中是否有中文
func HasChineseChar(s string) bool {
	cc := []rune(string(s))
	for _, c := range cc {
		if c > rune(127) {
			return true
		}
	}
	return false
}


//查找数组、切片、Map中是指定数据索引号(无返-1；Map返0)s
func IndexOf(in interface{}, v interface{}) int {
	targetValue := reflect.ValueOf(in)
	switch reflect.TypeOf(in).Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == v {
				return i
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(v)).IsValid() {
			return 0
		}
	case reflect.String:
		str := in.(string)
		sub := fmt.Sprintf("%v", v)
		index := strings.Index(str, sub)
		if index >= 0 {
			prefix := []byte(str)[0:index]
			rs := []rune(string(prefix))
			return len(rs)
		}
	}
	return -1
}

//查找[]map[key]value列表内指定key及value在数组、切片列表中的index
func LocateMapList(in interface{}, key interface{}, v interface{}) int {
	targetValue := reflect.ValueOf(in)
	switch reflect.TypeOf(in).Kind() {
	case reflect.Array, reflect.Slice:
		if targetValue.Len() > 0 {
			if reflect.TypeOf(targetValue.Index(0).Interface()).Kind() == reflect.Map {
				for i := 0; i < targetValue.Len(); i++ {
					m := reflect.ValueOf(targetValue.Index(i).Interface()).MapIndex(reflect.ValueOf(key))
					if m.IsValid() {
						if m.Interface() == v {
							return i
						}
					}
				}
			}
		}
	}
	return -1
}

/*函数修饰器
newFunc := oldFunc
Decorator(&newFunc, oldFunc, 原函数执行前代码, 原函数执行后代码)
newFunc(params...)
*/
func Decorator(decPtr, fn interface{}, before, after interface{}) {
	var decFunc, tarFunc reflect.Value
	decFunc = reflect.ValueOf(decPtr).Elem()
	tarFunc = reflect.ValueOf(fn)
	v := reflect.MakeFunc(tarFunc.Type(),
		func(in []reflect.Value) (out []reflect.Value) {
			if before != nil {
				reflect.ValueOf(before).Call(in)
			}
			out = tarFunc.Call(in)
			if after != nil {
				reflect.ValueOf(after).Call(in)
			}
			return
		})
	decFunc.Set(v)
}

//In功能
func In(haystack []int, needle int) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}

//支持像解释语言一样的通用 in 功能
func Contains(haystack interface{}, needle interface{}) (bool, error) {
	sVal := reflect.ValueOf(haystack)
	kind := sVal.Kind()
	if kind == reflect.Slice || kind == reflect.Array {
		for i := 0; i < sVal.Len(); i++ {
			if sVal.Index(i).Interface() == needle {
				return true, nil
			}
		}
		return false, nil
	}
	return false, fmt.Errorf("UnSupportHaystack")
}
