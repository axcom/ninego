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
	//以上的每个继承方法都要写的代码
    //以下是本结构体的处理代码
    
    //调用父类代码
	if ret, ok := o.Super()(a,b); ok {
		//得到父类结果ret[]
	}
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
	1. 可使用接口指针来指定你的继承方法，如：`(*ICommon)(nil)` 或 `new(ICommon)` 。(若结构体不支持指定的接口，则会panic。)
    2. 在没有指定“接口指针”参数的情况下(传入为非指针nil)，若子类与父类有同名的方法，但参数不一致，那该方法/函数不能被继承使用。
    3. Create、New方法中省略"父结构体Field名称"参数，在创建该Class对象时，先通过 `class:"parent"` 这个tag标识查找父结构体，没找到的话还会再查找首个匿名嵌入的基于Object的结构体。(若没找到有父类，则会panic。)
    4. 每个结构体只能实现继承1个父类，未指定父结构名称时父类优先匹配第1个匿名嵌入的基于Object的类。子类结构体可以指定实现多个接口。
    5. 若将原来结构体单元的NewXXX(args...)函数返回修改成使用以上的创建对象方法，如结构体有嵌套的NewXXX引用，_init_可能会被执行多次（但最终结果还是没问题的）。
```

### 调用继承后的方法Inherited

在每个可继承的接口方法中：均要先执行**Inerited(...)**代码，以实现多态。
Inherited方法会去调用执行结构体子类中的最新方法。

```go
if ret, handled := obj.Inherited(参数...); handled {
    return class.SC[bool](ret[0]), class[error](ret[1])
}
```
​	该方法返回调用结果[]ret及handled标识，只有在handled为true的情况下，调用结果才不为nil。
​	如果参数是变参时，须用`class.Variadic{变参名称}`包裹变参变量。
​	返回的调用结果中，若有可能有为nil的结果值，需用class的**SafeConvert->安全转换函数**进行转换后使用: `class.SC[值类型](返回值[i]))`

**特别提醒：**
	**Inherited()调用请务必要放在接口方法/函数的第1行代码执行!!!（否则写在Inherited前面的代码可能会被执行2次!）**

### 调用父结构体的方法Super

在继承的接口方法中，可以使用**Super()(...)**调用当前父类的方法。

```go
if ret, ok := obj.Super("方法名称")(参数...); ok { ... }
```
	1."方法名称"参数可省略，为空即自动赋为当前函数名称。若"方法名称"在父结构体中不存在，ok返回false,ret=nil。
	2.该方法相当于调用“父结构名称.方法名称(...)”，且不会因为父方法中有Inherited而执行到子类代码。
	3.注意该方法返回的是个func，执行这个fcunc()后的结果处理同Inherited方法。若找不到合适的父类方法，ok返回false,ret=nil。

**特别提醒：**
	**Super()(...)调用请务必只写在继承后的struct的代码中执行!!!（在struct以外调用无效!）**

## 总结：
	** Inherited() 与 Super()() 方法只能在结构体内使用! 且调用者只能是当前结构体！ **
	** Inherited(...) 必须写在func内的第1行 **
	** Super() 返回的是父结构体的当前func方法，不要忘记传参并执行Super()(...) **
	** 调用时,变参入参用`class.Variadic{变参名称}`包裹。返回值用`class.SC[值类型](返回值[i]))`转换。 **
	** 好处：不用接口，也能实现继承调用。代价：每方法都要Inherited(),只能使用Super()调用上级方法**

### 使用示例：

```go
package main

import (
	"fmt"
	"ninego/class"
)

// ----------------------------- 接口定义 -----------------------------

// IEmployee 员工接口，定义员工的基本行为
type IEmployee interface {
	Work() string       // 员工工作内容
	GetSalary() float64 // 获取员工薪资
	GetName() string    // 获取员工姓名
}

// IManager 经理接口，定义经理的特殊行为
type IManager interface {
	Manage() string   // 经理管理职责
	GetTeamSize() int // 获取团队规模
}

// ----------------------------- 基类定义 -----------------------------

// Employee 员工基类，实现IEmployee接口
type Employee struct {
	class.Object         // 嵌入Object实现继承机制
	Name         string  // 员工姓名
	Salary       float64 // 员工薪资
}

func NewEmployee(salary float64) *Employee {
	return &Employee{Salary: salary}
}

// Work 实现IEmployee接口的Work方法
func (e *Employee) Work() string {
	// 先调用Inherited实现继承
	if ret, handled := e.Inherited(); handled {
		return class.SC[string](ret[0])
	}
	return fmt.Sprintf("%s is working", e.Name)
}

// GetSalary 实现IEmployee接口的GetSalary方法
func (e *Employee) GetSalary() float64 {
	// 先调用Inherited实现继承
	if ret, handled := e.Inherited(); handled {
		return class.SC[float64](ret[0])
	}
	return e.Salary
}

// GetName 实现IEmployee接口的GetName方法
func (e *Employee) GetName() string {
	// 先调用Inherited实现继承
	if ret, handled := e.Inherited(); handled {
		return class.SC[string](ret[0])
	}
	return e.Name
}

func (e *Employee) Manage() string {
	// 先调用Inherited实现继承
	if ret, handled := e.Inherited(); handled {
		return class.SC[string](ret[0])
	}
	return "base work"
}

// ----------------------------- 子类定义 -----------------------------

// Developer 开发人员类，继承自Employee，实现IEmployee接口
type Developer struct {
	Employee               // 嵌入Employee实现继承
	ProgrammingLang string // 编程语言
}

func NewDeveloper(salary float64) *Developer {
	return &Developer{Employee: *NewEmployee(salary)}
}

// Work 重写父类的Work方法，实现多态
func (d *Developer) Work() string {
	// 先调用Inherited实现多态
	if ret, handled := d.Inherited(); handled {
		return class.SC[string](ret[0])
	}
	return fmt.Sprintf("%s is coding in %s", d.Name, d.ProgrammingLang)
}

// GetSalary 重写父类的GetSalary方法，实现多态
func (d *Developer) GetSalary() float64 {
	// 先调用Inherited实现多态
	if ret, handled := d.Inherited(); handled {
		return class.SC[float64](ret[0])
	}
	// 开发人员薪资比基础薪资高20%
	baseSalary := d.Salary
	return baseSalary * 1.2
}

// ----------------------------- 经理类定义 -----------------------------

// Manager 经理类，继承自Employee，实现IEmployee和IManager接口
type Manager struct {
	Employee     // 嵌入Employee实现继承
	TeamSize int // 团队规模
}

// Work 重写父类的Work方法，实现多态
func (m *Manager) Work() string {
	// 先调用Inherited实现多态
	if ret, handled := m.Inherited(); handled {
		return class.SC[string](ret[0])
	}

	s := ""
	if ret, ok := m.Super()(); ok {
		s = class.SC[string](ret[0])
	}
	/*注意：
	s = m.Employee.Work() --> fatal error: stack overflow
	使用了 Inherited() 之后，就不能直接像上一样调用父类的方法了。只能通过 Supper()()来调用！
	否则会发生循环调用直至最后堆栈溢出！
	*/
	return s + " ++ " + fmt.Sprintf("%s is managing the team", m.Name)
}

// GetSalary 重写父类的GetSalary方法，实现多态
func (m *Manager) GetSalary() float64 {
	// 先调用Inherited实现多态
	if ret, handled := m.Inherited(); handled {
		return class.SC[float64](ret[0])
	}
	// 经理薪资比基础薪资高50%
	return m.Salary * 1.5
}

// Manage 实现IManager接口的Manage方法
func (m *Manager) Manage() string {
	// 先调用Inherited实现多态
	if ret, handled := m.Inherited(); handled {
		return class.SC[string](ret[0])
	}
	s := ""
	if ret, ok := m.Super()(); ok {
		s = class.SC[string](ret[0])
	}
	/*注意：
	s = m.Employee.Manage() --> fatal error: stack overflow
	使用了 Inherited() 之后，就不能直接像上一样调用父类的方法了。只能通过 Supper()()来调用！
	否则会发生循环调用直至最后堆栈溢出！
	*/
	return s + " && " + fmt.Sprintf("%s is managing a team of %d people", m.Name, m.TeamSize)
}

// GetTeamSize 实现IManager接口的GetTeamSize方法
func (m *Manager) GetTeamSize() int {
	// 先调用Inherited实现多态
	if ret, handled := m.Inherited(); handled {
		return class.SC[int](ret[0])
	}
	return m.TeamSize
}

// ----------------------------- 高级经理类定义 -----------------------------

// SeniorManager 高级经理类，继承自Manager，实现IEmployee和IManager接口
type SeniorManager struct {
	Manager           // 嵌入Manager实现继承
	Department string // 部门名称
}

// Work 重写父类的Work方法，实现多态
func (sm *SeniorManager) Work() string {
	// 先调用Inherited实现多态
	if ret, handled := sm.Inherited(); handled {
		return class.SC[string](ret[0])
	}
	return fmt.Sprintf("%s is managing the %s department", sm.Name, sm.Department)
}

// GetSalary 重写父类的GetSalary方法，实现多态
func (sm *SeniorManager) GetSalary() float64 {
	// 先调用Inherited实现多态
	if ret, handled := sm.Inherited(); handled {
		return class.SC[float64](ret[0])
	}
	// 高级经理薪资比经理薪资高30%
	if ret, ok := sm.Super("GetSalary")(); ok {
		managerSalary := class.SC[float64](ret[0])
		return managerSalary * 1.3
	}
	// 默认情况，使用基础薪资
	return sm.Salary * 1.95 // 1.5 * 1.3 = 1.95
}

// Manage 重写父类的Manage方法，实现多态
func (sm *SeniorManager) Manage() string {
	// 先调用Inherited实现多态
	if ret, handled := sm.Inherited(); handled {
		return class.SC[string](ret[0])
	}
	// 调用父类的Manage方法
	if ret, ok := sm.Super("Manage")(); ok {
		parentManage := class.SC[string](ret[0])
		return fmt.Sprintf("%s for the %s department", parentManage, sm.Department)
	}
	return fmt.Sprintf("%s is managing the %s department", sm.Name, sm.Department)
}

// ----------------------------- 示例函数 -----------------------------

// ExamplePolymorphism 演示继承功能
func ExamplePolymorphism() {
	fmt.Println("\n=== 继承示例 ===")

	// 创建不同类型的员工实例
	//employee := New[*Employee]((*IEmployee)(nil))
	employee := class.Create(NewEmployee(5000) /*,注意：没有限定接口方法*/)
	employee.Name = "John"
	//employee.Salary = 5000.0

	//developer := New[*Developer]((*IEmployee)(nil))
	developer := class.Create(NewDeveloper(5000) /*,注意：没有限定接口方法*/) 
	developer.Name = "Alice"
	//developer.Salary = 5000.0
	developer.ProgrammingLang = "Go"

	manager := class.New[*Manager]((*IEmployee)(nil))
	manager.Name = "Bob"
	manager.Salary = 8000.0

	seniorManager := class.New[*SeniorManager]((*IEmployee)(nil))
	seniorManager.Name = "Charlie"
	seniorManager.Salary = 8000.0
	seniorManager.Department = "Engineering"

	// 使用IEmployee接口类型存储不同的实现(多态)
	//employees := []IEmployee{employee, developer, manager, seniorManager}
	employees := []*Employee{employee, &developer.Employee, &manager.Employee, &seniorManager.Employee}

	// 多态：不通过接口调用不同实现的方法
	for _, emp := range employees {
		fmt.Printf("%s:\n", emp.GetName())
		fmt.Printf("  Work: %s\n", emp.Work())
		fmt.Printf("  Salary: %.2f\n", emp.GetSalary())
		fmt.Println()
	}

	// 没有绑定(*IManager)(nil)接口方法（参数限定只有(*IEmployee)(nil)接口方法）
	fmt.Printf("manager %s\n", manager.Manage())
	fmt.Printf("seniorManager %s\n", seniorManager.Manage())
}

// ExampleInterfaceBinding 演示接口绑定功能
func ExampleInterfaceBinding() {
	fmt.Println("\n=== 接口绑定示例 ===")

	// 创建一个只绑定IEmployee接口的Manager实例
	managerEmployee := class.New[*Manager]((*IEmployee)(nil))
	managerEmployee.Name = "Bob"
	managerEmployee.Salary = 8000.0
	managerEmployee.TeamSize = 5

	// 测试接口方法
	fmt.Println("Manager as IEmployee:")
	fmt.Println("  Work:", managerEmployee.Work())
	fmt.Println("  Salary:", managerEmployee.GetSalary())
	fmt.Println("  Name:", managerEmployee.GetName())

	// 创建一个同时绑定IEmployee和IManager接口的Manager实例
	managerFull := class.New[*Manager]((*IEmployee)(nil), (*IManager)(nil))
	managerFull.Name = "Bob"
	managerFull.Salary = 8000.0
	managerFull.TeamSize = 5

	fmt.Println("\nManager as IManager:")
	// 这里只能通过managerFull调用IManager接口方法
	fmt.Println("  Manage:", managerFull.Manage())
	fmt.Println("  Team Size:", managerFull.GetTeamSize())

	// 创建一个同时绑定IEmployee和IManager接口的SeniorManager实例
	s_managerFull := class.New[*SeniorManager]((*IEmployee)(nil), (*IManager)(nil))
	s_managerFull.Name = "Charlie"
	s_managerFull.Salary = 8000.0
	s_managerFull.TeamSize = 15

	fmt.Println("\nSeniorManager as IManager:")
	// 这里只能通过smanagerFull调用IManager接口方法
	fmt.Println("  Manage:", s_managerFull.Manage())
	fmt.Println("  Team Size:", s_managerFull.GetTeamSize())
}

func main() {
	// 运行多态示例
	ExamplePolymorphism()

	// 运行接口绑定示例
	ExampleInterfaceBinding()
}
```