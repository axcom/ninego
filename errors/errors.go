package errors

/*
实现了「错误码 + 错误信息 + 调用栈」三位一体的错误模型，解决了原生 error 无错误码、无栈信息的痛点；
兼容 github.com/pkg/errors 生态，支持错误包装（Wrap）、错误链追踪（Cause）；
支持 Go 1.13+ 错误链标准（实现 Unwrap 方法），可配合 errors.Is/errors.As 使用；
提供了 New/NewCode/Errorf 等友好的错误构造函数。
*/

import (
	"errors"
	"fmt"
	"io"
	"path"
	"runtime"

	"strconv"
	"strings"
)

// -------------------------- 错误模型定义 --------------------------
// Fault 包含错误码、错误信息和调用栈的错误模型
type Fault struct {
	code  string // 错误码（如"401""500"）
	msg   string // 错误描述信息
	stack *stack // 调用栈信息
	cause error  // 根因错误（支持错误链）
}

// Error 实现 error 接口
func (f *Fault) Error() string {
	if f.code == "" {
		return fmt.Sprintf("error: %v", f.msg)
	}
	return fmt.Sprintf("code: %s; error: %v", f.code, f.msg)
}

// Format 实现 fmt.Formatter 接口，支持 %+v 打印调用栈
func (f *Fault) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		io.WriteString(s, f.Error())
		if s.Flag('+') {
			if f.stack != nil {
				f.stack.Format(s, verb)
			}
			// 递归打印根因错误栈
			if f.cause != nil {
				fmt.Fprintf(s, "\ncause: %+v", f.cause)
			}
		}
	case 's':
		io.WriteString(s, f.msg)
	case 'q':
		fmt.Fprintf(s, "%q", f.msg)
	}
}

// Unwrap 实现 Go 1.13+ 错误链 Unwrap 接口
func (f *Fault) Unwrap() error {
	return f.cause
}

// Cause 实现 pkg/errors 的 causer 接口
func (f *Fault) Cause() error {
	return f.cause
}

// Code 返回错误码
func (f *Fault) Code() string {
	return f.code
}

// -------------------------- 构造函数 --------------------------
// New 创建无错误码的错误（包含调用栈）
func New(msg string) error {
	return &Fault{
		msg:   msg,
		stack: callers(1), // 偏移1：跳过当前 New 函数
	}
}

// NewCode 创建带错误码的错误（包含调用栈）
func NewCode(code, msg string) error {
	// 简单校验错误码合法性（可根据业务扩展）
	if code != "" && !isValidCode(code) {
		panic(fmt.Sprintf("invalid error code: %s (only digits or letters allowed)", code))
	}
	return &Fault{
		code:  code,
		msg:   msg,
		stack: callers(1),
	}
}

// Errorf 格式化创建错误（包含调用栈）
func Errorf(format string, args ...interface{}) error {
	return &Fault{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(1),
	}
}

// ErrorfCode 格式化创建带错误码的错误
func ErrorfCode(code, format string, args ...interface{}) error {
	if code != "" && !isValidCode(code) {
		panic(fmt.Sprintf("invalid error code: %s", code))
	}
	return &Fault{
		code:  code,
		msg:   fmt.Sprintf(format, args...),
		stack: callers(1),
	}
}

// -------------------------- 错误包装 --------------------------
// Wrap 包装错误，添加消息（已包含栈则复用，未包含则新增）
func Wrap(err error, msg ...string) error {
	if err == nil {
		return nil
	}
	// 合并消息（去重空消息）
	var message string
	for _, m := range msg {
		if m != "" {
			if message != "" {
				message += "; " + m
			} else {
				message = m
			}
		}
	}

	// 无论原错误是否为 Fault，都创建一个新的 Fault 实例
	// 这样可以保证每一次 Wrap 都能记录当前的调用栈
	newFault := &Fault{
		msg:   message,
		stack: callers(1), // 记录当前 Wrap 调用的位置
		cause: err,        // 将原始错误作为 cause
	}

	// 如果原始错误是 Fault 类型，则继承其错误码
	if fault, ok := err.(*Fault); ok {
		newFault.code = fault.code
	}

	return newFault
}

// Wrapf 格式化包装错误
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)

	//return Wrap(err, msg)

	// 无论原错误是否为 Fault，都创建一个新的 Fault 实例
	newFault := &Fault{
		msg:   msg,
		stack: callers(1), // 记录当前 Wrapf 调用的位置
		cause: err,        // 将原始错误作为 cause
	}

	// 如果原始错误是 Fault 类型，则继承其错误码
	if fault, ok := err.(*Fault); ok {
		newFault.code = fault.code
	}

	return newFault
}

// -------------------------- 工具函数 --------------------------
// Cause 获取错误链的根因错误
func Cause(err error) error {
	rootErr := err
	for rootErr != nil {
		if fault, ok := rootErr.(*Fault); ok {
			rootErr = fault.cause
			err = fault
			continue
		}
		err = rootErr
		break
	}
	return err
}

// ErrorCode 从错误链中获取根因错误的错误码
func ErrorCode(err error) string {
	if fault, ok := err.(*Fault); ok {
		return fault.code
	}
	/*rootErr := Cause(err)
	if fault, ok := rootErr.(*Fault); ok {
		return fault.code
	}*/
	return ""
}

// Is 包装 errors.Is，支持错误链匹配
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 包装 errors.As，支持错误链类型断言
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// -------------------------- 内部辅助函数 --------------------------
// isValidCode 简单校验错误码（仅字母/数字）
func isValidCode(code string) bool {
	for _, c := range code {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			return false
		}
	}
	return true
}

// -------------------------- 调用栈实现 --------------------------
// stack 调用栈（程序计数器切片）
type stack []uintptr

// callers 获取调用栈，skip 表示跳过的栈帧数（从当前函数开始计数）
func callers(skip int) *stack {
	const depth = 32 // 限制栈深度，避免性能问题
	var pcs [depth]uintptr
	// runtime.Callers(skip+2, ...)：+2 跳过 callers 自身和其调用者
	n := runtime.Callers(skip+2, pcs[:])
	st := stack(pcs[:n])
	return &st
}

// Format 格式化打印调用栈
func (s *stack) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		if st.Flag('+') {
			// 正序打印栈（从调用点到最底层）
			for _, pc := range *s {
				frame := Frame(pc)
				fmt.Fprintf(st, "\n%+v", frame)
			}
		}
	}
}

// Frame 单个栈帧
type Frame uintptr

// pc 返回程序计数器（修正+1偏移）
func (f Frame) pc() uintptr {
	return uintptr(f) - 1
}

// file 返回栈帧所在文件路径（相对路径，优化可读性）
func (f Frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	// 简化路径：保留 GOPATH 后的路径（若存在）
	/*if gopath := runtime.GOROOT(); strings.HasPrefix(file, gopath) {
		return strings.TrimPrefix(file, gopath+"/src/")
	}*/
	if runtime.GOROOT() != "" {
		return file
	}
	if runtime.GOOS == "windows" {
		if i := strings.Index(file, "\\src\\"); i >= 0 {
			return string(file[i+5:])
		}
	}
	if i := strings.Index(file, "/src/"); i >= 0 {
		return string(file[i+5:])
	}
	return path.Base(file) // 否则保留文件名
}

// line 返回栈帧所在行号
func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// name 返回函数名（简化包路径）
func (f Frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return funcname(fn.Name())
}

// Format 格式化栈帧
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			if runtime.GOROOT() != "" {
				io.WriteString(s, f.name())
				io.WriteString(s, "\n\t"+f.file())
			} else {
				// %+s：文件路径(函数名):行号
				io.WriteString(s, "\t"+f.file()+"("+f.name()+")")
			}
		default:
			// %s：仅文件名
			io.WriteString(s, path.Base(f.file()))
		}
	case 'd':
		// %d：行号
		io.WriteString(s, strconv.Itoa(f.line()))
	case 'n':
		// %n：函数名
		io.WriteString(s, f.name())
	case 'v':
		// %v：文件:行号
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// funcname 简化函数名（去除包路径前缀）
func funcname(name string) string {
	// 保留最后一个 . 后的函数名（如 "github.com/xxx/pkg.Func" → "pkg.Func"）
	if lastDot := strings.LastIndex(name, "."); lastDot != -1 {
		return name[lastDot+1:]
	}
	return name
}

// GetCallerFuncName 获取调用者函数名
func GetCallerFuncName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return "???"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "???"
	}
	return fn.Name()
}
