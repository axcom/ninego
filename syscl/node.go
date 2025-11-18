package syscl

type TreeNode struct {
	nodelist *List[*TreeNode]
	
	Parent   *TreeNode   `json:"-"`
	Data     interface{} `json:"-"`
}

func NewNode() (result *TreeNode) {
	result = &TreeNode{}
	result.Data = result
	//result.nodelist = classes.NewList() 留到使用子节点时再创建(直接使用Children()函数)
	return
}

/*
type TAppMenu struct {
	syscl.TreeNode `json:"-"`
	Text string
	Url  string
}
func (self *TAppMenu) Add() (result *TAppMenu) {
	result = NewAppMenu()
	self.TreeNode.AddChild(&result.TreeNode)
}
//创建根节点
func NewAppMenu() *TAppMenu {
	menu = &TAppMenu{TreeNode: *syscl.NewNode()}
	menu.Data = menu //创建根节点别忘了这句
	return &menu
}
// 递归遍历菜单树
func traverseMenu(menu *TAppMenu, depth int) {
	// 通过 This 字段直接访问菜单业务字段（类型安全，无断言）
	indent := strings.Repeat("  ", depth)
	fmt.Printf("%s菜单名称：%s，路径：%s\n", indent, menu.Text, menu.Url)

	// 遍历子菜单
	for _, childNode := range menu.Children().Items {
		traverseMenu(childNode.Data.(*TAppMenu), depth+1)
	}
}
*/

/*func (self *TreeNode) Add() (result *TreeNode) {
	//panic("没有覆写Add")

	result = NewNode()
	//result.Parent = self
	self.AddChild(result)
	return
}*/

func (self *TreeNode) Count() (result int) {
	if self.nodelist == nil {
		result = 0
	} else {
		result = self.nodelist.Count()
	}
	return
}

func (self *TreeNode) AddChild(aNode *TreeNode) {
	aNode.Parent = self
	self.Children().Add(aNode)
}

func (self *TreeNode) Remove(aNode *TreeNode) {
	self.Delete(self.nodelist.IndexOf(aNode))
}

func (self *TreeNode) Delete(Index int) {
	if (Index < self.Count()) && (Index > -1) {
		self.nodelist.Delete(Index)
	}
}

func (self *TreeNode) ClearAllNodes() {
	for I := self.Count() - 1; I >= 0; I-- {
		self.Delete(I)
	}
}

func (self *TreeNode) GetAllChildrenCount() (result int) {
	var processNode func(ANode *TreeNode) (AllCount int)
	processNode = func(ANode *TreeNode) (AllCount int) {
		AllCount = 1
		for I := 0; I <= ANode.Count()-1; I++ {
			AllCount += processNode(ANode.Index(I))
		}
		return
	} //process_recursion

	result = processNode(self)
	return
}

func (self *TreeNode) Index(Index int) (result *TreeNode) {
	if self.nodelist == nil {
		result = nil
	} else {
		result = self.nodelist.Items[Index]
	}
	return
}

func (self *TreeNode) Child(Index int) (result interface{}) {
	if self.nodelist == nil {
		result = nil
	} else {
		result = self.nodelist.Items[Index].Data
	}
	return
}

func (self *TreeNode) HasChild() (result bool) {
	result = (self.Count() > 0)
	return
}

func (self *TreeNode) GetRootNode() (result *TreeNode) {
	tmpResult := self
	if tmpResult.Parent == nil {
		result = self
		return
	}
	for {
		if tmpResult == nil {
			break
		}
		result = tmpResult

		tmpResult = tmpResult.Parent
	}
	return
}

/*func (self *TreeNode) Assign(Source *TreeNode) {
	self.ClearAllNodes()
	for i := 0; i <= Source.Count()-1; i++ {
		self.Add().Assign(Source.Index(i))
	}
}*/

func (self *TreeNode) GetRoot() (result *TreeNode) {
	root := self
	for {
		if root.Parent == nil {
			break
		}
		root = root.Parent
	}
	return root
}

type TNodeAttachMode int

const (
	NaAddChild      TNodeAttachMode = iota //移到节点下的子节点末
	NaAdd                                  //移到节点同级末
	NaAddFirst                             //移到节点同级首
	NaAddChildFirst                        //移到节点下的子节点首
	NaInsert                               //移到节点同级前
)

func (self *TreeNode) Children() (result *List[*TreeNode]) {
	if self.nodelist == nil {
		self.nodelist = NewList[*TreeNode]()
	}
	return self.nodelist
}

func (self *TreeNode) MoveTo(ToNode *TreeNode, Mode ...TNodeAttachMode) {
	mode := NaAddChild
	if len(Mode) > 0 {
		mode = Mode[0]
	}
	if self.Parent == nil {
		return
	}

	switch mode {
	case NaAddChild:
		{
			self.Parent.Remove(self)
			ToNode.Children().Add(self)
			self.Parent = ToNode
		}
	case NaAdd:
		{
			if ToNode.Parent == nil {
				break
			}
			self.Parent.Remove(self)
			ToNode.Parent.Children().Add(self)
			self.Parent = ToNode.Parent
		}
	case NaAddFirst:
		{
			if ToNode.Parent == nil {
				break
			}
			self.Parent.Remove(self)
			if ToNode.Parent.Count() == 0 {
				ToNode.Parent.Children().Add(self)
			} else {
				ToNode.Parent.Children().Insert(0, self)
			}
			self.Parent = ToNode.Parent
		}
	case NaAddChildFirst:
		{
			self.Parent.Remove(self)
			if ToNode.Count() == 0 {
				ToNode.Children().Add(self)
			} else {
				ToNode.Children().Insert(0, self)
			}
			self.Parent = ToNode
		}
	case NaInsert:
		{
			if ToNode.Parent == nil {
				break
			}
			self.Parent.Remove(self)
			ToNode.Parent.Children().Insert(ToNode.Parent.Children().IndexOf(ToNode), self)
			self.Parent = ToNode.Parent
		}
	}
}
