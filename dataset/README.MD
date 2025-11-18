### 数据集(DataSet)

将sql.Rows record数据转换为DataSet数据集struct(相当于**"内存表"**)。

//DataSet数据集

func New(records *sql.Rows) (ds *DataSet) 


//用record为对象属性赋值（注：第1个为目标参数，须传入指针方能赋值成功）

func FetchFieldsFromRecord(oPtr interface{}, ds *DataSet) error 

#### Dataset方法函数

//前当记录指针
field: RecIndex int

//获取字段值 
- func (ds *DataSet) Value(name string) interface{} 
- func (ds *DataSet) ValueAsString(name string) string 
- func (ds *DataSet) ValueAsInteger(name string) int 
- func (ds *DataSet) ValueAsFloat(name string) float64 
- func (ds *DataSet) ValueAsBoolean(name string) bool 
- func (ds *DataSet) ValueAsDateTime(name string) time.Time 


//数据集是否为空

func (ds *DataSet) IsEmpty() bool 

//表字段个数

func (ds *DataSet) FieldCount() int 

//表记录条数

func (ds *DataSet) RecordCount() int

//记录指针移到第i条记录(-1=Bof, 0...n=records, n+1=Eof)

func (ds *DataSet) Row(i int) *DataSet 


//记录指针是否已越过末条记录

func (ds *DataSet) Eof() bool 

//记录指针是否已越过首条记录

func (ds *DataSet) Bof() bool 


//记录指针移到第1条记录

func (ds *DataSet) First() 

//记录指针移到最末条记录

func (ds *DataSet) Last() 


//记录指针移到一条记录

func (ds *DataSet) Next() bool 

//记录指针往前移一条记录

func (ds *DataSet) Prior() bool 


//从首记录开始查找

func (ds *DataSet) Locate(KeyFields string, Values ...interface{}) bool 

//从当前记录向下查找（含当前记录）

func (ds *DataSet) Find(KeyFields string, Values ...interface{}) bool 

//向后查找

func (ds *DataSet) FindNext() bool 

//向前查找

func (ds *DataSet) FindPrior() bool 

