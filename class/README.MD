## Go模拟实现java/delphi等语言的class--类继承及多态

### 背景
​	在将java\python\delphi等语言的代码移植到go项目中时，经常因为类继承的问题，不得不重新设计原来的对象。
​	本项目通过Inherited(...)、Super()(...)这2个方法，让go也实现了类的继承和多态调用。
​	
典型代码示例：

```go
func(o *MyObject)Add(a,b int) int,error {
    //跳转调用子类代码(若有子类代码则handled=true,跳转到子类代码执行后退出)
    if ret, handled := o.Inherited(a, b1); handled {
        return class.SC[int](ret[0]), class.SC[error](ret[1])
	}
    //以下是本结构体的处理代码
    
    //调用父类代码
	ret, ok := o.Super()(a,b)
	// 本结构体的扩展代码
	......
}
```

### 创建Class对象：

```go
	//创建一个新的Class对象，可指定实现的接口指针
	New[*结构体]((*接口指针)(nil)...)
	//创建一个新的Class对象，并指定父类在本结构体的名字及实现的接口指针
	Extends[*结构体]("父结构体Field名称", (*接口指针)(nil)...)
	//将一个结构体创建(转化)为Class对象，可指定父类及实现的接口指针
	Create(&结构体{}, "父结构体Field名称"/(*接口指针)(nil)...)
```
​    所有想要支持继承的结构体，都必须要在**基类结构体**中引入Object对象。Object对象实现了Inherited(...)及Super()(...)方法。
​    以上创建结构体对象的3种方式，都会自动执行Object的`_init_`方法：将结构体转换为类对象。在创建类对象时，其结构体内的父类结构体会被自动创建。

**注意：**

```
	1. 请尽量使用接口指针来指定你的继承方法，如：`(*ICommon)(nil)` 或 `new(ICommon)` 。(若结构体不支持指定的接口，则会panic。)
    2. 在没有指定“接口指针”参数的情况下(传入为非指针nil)，若子类与父类有同名的方法，但参数不一致，那该方法/函数不能被继承使用。
    3. Create、New方法中省略"父结构体Field名称"参数，在创建该Class对象时，先通过 `class:"parent"` tag标识查找父结构体，没找到的话还会再查找首个匿名嵌入的基于Object的结构体。(若没找到有父类，则会panic。)
    4. 每个结构体只能实现继承1个父类，未指定父结构名称时父类优先匹配第1个匿名嵌入的基于Object的类。子类结构体可以指定实现多个接口。
    5. 若将原来结构体单元的NewXXX(args...)函数返回修改成使用以上的创建对象方法，如结构体有嵌套的NewXXX引用，_init_可能会被执行多次（但最终结果还是没问题的）。
```

### 调用继承后的方法

在每个可继承的接口方法中：均要先执行**Inerited(...)**代码，以实现多态。
Inherited方法会去调用执行结构体子类中的最新方法。

```go
if ret, handled := obj.Inherited(参数...); handled {
    return class.SC[bool](ret[0]), class[error](ret[1])
}
```
​	该方法返回调用结果[]ret及handled标识，只有在handled为true的情况下，调用结果才不为nil。
	如果参数是变参时，须用`class.Variadic{变参名称}`包裹变参变量。
​	返回的调用结果中，若有可能有为nil的结果值，需用class的**SafeConvert->安全转换函数**进行转换后使用: `class.SC[值类型](返回值[i]))`

**特别提醒：**
	**Inherited()调用请务必要放在接口方法/函数的第1行代码执行!!!（否则写在Inherited前面的代码可能会被执行2次!）**

### 调用父结构体的方法

在继承的接口方法中，可以使用**Super()(...)**调用当前父类的方法。

```go
if ret, ok := obj.Super("方法名称")(参数...); ok { ... }
```
	1."方法名称"参数可省略，为空即自动赋为当前函数名称。若"方法名称"在父结构体中不存在，ok返回false,ret=nil。
	2.该方法相当于调用“父结构名称.方法名称(...)”，且不会因为父方法中有Inherited而执行到子类代码。
	3.注意该方法返回的是个func，执行这个fcunc()后的结果处理同Inherited方法。若找不到合适的父类方法，ok返回false,ret=nil。

**注意：**
	**Inherited() 与 Super()() 方法只能在结构体内使用! 且调用者只能是当前结构体！**

### 使用示例：

```go
// ICommon 业务接口
type ICommon interface {
    Add(a, b int) (int, error)
	Dec()
}

// BaseObject 业务基类
type BaseObject struct {
	Object
}

// ICommon 接口实现（转发版本）
func (b *BaseObject) Add(a, b1 int) int,error {
	ret, handled := b.Inherited(a, b1)
	if handled {
        return class.SC[int](ret[0]), class.SC[error](ret[1])
	}
	fmt.Printf("[%T] 执行 BaseAdd(%d, %d)\n", b, a, b1)
	b.Dec() // 这里可实现多态!!!
	return a + b1
}

func (b *BaseObject) Dec() {
	_, handled := b.Inherited()
	if handled {
		return
	}

	fmt.Printf("[%T] 执行 Dec()\n", b)
}

// 子类1
type InheritedObject1 struct {
	BaseObject
}

/*func (i *InheritedObject1) Add(a, bi int) int {
	fmt.Printf("[%T] 执行 Add(%d, %d)\n", i, a, bi)
	return a + bi
}*/

func (i *InheritedObject1) Dec() {
	fmt.Printf("[%T] 执行 1-Dec()\n", i)
    i.Super()() // 调用父类方法
}

// 子类2
type InheritedObject2 struct {
	InheritedObject1
    Field1 int
}

func (i *InheritedObject2) Dec() {
	fmt.Printf("[%T] 执行 2-Dec 开始\n", i)
	// 直接调用父类方法
    i.Super("Dec")()
	// 当前方法
	fmt.Printf("[%T] 执行 2-Dec 结束\n", i)
}

func main() {
	base := Extends[*BaseObject]("Object", (*ICommon)(nil))
	child1 := Extends[*InheritedObject1]("BaseObject", (*ICommon)(nil))
	child2 := New[*InheritedObject2](nil)

	objects := []IAdd{base, child1, child2}
	for _, obj := range objects {
		result := obj.Add(3, 4)
		fmt.Println("返回值:", result)
		obj.Dec()
		fmt.Println("---")
	}

    obj := Create(&InheritedObject2{Field1: 100})
	fmt.Println("obj.Add(5,6)=", obj.Add(5, 6))
	obj.Dec()
	fmt.Println("---")
	obj.Super("Dec")()
	fmt.Println("+++")
	obj.InheritedObject1.Dec()
}
```