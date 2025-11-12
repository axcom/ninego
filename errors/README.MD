# Go 错误处理扩展



## 功能

### 一、error增加Code信息

**1.** 实现了「错误码+错误信息+调用栈」三位一体的错误模型，解决了原生 error 无错误码、无栈信息的痛点；
**2.** 兼容 `github.com/pkg/errors` 生态，支持错误包装（Wrap）、错误链追踪（Cause）；`Wrap` 合并多消息（去重空消息），复用已有栈信息，避免栈冗余；支持递归打印根因错误栈。
**3.** 支持 Go 1.13+ 错误链标准（实现 `Unwrap` 方法），可配合包装的 `errors.Is`/`errors.As` 使用。
**4.** 提供了 `New`/`NewCode`/`Errorf`/`ErrorfCode` 等友好的错误构造函数。

### 二、TryCatch/Raise/ProtectRun

**1.** Try/TryCatch 模拟 try...catch...finally 语法。
场景：简化 panic 捕获逻辑，在 catch 中处理错误。

**2.** Raise 将错误（或任意值）包装调用栈后重新 panic。
场景：在 recover 后需要重新抛出错误时使用，保留完整调用栈。

**3.** ProtectRun 安全运行一个函数，捕获并返回所有panic错误。
场景：包装可能 panic 的函数，将 panic 转为可捕获的 error。



## 应用示例

### 一、error应用示例
#### 场景1：基础错误创建与使用
```go
package main

import (
	"fmt"
	"ninego/errors" 
)

func main() {
	// 1. 创建无错误码的错误
	err1 := errors.New("数据库连接超时")
	fmt.Println("err1:", err1)          // 输出：err1: error: 数据库连接超时
	fmt.Println("err1 Code:", errors.ErrorCode(err1)) // 输出：err1 Code: 

	// 2. 创建带错误码的错误
	err2 := errors.NewCode("DB001", "数据库连接超时")
	fmt.Println("err2:", err2)          // 输出：err2: code: DB001; error: 数据库连接超时
	fmt.Println("err2 Code:", errors.ErrorCode(err2)) // 输出：err2 Code: DB001

	// 3. 格式化创建错误
	err3 := errors.ErrorfCode("DB002", "数据库查询失败：%s", "表不存在")
	fmt.Println("err3:", err3) // 输出：err3: code: DB002; error: 数据库查询失败：表不存在

	// 4. 打印调用栈（%+v）
	fmt.Printf("\nerr3 详细信息：%+v\n", err3)
}
```
**说明**：基础构造函数覆盖「有无错误码」「是否格式化」场景，`ErrorCode` 可快速提取根因错误码，`%+v` 打印调用栈便于调试。


#### 场景2：错误包装与错误链追踪
```go
package main

import (
	"fmt"
	"ninego/errors"
)

// 模拟业务层函数
func queryDB(sql string) error {
	// 模拟底层错误
	return errors.NewCode("DB003", "执行SQL失败")
}

// 模拟服务层函数
func getUser(id int) error {
	sql := fmt.Sprintf("SELECT * FROM user WHERE id=%d", id)
	if err := queryDB(sql); err != nil {
		// 包装错误（添加业务上下文）
		return errors.Wrapf(err, "获取用户信息失败（id=%d）", id)
	}
	return nil
}

func main() {
	err := getUser(100)
	if err != nil {
		// 1. 打印完整错误链（含栈）
		fmt.Printf("错误详情：%+v\n", err)

		// 2. 获取根因错误
		rootErr := errors.Cause(err)
		fmt.Println("根因错误：", rootErr) // 输出：根因错误： code: DB003; error: 执行SQL失败

		// 3. 获取根因错误码
		fmt.Println("错误码：", errors.ErrorCode(err)) // 输出：错误码： DB003

		// 4. 错误类型断言
		var fault *errors.Fault
		if errors.As(err, &fault) {
			fmt.Println("断言成功，错误码：", fault.Code()) // 输出：断言成功，错误码： DB003
		}
	}
}
```
**说明**：`Wrap`/`Wrapf` 为错误添加业务上下文，不丢失原始错误信息；`Cause` 获取根因错误，`As` 实现错误类型断言，便于业务针对性处理。

### 二、TryCatch/Raise/ProtectRun应用示例

#### 示例1：使用`ProtectRun`捕获panic
```go
func riskyOperation() {
	panic("数据库连接失败")
}

func main() {
	err := ProtectRun(riskyOperation)
	if err != nil {
		fmt.Printf("捕获到错误：%+v\n", err) 
		// 输出包含调用栈：
		// 捕获到错误：数据库连接失败
		// 	/path/to/main.go:10 (riskyOperation)
		// 	/path/to/errors/errors.go:20 (ProtectRun)
	}
}
```
**说明**：`ProtectRun`将`panic`转为`error`并附加调用栈，适合在不希望程序崩溃的场景使用（如服务入口、任务调度）。


#### 示例2：使用`TryCatch`简化错误处理
```go
func main() {
	TryCatch(
		func() {
			// 可能panic的代码
			fmt.Println("执行操作...")
			panic(errors.New("操作失败"))
		},
		func(e interface{}) {
			// 处理错误
			fmt.Printf("捕获到错误：%v\n", e) // 输出：捕获到错误：操作失败
		},
	)
}
```
**说明**：`TryCatch`模拟了基本的异常捕获逻辑，适合简单场景下快速处理panic。


#### 示例3：使用`try/catch/finally`处理复杂错误
```go
// 定义自定义错误类型
type DBError struct {
	Code    string
	Message string
}

func (e *DBError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

type NetworkError struct {
	Reason string
}

func (e *NetworkError) Error() string {
	return "网络错误：" + e.Reason
}

func main() {
	Try(func() {
		// 模拟业务逻辑，可能抛出不同类型错误
		//panic(&DBError{Code: "DB001", Message: "连接超时"})
		//panic(&NetworkError{Reason: "超时"})
		panic("this is a error")
	}).Catch(&DBError{}, func(err error) {
		// 处理DBError类型错误
		dbErr := err.(*DBError)
		fmt.Printf("处理数据库错误：%s\n", dbErr.Code)
	}).Catch(&NetworkError{}, func(err error) {
		// 处理NetworkError类型错误
		fmt.Printf("处理网络错误：%v\n", err) // 输出：处理网络错误：网络错误：超时
	}).CatchAll(func(err error) {
		// 处理未匹配的错误
		fmt.Printf("处理未知错误：%v\n", err)
	}).Finally(func() {
		// 无论是否出错都会执行
		fmt.Println("操作结束，清理资源...") // 输出：操作结束，清理资源...
	})
}
```
**说明**：`try/catch/finally`适合复杂场景，可按错误类型分别处理，`finally`用于释放资源（如关闭文件、连接），逻辑更清晰。


#### 示例4：使用`Raise`重新抛出错误
```go
func process() {
	defer func() {
		if e := recover(); e != nil {
			// 处理部分逻辑后重新抛出错误
			fmt.Println("局部处理错误，准备向上传递...")
			Raise(e) // 包装调用栈后重新panic
		}
	}()
	panic("原始错误")
}

func main() {
	err := ProtectRun(process)
	fmt.Printf("最终捕获：%+v\n", err)
	// 输出：
	// 局部处理错误，准备向上传递...
	// 最终捕获：原始错误
	// 	/path/to/main.go:45 (process)
	// 	/path/to/errors/errors.go:50 (process.func1)
	// 	/path/to/errors/errors.go:80 (Raise)
}
```
**说明**：`Raise`在重新抛出错误时保留完整调用栈，便于追踪错误传播路径，适合多层级错误处理场景。
