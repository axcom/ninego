package syscl

import (
	"fmt"
	"strings"
)

//模拟delphi的TList类、TStringList类(name,values; 无Object-->参见StrMapList)
/*
package main

import (
    "fmt"
    "your-project-path/classes" // 替换为你的实际包路径
)

func main() {
    // 1. 创建一个存储 int 类型的列表
    intList := classes.NewList[int]()
    intList.Add(10)
    intList.Add(20)
    intList.Add(10)
    fmt.Println("Int List Count:", intList.Count()) // 输出: 3

    intList.Insert(1, 15) // 在索引1的位置插入15
    fmt.Println("Index of 20:", intList.IndexOf(20)) // 输出: 2

    intList.Remove(10) // 删除第一个出现的10
    fmt.Println("After Remove(10), Count:", intList.Count()) // 输出: 2

    // 使用 Walk 遍历
    fmt.Println("Int List elements:")
    intList.Walk(func(i int, v int) {
        fmt.Printf("  [%d]: %d\n", i, v)
    })
    // 输出:
    //   [0]: 15
    //   [1]: 20

    // 使用 Sort 排序
    intList.Sort(func(a, b int) bool {
        return a > b // 降序排序
    })
    fmt.Println("After sorting in descending order:")
    intList.Walk(func(i int, v int) {
        fmt.Printf("  [%d]: %d\n", i, v)
    })

    // 2. 创建一个存储 string 类型的列表
    strList := classes.NewList[string]()
    strList.Add("Apple")
    strList.Add("Banana")
    strList.Add("Cherry")

    fmt.Println("\nString List elements:")
    strList.Walk(func(i int, v string) {
        fmt.Printf("  [%d]: %s\n", i, v)
    })

    // 错误示例（编译器会报错）
    // intList.Add("hello") //  cannot use "hello" (type string) as type int in argument to intList.Add
}
*/

// List 是一个泛型列表结构，可以存储任意指定类型的元素。
// T 是类型参数，代表列表中元素的类型。
type List[T comparable] struct {
    Items []T
}

// NewList 创建并返回一个新的、空的泛型 List。
func NewList[T comparable]() *List[T] {
    return &List[T]{Items: make([]T, 0)}
}

// Count 返回列表中元素的数量。
func (li *List[T]) Count() int {
    return len(li.Items)
}

// Swap 交换列表中索引 i 和 j 位置的元素。
// 为了实现 sort.Interface 接口。
func (li *List[T]) Swap(i, j int) {
    li.Items[i], li.Items[j] = li.Items[j], li.Items[i]
}

// Less 比较列表中索引 i 和 j 位置的元素。
// 注意：此实现总是返回 false。如需排序，你需要根据具体类型 T 提供一个真正的比较逻辑，
// 或者定义一个接受比较函数的 Sort 方法。
// 为了实现 sort.Interface 接口。
func (li *List[T]) Less(i, j int) bool {
    // 泛型无法直接比较任意类型 T，因此默认返回 false。
    // 这使得 List 可以“满足”sort.Interface，但直接调用 sort.Sort(li) 不会改变顺序。
    // 建议使用下面的 Sort 方法。
    return false
}

// Sort 使用提供的比较函数对列表进行排序。
// 这是一种更灵活、类型安全的排序方式。
func (li *List[T]) Sort(less func(a, b T) bool) {
    // 简单的冒泡排序示例，你可以替换为更高效的排序算法
    n := li.Count()
    for i := 0; i < n-1; i++ {
        for j := 0; j < n-i-1; j++ {
            if less(li.Items[j], li.Items[j+1]) {
                li.Swap(j, j+1)
            }
        }
    }
}

// Insert 在指定的索引位置插入一个元素。
// 如果 index 超出范围 (0 <= index <= Count())，则 panic。
func (li *List[T]) Insert(index int, v T) {
    if index < 0 || index > li.Count() {
        panic(fmt.Sprintf("List.Insert: index %d out of range [0, %d]", index, li.Count()))
    }
    // 创建一个新切片，容量+1
    newItems := make([]T, 0, li.Count()+1)
    // 拼接：前半部分 + 新元素 + 后半部分
    newItems = append(append(newItems, li.Items[:index]...), v)
    newItems = append(newItems, li.Items[index:]...)
    li.Items = newItems
}

// Add 在列表的末尾添加一个元素。
func (li *List[T]) Add(v T) {
    li.Items = append(li.Items, v)
}

// Delete 删除指定索引位置的元素。
// 如果 index 超出范围，则 panic。
func (li *List[T]) Delete(index int) {
    if index < 0 || index >= li.Count() {
        panic(fmt.Sprintf("List.Delete: index %d out of range [0, %d)", index, li.Count()))
    }
    li.Items = append(li.Items[:index], li.Items[index+1:]...)
}

// Remove 删除列表中第一个出现的指定元素 v。
// 元素的比较使用 == 操作符。
// 如果元素不存在，则列表保持不变。
func (li *List[T]) Remove(v T) {
    if i := li.IndexOf(v); i != -1 {
        li.Delete(i)
    }
}

// Clear 清空列表中的所有元素。
func (li *List[T]) Clear() {
    // 直接指向一个新的空切片，让旧切片被GC回收
    li.Items = make([]T, 0)
}

// IndexOf 返回元素 v 在列表中第一次出现的索引。
// 元素的比较使用 == 操作符。
// 如果未找到，则返回 -1。
func (li *List[T]) IndexOf(v T) int {
    for i, item := range li.Items {
        // 泛型 T 在这里可以直接使用 == 进行比较
        if item == v {
            return i
        }
    }
    return -1
}
/*在 Go 1.18+ 中，你可以使用 comparable 约束来实现这一点。comparable 是一个预定义的接口，它包含了所有可以使用 == 和 != 进行比较的类型（如 int, string, struct（如果其所有字段都可比较）等）
在定义 List 结构体时，为 T 增加 comparable 约束
type List[T comparable] struct {
    Items []T
}
*/

// Walk 遍历列表中的所有元素，并对每个元素执行 do 函数。
// do 函数的参数为元素的索引和值。
func (li *List[T]) Walk(do func(i int, v T)) {
    for i, v := range li.Items {
        do(i, v)
    }
}

// Get 获取指定索引位置的元素。
// 如果 index 超出范围，则 panic。
func (li *List[T]) Get(index int) T {
    if index < 0 || index >= li.Count() {
        panic(fmt.Sprintf("List.Get: index %d out of range [0, %d)", index, li.Count()))
    }
    return li.Items[index]
}

// Set 设置指定索引位置的元素的值。
// 如果 index 超出范围，则 panic。
func (li *List[T]) Set(index int, v T) {
    if index < 0 || index >= li.Count() {
        panic(fmt.Sprintf("List.Set: index %d out of range [0, %d)", index, li.Count()))
    }
    li.Items[index] = v
}

// 迭代器：Iter()（优先级更高，range 会优先使用）
func (li *List[T]) Iter() func() (T, bool) {
	i := 0
	return func() (T, bool) {
		if i >= len(li.Items) {
			var zero T // 返回元素类型的零值
			return zero, false
		}
		item := li.Items[i]
		i++
		return item, true // 加标记区分迭代方式
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
var DefaultDelimiter string = ","

type Strings struct {
	Items     []string
	Delimiter string `json:"-"`
}

func NewStringList(sls ...string) *Strings {
	if len(sls) > 1 {
		return &Strings{Items: strings.Split(sls[0], sls[1]), Delimiter: DefaultDelimiter}
	} else if len(sls) > 0 {
		return &Strings{Items: strings.Split(sls[0], DefaultDelimiter), Delimiter: DefaultDelimiter}
	}
	return &Strings{Items: make([]string, 0), Delimiter: DefaultDelimiter}
}

func (li *Strings) Split(s string, sep ...string) int {
	if len(sep) > 0 {
		li.Items = strings.Split(s, sep[0])
	} else {
		li.Items = strings.Split(s, li.Delimiter)
	}
	return len(li.Items)
}

func (li *Strings) Text(sep ...string) string {
	s := li.Delimiter
	if len(sep) > 0 {
		s = sep[0]
	}
	str := ""
	for i := 0; i < li.Count(); i++ {
		if i == 0 {
			str = li.Items[0]
		} else {
			str += s + li.Items[i]
		}
	}
	return str
}

func (li *Strings) Strings(index int) string {
	return li.Items[index]
}

func (li *Strings) Swap(i, j int) {
	li.Items[i], li.Items[j] = li.Items[j], li.Items[i]
}

func (li *Strings) Less(i, j int) bool {
	return li.Items[i] < li.Items[j]
}

func (li *Strings) Insert(index int, s string, v ...string) {
	rear := append([]string{}, (li.Items)[index:]...)
	//li.Items = append(append((li.Items)[:index], v), rear...)
	if len(v) == 0 || v[0] == "" {
		li.Items = append(append((li.Items)[:index], s), rear...)
	} else {
		li.Items = append(append((li.Items)[:index], s+"="+v[0]), rear...)
	}
}

func (li *Strings) Add(s string, v ...string) {
	if len(v) == 0 || v[0] == "" {
		li.Items = append(li.Items, s)
	} else {
		li.Items = append(li.Items, s+"="+v[0])
	}

}

func (li *Strings) Delete(index int) {
	li.Items = append(li.Items[:index], li.Items[index+1:]...)
}

func (li *Strings) Clear() {
	li.Items = append([]string{})
}

func (li *Strings) Count() int {
	return len(li.Items)
}

func (li *Strings) Walk(do func(i int, v string)) {
	for i, v := range li.Items {
		do(i, v)
	}
}

func (li *Strings) IndexOf(v string) int {
	for i, m := range li.Items {
		if m == v {
			return i
		}
	}
	return -1
}

func (li *Strings) IndexOfName(v string) int {
	for i := 0; i < li.Count(); i++ {
		if li.Names(i) == v {
			return i
		}
	}
	return -1
}

func (li *Strings) Values(v string) string {
	i := li.IndexOfName(v)
	if i != -1 {
		return li.ValueFromIndex(i)
	}
	return ""
}

func (li *Strings) Names(i int) string {
	k := strings.Index(li.Items[i], "=")
	fmt.Println(k)
	if k < 0 {
		return li.Items[i]
	} else {
		return string(li.Items[i][0:k])
	}
}

func (li *Strings) ValueFromIndex(i int) string {
	k := strings.Index(li.Items[i], "=")
	if k < 0 {
		return ""
	} else {
		return string(li.Items[i][k+1:])
	}
}
