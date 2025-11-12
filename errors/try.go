package errors

/*
使用ProtectRun捕获调用func的 panic
Raise 将错误（或任意值）包装调用栈后重新 panic
使用TryCatch简化错误处理
使用Try/catch/finally处理复杂错误
*/

import (
	"fmt"
	"reflect"
)

// ProtectRun 安全运行一个函数，捕获并返回所有panic错误
// 场景：包装可能 panic 的函数，将 panic 转为可捕获的 error
func ProtectRun(entry func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			// 统一将 panic 内容转为 error 类型
			switch v := e.(type) {
			case error:
				err = v.(error)
			case string:
				err = fmt.Errorf(v)
			default:
				err = fmt.Errorf("%v", v)
			}
		}
	}()

	entry()
	return
}

// Raise 将错误（或任意值）包装调用栈后重新 panic
// 场景：在 recover 后需要重新抛出错误时使用，保留完整调用栈
func Raise(e interface{}) {
	if e == nil {
		return
	}

	var err error
	switch v := e.(type) {
	case error:
		err = v
	case string:
		err = fmt.Errorf(v)
	default:
		err = fmt.Errorf("%v", v)
	}

	// 确保错误包含调用栈
	err = withStackIfNeeded(err, 2) // 跳过 Raise 自身调用栈
	panic(err)
}

// TryCatch 模拟 try...catch 语法
// 场景：简化 panic 捕获逻辑，在 catch 中处理错误
func TryCatch(exec func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			handler(err) // 直接将 panic 内容传递给处理器
		}
	}()
	exec()
}

// ------------------------------ try/catch/finally 实现 ------------------------------

// tryCatch 用于模拟 try...catch...finally 语法的结构体
type tryCatch struct {
	err      interface{}                      // 捕获到的错误（非channel实现，避免并发问题）
	catches  map[reflect.Type]func(err error) // 按类型存储的catch处理器
	catchAll func(err error)                  // 兜底处理器
}

// Try 启动一个 try 块，执行可能 panic 的代码
// 注意：内部不使用goroutine，避免并发导致的逻辑混乱
func Try(block func()) (t *tryCatch) {
	t = &tryCatch{
		catches:  make(map[reflect.Type]func(err error)),
		catchAll: func(err error) {}, // 默认空实现，避免nil调用
	}

	// 直接在当前goroutine执行，确保代码执行顺序可控
	defer func() {
		// 捕获当前块中的panic
		if r := recover(); r != nil {
			t.err = r
		}
	}()
	// 执行可能发生panic的代码块
	block()

	return t
}

// Catch 注册指定类型错误的处理器（支持接口类型匹配）
// 参数e：错误类型示例（如 &MyError{}），用于匹配错误类型
func (t *tryCatch) Catch(e error, block func(err error)) *tryCatch {
	// 防护1：如果e是nil，直接返回（避免后续reflect操作空指针）
	if e == nil {
		fmt.Println("warning: Catch() received nil error instance, skip registration")
		return t
	}

	// 获取错误的动态类型
	errType := reflect.TypeOf(e)
	// 防护2：双重确认errType不为nil（极端情况防护）
	if errType == nil {
		fmt.Println("warning: failed to get type of error instance, skip registration")
		return t
	}

	// 处理指针类型：如果是指针，同时注册其指向的类型（兼容非指针错误）
	if errType.Kind() == reflect.Ptr {
		// 此时errType是指针，Elem()安全
		t.catches[errType.Elem()] = block
	}
	// 注册原类型（指针/非指针）
	t.catches[errType] = block
	return t
}

// CatchAll 注册未被匹配的错误的兜底处理器
func (t *tryCatch) CatchAll(block func(err error)) *tryCatch {
	t.catchAll = block
	return t
}

// Finally 注册最终执行的代码块（无论是否发生错误都会执行）
// 注意：调用Finally后才会实际处理捕获的错误
func (t *tryCatch) Finally(block func()) {
	defer block() // 确保finally在最后执行

	if t.err == nil {
		return // 无错误，直接执行finally
	}

	// 将错误转为error类型
	var err error
	switch v := t.err.(type) {
	case error:
		err = v
	default:
		err = fmt.Errorf("%v", v)
	}

	// 查找匹配的catch处理器
	errType := reflect.TypeOf(err)
	if handler, ok := t.catches[errType]; ok {
		handler(err)
		return
	}

	// 如果错误是一个指针，也尝试匹配其指向的类型（处理用户传入非指针实例的情况）
	if errType.Kind() == reflect.Ptr {
		if handler, ok := t.catches[errType.Elem()]; ok {
			handler(err)
			return
		}
	}

	// 未匹配到任何处理器，执行兜底
	t.catchAll(err)
}

// ------------------------------ 内部辅助函数 ------------------------------

// withStackIfNeeded 为非stackError的错误添加调用栈
func withStackIfNeeded(err error, skip int) error {
	if err == nil {
		return nil
	}

	// 创建一个新的 Fault 实例, 记录当前的调用栈
	newFault := &Fault{
		msg:   funcname(GetCallerFuncName(skip-1)) + " -> " + fmt.Sprintf("%v", err),
		stack: callers(skip), // 记录当前 Raise 调用的位置
		cause: err,           // 将原始错误作为 cause
	}

	// 如果原始错误是 Fault 类型，则继承其错误码
	if fault, ok := err.(*Fault); ok {
		newFault.code = fault.code
	}

	return newFault
}
