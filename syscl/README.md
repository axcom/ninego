### 常用class

#### List 为[]interface{}添加常用处理方法
- func (li *List[T]) Count() int 
- func (li *List[T]) Swap(i, j int) 
- func (li *List[T]) Sort(less func(a, b T) bool) 
- func (li *List[T]) Insert(index int, v T) 
- func (li *List[T]) Add(v T) 
- func (li *List[T]) Delete(index int) 
- func (li *List[T]) Remove(v T) 
- func (li *List[T]) Clear() 
- func (li *List[T]) IndexOf(v T) int 
- func (li *List[T]) Walk(do func(i int, v T)) 
- func (li *List[T]) Get(index int) T 
- func (li *List[T]) Set(index int, v T) 

#### Strings 为[]string添加常用处理方法
- func (li *Strings) Split(s string, sep ...string) int 
- func (li *Strings) Text(sep ...string) string 
- func (li *Strings) Strings(index int) string 
- func (li *Strings) Swap(i, j int) 
- func (li *Strings) Less(i, j int) bool 
- func (li *Strings) Insert(index int, s string, v ...string) 
- func (li *Strings) Add(s string, v ...string) 
- func (li *Strings) Delete(index int) 
- func (li *Strings) Clear() 
- func (li *Strings) Count() int 
- func (li *Strings) Walk(do func(i int, v string)) 
- func (li *Strings) IndexOf(v string) int 
- func (li *Strings) IndexOfName(v string) int 
- func (li *Strings) Names(i int) string 
- func (li *Strings) Values(v string) string 
- func (li *Strings) ValueFromIndex(i int) string 

#### StringList 为map[string]interface{}添加常用处理方法
- 同Strings
- func (this *StrMapList) ObjectFromIndex(index int) interface{} 
- func (this *StrMapList) Objects(key string) interface{} 

#### Buffer 为内嵌bytes.Buffer添加常用处理方法，支持连写

#### TreeNode 多叉树
