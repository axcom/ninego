package class

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// IClass 核心接口：包含初始化方法（interfaces多个接口指针参数）
type IClass interface {
	_Init_(self, parent IClass, interfaces ...interface{})
}

// Object 业务基类（实现IClass）
type Object struct {
	self    IClass //让父对象知道自已真实存在于哪个类中
	methods map[string][]MethodInfo // 子类方法->父类方法
}

// MethodInfo 方法信息（包含反射值和接收者类型）
type MethodInfo struct {
	Value    reflect.Value // 方法反射值
	RecvType reflect.Type  // 方法接收者类型
}

// ------------------------------ 核心：Init方法（根据iface过滤方法） ------------------------------
func (o *Object) _Init_(self, parent IClass, interfaces ...interface{}) {
    // 在 Go 中，当你嵌入匿名结构体时：Object 字段会被自动创建（零值），但是它的内部状态（比如 self、methods）是零值，需要手动初始化。
	o.self = self
	o.methods = make(map[string][]MethodInfo)

	// 1. 解析所有接口（收集所有接口方法）
	methodNames := make(map[string]bool)
	for _, iface := range interfaces {
		if iface != nil {
			t := reflect.TypeOf(iface)
			if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Interface {
				panic("interfaces中的元素必须是接口指针，例如 (*IAdd)(nil) 或 new(IAdd)")
			}
			iType := t.Elem() // 提取接口类型

			// 检查self是否实现了该接口
			selfType := reflect.TypeOf(self)
			if !selfType.Implements(iType) {
				panic(fmt.Sprintf("%T 未实现接口 %v", self, iType))
			}
			
			// 收集当前接口中的所有方法名
			for i := 0; i < iType.NumMethod(); i++ {
				methodNames[iType.Method(i).Name] = true
			}
		} else if len(interfaces)>1 {
			panic("interfaces中的元素必须是接口指针，例如 (*IAdd)(nil) 或 new(IAdd)")
		}
	}

	// 2. 收集self的方法（当前实例）
	selfVal := reflect.ValueOf(o.self)
	if !selfVal.IsValid() {
		panic("self实例无效")
	}
	selfType := selfVal.Type()
	// 根据是否有接口过滤，决定收集方式
	if len(methodNames) > 0 {
		// 有接口限制，只收集指定方法
		for methodName := range methodNames {
			if m, ok := selfType.MethodByName(methodName); ok {
				methodVal := selfVal.MethodByName(methodName)
				if methodVal.IsValid() && methodVal.Kind() == reflect.Func {
					o.methods[methodName] = []MethodInfo{
						{
							Value:    methodVal,
							RecvType: m.Type.In(0),
						},
					}
				}
			}
		}
	} else {
		// 无接口限制，收集所有导出方法
		o.collectMethods(selfVal, selfType, nil)
	}


	// 3. 收集parent的方法（父类实例）
	if parent != nil {
		o.collectParentMethodChain(parent)
	}

	/*/ 打印methods方法链信息
	fmt.Printf("%T初始化完成：methods=%v\n", self, o.methods)
	for methodName, methodChain := range o.methods {
		fmt.Printf("methods.%s 方法链长度: %d\n", methodName, len(methodChain))
		for _,m:= range methodChain{ fmt.Printf("  %v\n",m)}
	}*/

}

// collectMethods 收集方法（根据iface过滤：iface=nil则收集所有导出方法，否则仅收集接口中声明的方法）
func (o *Object) collectMethods(val reflect.Value, typ reflect.Type, iType reflect.Type) {
	// 情况1：iface=nil → 收集所有导出方法
	if iType == nil {
		for i := 0; i < typ.NumMethod(); i++ {
			method := typ.Method(i)
			if method.Name=="Inherited"||method.Name=="Super" { continue }
			methodVal := val.MethodByName(method.Name)
			if methodVal.IsValid() && methodVal.Kind() == reflect.Func {
				o.methods[method.Name] = []MethodInfo{
					{
						Value:    methodVal,
						RecvType: method.Type.In(0),
					},
				}
			}
		}
		return
	}

	// 情况2：iface为接口 → 仅收集接口中声明的方法
	for i := 0; i < iType.NumMethod(); i++ {
		methodName := iType.Method(i).Name // 接口中声明的方法名
		// 查找实例中是否有该方法
		if m, ok := typ.MethodByName(methodName); ok {
			methodVal := val.MethodByName(methodName)
			if methodVal.IsValid() && methodVal.Kind() == reflect.Func {
				o.methods[methodName] = []MethodInfo{
					{
						Value:    methodVal,
						RecvType: m.Type.In(0),
					},
				}
			}
		}
	}
}

// collectParentMethodChain 收集父类方法链（一级级往前直到Object基类）
func (o *Object) collectParentMethodChain(parent IClass) {
	// 递归收集父类方法链
	currentParent := parent
	visitedTypes := make(map[reflect.Type]bool)
	
	// 从o.methods中提取方法名和方法类型，用于后续比较
	methodSignatures := make(map[string]reflect.Type)
	for methodName, mi := range o.methods {
		methodSignatures[methodName] = mi[0].Value.Type()
	}

	for currentParent != nil {
		parentVal := reflect.ValueOf(currentParent)
		parentType := parentVal.Type()
		
		// 避免循环引用导致的无限递归
		if visitedTypes[parentType] {
			break
		}
		visitedTypes[parentType] = true
		
		// 只收集o.methods中存在的父类的方法
		for methodName, expectedType := range methodSignatures {
			if m, ok := parentType.MethodByName(methodName); ok {
				methodVal := parentVal.MethodByName(methodName)
				if methodVal.IsValid() && methodVal.Kind() == reflect.Func {
					// 比较方法签名是否匹配
					actualType := methodVal.Type()
					if actualType == expectedType {
						methodInfo := MethodInfo{
							Value:    methodVal,
							RecvType: m.Type.In(0),
						}
						// 将方法添加到对应函数名的方法链中
						o.methods[methodName] = append(o.methods[methodName], methodInfo)
					} else {
						//fmt.Println(methodName,expectedType,"-->",parentType,actualType)
						// 父类有同名方法，但参数不一致
						delete(o.methods, methodName) //删掉这个方法
					}
				}
			}
		} 
		
		// 2. 查找父类（先tag后匿名）
		var parentField reflect.Value
		var found bool
		// 优先找tag父类
		parentField, found = findParentFieldByTag(parentVal)
		if !found {
			// 再找第一个匿名的基于Object的父类
			parentField, found = FindParentField(parentVal, reflect.TypeOf(Object{}))
			if !found {
				break
				//panic(fmt.Sprintf("结构体 %s 未找到父类（需通过 `class:\"parent\"` tag 或匿名嵌入基于Object的类）", structTypeName))
			}
		}
		if !parentField.CanSet() {
			break
			//panic(fmt.Sprintf("结构体 %s 的父类字段不可设置（非导出或无权限）", structTypeName))
		}
		if parentField.IsValid() && parentField.CanAddr() {
			currentParent = parentField.Addr().Interface().(IClass)
			continue
		}
		
		break
	}
}

// -------------------------- 父类查找逻辑（先tag后匿名） --------------------------
// findParentFieldByTag 查找带 `class:"parent"` tag 的字段
func findParentFieldByTag(structVal reflect.Value) (reflect.Value, bool) {
	if structVal.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	structType := structVal.Type()

	var parentField reflect.Value
	foundCount := 0

	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		if strings.TrimSpace(fieldType.Tag.Get("class")) == "parent" {
			parentField = structVal.Field(i)
			foundCount++
			if foundCount > 1 {
				panic(fmt.Sprintf("结构体 %s 存在多个 `class:\"parent\"` 字段（仅允许1个）", structType.Name()))
			}
		}
	}

	return parentField, foundCount == 1
}

// FindParentField 查找最外层匿名父类字段（用于New方法）
func FindParentField(structVal reflect.Value, parentType reflect.Type) (reflect.Value, bool) {
	var outermostField reflect.Value
	return findOutermostObjectField(structVal, parentType, true, &outermostField)
}

// findOutermostObjectField 递归查找结构体中所有匿名嵌入的Object字段
// 找到后返回最外层结构体中直接嵌入的那个Object` 字段（非嵌套的顶层字段）
// 参数：
//	structVal: 要查找的结构体（可能包含多层嵌套）
//	parentType: 要设置给找到的Object字段的值
//	isOuter: 标记是否为最外层调用（外部调用时传true）
//	outermostField: 用于记录最外层的匿名Object字段（仅内部递归用）
// 返回值：
//	最外层结构体中直接嵌入的继承于Object字段（reflect.Value）
//	是否找到并成功设置（bool）
func findOutermostObjectField(structVal reflect.Value, parentType reflect.Type, isOuter bool, outermostField *reflect.Value) (reflect.Value, bool) {
	// 确保处理的是结构体类型（解引用指针）
	val := structVal
	for val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	structType := val.Type()
	// 遍历当前结构体的字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := structType.Field(i)
		// 仅处理匿名字段
		if !fieldType.Anonymous {
			continue
		}

		// 解引用字段的指针类型
		fieldVal := field
		if fieldVal.Kind() == reflect.Ptr && !fieldVal.IsNil() {
			fieldVal = fieldVal.Elem()
		}

		// 情况1：当前字段就是目标父类类型
		if fieldVal.Type() == parentType {
			// 如果是外层调用，记录这个最外层字段
			if isOuter {
				*outermostField = field
			}
			return field, true
		}

		// 情况2：字段是结构体，递归查找（支持多层继承）
		if fieldVal.Kind() == reflect.Struct {
			// 递归查找，传递当前层级是否为外层
			foundField, ok := findOutermostObjectField(fieldVal, parentType, false, outermostField)
			if ok {
				// 如果是最外层调用且尚未记录最外层字段，记录当前字段
				if isOuter && !outermostField.IsValid() {
					*outermostField = field
				}
				// 如果是最外层调用，返回记录的最外层字段
				if isOuter && outermostField.IsValid() {
					return *outermostField, true
				}
				return foundField, true
			}
		}
	}
	return reflect.Value{}, false
}

// ------------------- New函数（支持tag/匿名父类和iface参数） ----------------------
// New[T] 泛型函数：调用形式 New[结构体](接口指针)
// 约束：T 必须是实现 IClass 的结构体指针（如 *MyObject）
// 入参：iface 是接口指针（如 (*IAdd)(nil)），用于指定要绑定的接口。（也可以为nil）
// 返回值：初始化后的结构体指针（已绑定接口方法，支持多态）
func New[T IClass](interfaces ...interface{}) T {
	var zero T
	tType := reflect.TypeOf(zero)

	// 1. 创建实例（确保为指针类型）
	var objVal reflect.Value
	if tType.Kind() == reflect.Ptr {
		objVal = reflect.New(tType.Elem())
	} else {
		objVal = reflect.New(tType)
	}
	structVal := objVal.Elem()
	structTypeName := structVal.Type().Name()

	// 2. 查找父类（先tag后匿名）
	var parentField reflect.Value
	var found bool
	// 优先找tag父类
	parentField, found = findParentFieldByTag(structVal)
	if !found {
		// 再找第一个匿名的基于Object的父类
		parentField, found = FindParentField(structVal, reflect.TypeOf(Object{}))
		if !found {
			panic(fmt.Sprintf("结构体 %s 未找到父类（需通过 `class:\"parent\"` tag 或匿名嵌入基于Object的类）", structTypeName))
		}
	}
	if !parentField.CanSet() {
		panic(fmt.Sprintf("结构体 %s 的父类字段不可设置（非导出或无权限）", structTypeName))
	}

	// 3. 初始化父类实例（处理指针/nil情况）
	var parentObj IClass
	switch parentField.Kind() {
	case reflect.Ptr:
		if parentField.IsNil() {
			parentInst := reflect.New(parentField.Type().Elem())
			parentField.Set(parentInst)
		}
		p, ok := parentField.Interface().(IClass)
		if !ok {
			panic(fmt.Sprintf("结构体 %s 的父类字段未实现 IClass 接口", structTypeName))
		}
		parentObj = p
	case reflect.Struct:
		p, ok := parentField.Addr().Interface().(IClass)
		if !ok {
			panic(fmt.Sprintf("结构体 %s 的父类字段（值类型）未实现 IClass 接口", structTypeName))
		}
		parentObj = p
	default:
		panic(fmt.Sprintf("结构体 %s 的父类字段类型不支持（仅支持结构体/结构体指针）", structTypeName))
	}

	// 4. 调用Init初始化（传递iface参数）
	obj := objVal.Interface().(T)
	obj._Init_(obj, parentObj, interfaces...)

	return obj
}

// 格式：Extends[结构体(子类)]("父类属性名",接口)
func Extends[T IClass](parentField string, interfaces ...interface{}) T {
	var zero T
	tType := reflect.TypeOf(zero)

	// 确保 objVal 始终是指针类型
	var objVal reflect.Value
	if tType.Kind() == reflect.Ptr {
		objVal = reflect.New(tType.Elem())
	} else {
		objVal = reflect.New(tType)
	}

	// 获取结构体值
	structVal := objVal.Elem() // 获取结构体值（如 MyObject）
	// 找到父类字段
	var parentObj IClass
	field := structVal.FieldByName(parentField)
	if field.IsValid() && field.CanSet() {
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				panic("Object 字段是 nil 指针")
			}
			// 指针类型：直接断言为 IClass（假设 Object 实现了 IClass）
			obj, ok := field.Interface().(IClass)
			if !ok {
				panic("Object 指针未实现 IClass 接口")
			}
			parentObj = obj
		} else {
			// 值类型：取地址后断言为 IClass（需要 Object 实现 IClass）
			obj, ok := field.Addr().Interface().(IClass)
			if !ok {
				panic("Object 值类型未实现 IClass 接口")
			}
			parentObj = obj
		}
	} else {
		panic(fmt.Sprintf("类型 %T 中没有父类的字段 %s", zero, parentField))
	}

	Obj := objVal.Interface()
	// 再 Init
	if c, ok := Obj.(IClass); ok {
		c._Init_(c, parentObj, interfaces...)
	}
	return Obj.(T)
}

// Create函数：直接使用传入的结构体，执行Init后返回
func Create[T IClass](obj T, args...interface{}) T {
	var parentName string = ""
	var interfaces []interface{}
	for _, arg := range args {
		if arg!=nil && reflect.TypeOf(arg).Kind()==reflect.String {
			parentName = arg.(string)
		} else {
			interfaces = append(interfaces, arg)
		}
	}
	
    // 1. 查找父类（复用new/extends原有逻辑，但直接操作传入的obj）
    objVal := reflect.ValueOf(obj).Elem()
    structTypeName := objVal.Type().Name()
    
	// 找到父类字段
    var parentField reflect.Value
    if parentName!=""{
		field := objVal.FieldByName(parentName)
		if field.IsValid() && field.CanSet() {
			parentField = field
		} else {
			panic(fmt.Sprintf("类型 %T 中没有父类的字段 %s", obj, parentName))
		}
    }
    // 查找父类字段（和原New函数逻辑一致）
    if !parentField.IsValid() {
	    var found bool
	    parentField, found = findParentFieldByTag(objVal)
	    if !found {
	        parentField, found = FindParentField(objVal, reflect.TypeOf(Object{}))
	        if !found {
	            panic(fmt.Sprintf("结构体 %s 未找到父类", structTypeName))
	        }
	    }
	    if !parentField.CanSet() {
	        panic(fmt.Sprintf("结构体 %s 的父类字段不可设置", structTypeName))
	    }
    }
    
    // 2. 初始化父类（确保父类已正确实例化）
    var parentObj IClass
    switch parentField.Kind() {
    case reflect.Ptr:
        if parentField.IsNil() {
            parentInst := reflect.New(parentField.Type().Elem())
            parentField.Set(parentInst)
        }
        p, ok := parentField.Interface().(IClass)
        if !ok {
            panic(fmt.Sprintf("结构体 %s 的父类未实现IClass", structTypeName))
        }
        parentObj = p
    case reflect.Struct:
        p, ok := parentField.Addr().Interface().(IClass)
        if !ok {
            panic(fmt.Sprintf("结构体 %s 的父类未实现IClass", structTypeName))
        }
        parentObj = p
    default:
        panic(fmt.Sprintf("结构体 %s 的父类类型不支持", structTypeName))
    }
    
    // 3. 执行Init方法完成初始化
    obj._Init_(obj, parentObj, interfaces...)
    return obj	
}	
// ------------------------------ 辅助函数及方法 ---------------------------------

// getMethodNames 提取方法名（调试用）
func getMethodNames(miMap map[string]MethodInfo) []string {
	names := make([]string, 0, len(miMap))
	for name := range miMap {
		names = append(names, name)
	}
	return names
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

// GetCallerMethodName 获取调用者方法名
func GetCallerMethodName(skip int) string {
	fullName := GetCallerFuncName(skip + 1)
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}

// 辅助函数：从方法名中提取类型名（如从"main.(*InheritedObject1).Add"中提取"InheritedObject1"; 从"main.*InheritedObject1"中提取InheritedObject1）
func getTypeNameFromFunc(fnName string) string {
	result := fnName
	
	// 处理格式如 "main.(*InheritedObject1).Add"
	if idx := strings.Index(result, "(*"); idx != -1 {
		start := idx + 2
		if end := strings.Index(result[start:], ")"); end != -1 {
			result = result[start : start+end]
		}
	}
	// 处理格式如 "main.InheritedObject1"
	if idx := strings.Index(result, "."); idx != -1 {
		parts := strings.Split(result[idx+1:], ".")
		result = parts[len(parts)-1]
	}	
	// 处理格式如 "main.*InheritedObject1"
	if idx := strings.Index(result, "*"); idx != -1 {
		result = result[idx+1:]
	}
	return result
}

// -------------------------- 继承类中使用的函数及方法 -----------------------------

func _Call_(method *reflect.Value, args ...interface{}) []reflect.Value {
	// 构造参数并调用方法
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}
	return method.Call(in)
}

// Inherited 调用顶层方法
// 动态调用方法（自动识别调用者，实现多态转发）
func (o *Object) Inherited(args ...interface{}) (result []reflect.Value, handled bool) {
	// 跳过 Inherited 和 GetCallerMethodName，获取真实调用者方法名
	fn := GetCallerMethodName(2)
	if fn == "call" {
		//ninego/class._call_ -4> reflect.Value.Call -3> reflect.Value.call -2> main.(*BaseObject).Dec 1
		fn = GetCallerMethodName(4) 
		if fn == "_Call_" {
		return nil, false
		}
	}
	// 查找方法信息
	fn = GetCallerMethodName(1)
	mi, ok := o.methods[fn]
	if !ok {
		return nil, false
	}
	// 避免调用自身（确保转发到子类实现）
	if reflect.TypeOf(o) == o.methods[fn][0].RecvType {
		return nil, false
	}

	// 构造参数并调用方法
	return _Call_(&mi[0].Value,args...), true
	/*in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}
	return mi.Value.Call(in), true*/
}

// Super 返回父类调用方法 - 可自动获取调用者函数名，支持obj.Super()(参数...)的调用方式
func (o *Object) Super(name ...string) func(...interface{}) ([]reflect.Value, bool) {
	// 获取调用者的函数名
    callerMethodName := GetCallerMethodName(1)
	if len(name)>0 && name[0]!="" {
  		callerMethodName = name[0]
	}
	
	// 返回一个闭包函数，接收任意参数
	return func(args ...interface{}) ([]reflect.Value, bool) {
		// 获取方法链
		methodChain, ok := o.methods[callerMethodName]
		if !ok || len(methodChain) == 0 {
			return nil, false
		}
		
		// 获取当前调用者的类型
		callerTypeName := getTypeNameFromFunc(GetCallerFuncName(1)) 
		// 在方法链中查找当前类的位置，然后调用下一个方法（父类方法）
		for i, methodInfo:= range methodChain {
			methodTypeName := getTypeNameFromFunc(methodInfo.RecvType.String())
			if methodTypeName!=callerTypeName { continue }
			if i<len(methodChain)-1 {
				// 准备参数并调用父方法
				return _Call_(&methodChain[i+1].Value, args...), true
				/*in := make([]reflect.Value, len(args))
				for i, arg := range args {
					in[i] = reflect.ValueOf(arg)
				}
				return methodChain[i+1].Value.Call(in), true*/
			}
		}
		
		// 如果找不到合适的父类方法，返回false
		return nil, false
	}
}


// SafeConvert 安全地将反射值转换为指定的泛型类型T
// 出错时返回T的零值，不返回错误
func SC[T any](val reflect.Value) T {
	var zero T
	if !val.IsValid() {
		return zero
	}
	
	defer func() {
		recover()
	}()
	
	return val.Interface().(T)
}