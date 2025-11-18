package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/axcom/ninego/skit"
)

type DataSet struct {
	Error    error
	Fields   []*sql.ColumnType
	Records  []map[string]interface{}
	RecIndex int //recno -> -1=Bof, 0..N=recores, N+1=Eof

	recordCount int
	findKey     string
	findValue   []interface{}
}

//DataSet数据集
func New(records *sql.Rows) (ds *DataSet) {
	ds = &DataSet{}
	ds.Fields, _ = records.ColumnTypes()
	ds.RecIndex = -1
	ds.recordCount = 0
	if records != nil {
		var err error
		ds.Records, err = Rows2mapObjects(records)
		if err == nil {
			ds.recordCount = len(ds.Records)
		}
	} else {
		ds.Records = make([]map[string]interface{}, 0)
	}
	return
}

//用record为对象属性赋值（注：第1个为目标参数，须传入指针方能赋值成功）
func FetchFieldsFromRecord(oPtr interface{}, ds *DataSet) error {
	if reflect.ValueOf(oPtr).Kind() != reflect.Ptr {
		return fmt.Errorf("SetFieldValue Check value error not Ptr")
	}
	v := reflect.ValueOf(oPtr).Elem()
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("Check type error not Struct")
	}
	b := false
	for _, col := range ds.Fields {
		name := col.Name()
		if v, t, b = findField(v, t, name); b {
			val := ds.Value(name)
			if &val == nil {
				continue
			}
			sv := reflect.ValueOf(val)
			st := sv.Type()
			if st.Kind() == reflect.Ptr {
				sv = sv.Elem()
				st = sv.Type()
			}
			skit.AssignVal(v.FieldByName(name).Type(), st, v.FieldByName(name), sv)
		}
	}
	ds.RecIndex = -1
	return nil
}

//用于递归处理（字段名忽略大小写）
func findField(v reflect.Value, t reflect.Type, name string) (reflect.Value, reflect.Type, bool) {
	for i := 0; i < v.NumField(); i++ {
		if strings.EqualFold(t.Field(i).Name, name) {
			return v, t, true
		}
	}
	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).Type.Kind() == reflect.Struct && t.Field(i).Name == t.Field(i).Type.Name() {
			vv, tt, b := findField(v.Field(i), v.Field(i).Type(), name)
			if b {
				return vv, tt, b
			}
		}
		if t.Field(i).Type.Kind() == reflect.Ptr {
			if v.Field(i).IsNil() {
				continue
			}
			if v.Field(i).Elem().Type().Kind() == reflect.Struct && v.Field(i).Elem().Type().Name() == t.Field(i).Name {
				vv, tt, b := findField(v.Field(i).Elem(), v.Field(i).Elem().Type(), name)
				if b {
					return vv, tt, b
				}
			}
		}
	}
	return v, t, false
}

func Rows2mapObjects(rows *sql.Rows) ([]map[string]interface{}, error) {
	// 数据列
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	// 列的个数
	count := len(columns)

	// 返回值 Map切片
	mData := make([]map[string]interface{}, 0)
	// 一条数据的各列的值（需要指定长度为列的个数，以便获取地址）
	values := make([]interface{}, count)
	// 一条数据的各列的值的地址
	valPointers := make([]interface{}, count)
	for rows.Next() {

		// 获取各列的值的地址
		for i := 0; i < count; i++ {
			valPointers[i] = &values[i]
		}

		// 获取各列的值，放到对应的地址中
		rows.Scan(valPointers...)

		// 一条数据的Map (列名和值的键值对)
		entry := make(map[string]interface{})

		// Map 赋值
		for i, col := range columns {
			var v interface{}

			// 值复制给val(所以Scan时指定的地址可重复使用)
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				// 字符切片转为字符串
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}

		mData = append(mData, entry)
	}

	return mData, nil
}

//==============================================================================
// TDataSet
//==============================================================================
func (ds *DataSet) recno() int {
	if ds.RecIndex < 0 {
		if ds.recordCount == 0 {
			return -1
		}
		return 0
	} else if ds.RecIndex >= ds.recordCount {
		return ds.recordCount - 1
	}
	return ds.RecIndex
}

func (ds *DataSet) RecordCount() int {
	return ds.recordCount
}

func (ds *DataSet) Value(name string) interface{} {
	return ds.Records[ds.recno()][name]
}

func (ds *DataSet) ValueAsString(name string) string {
	if ds.Records[ds.recno()][name] == nil {
		return ""
	}
	return skit.String(ds.Records[ds.recno()][name])
}

func (ds *DataSet) ValueAsInteger(name string) int {
	if ds.Records[ds.recno()][name] == nil {
		return 0
	}
	var ret int
	skit.SetValue(&ret, ds.Records[ds.recno()][name])
	return ret
}

func (ds *DataSet) ValueAsFloat(name string) float64 {
	if ds.Records[ds.recno()][name] == nil {
		return 0
	}
	var ret float64
	skit.SetValue(&ret, ds.Records[ds.recno()][name])
	return ret
}

func (ds *DataSet) ValueAsBoolean(name string) bool {
	if ds.Records[ds.recno()][name] == nil {
		return false
	}
	var ret bool
	skit.SetValue(&ret, ds.Records[ds.recno()][name])
	return ret
}

func (ds *DataSet) ValueAsDateTime(name string) time.Time {
	if ds.Records[ds.recno()][name] == nil {
		return time.Time{}
	}
	var ret time.Time
	skit.SetValue(&ret, ds.Records[ds.recno()][name])
	return ret
}

func (ds *DataSet) IsEmpty() bool {
	return ds.Records == nil || ds.recordCount == 0
}

func (ds *DataSet) Row(i int) *DataSet {
	if i != ds.RecIndex {
		if i >= -1 && i <= ds.recordCount {
			ds.RecIndex = i
		}
	}
	return ds
}

func (ds *DataSet) Eof() bool {
	return ds.IsEmpty() || ds.RecIndex >= ds.recordCount
}

func (ds *DataSet) Bof() bool {
	return ds.IsEmpty() || ds.RecIndex < 0
}

func (ds *DataSet) First() {
	ds.RecIndex = -1
}

func (ds *DataSet) Last() {
	ds.RecIndex = ds.recordCount
}

func (ds *DataSet) Next() bool {
	ds.RecIndex++
	if ds.RecIndex >= ds.recordCount {
		if ds.RecIndex > ds.recordCount {
			ds.RecIndex--
		}
		return false
	}
	return true
}

func (ds *DataSet) Prior() bool {
	ds.RecIndex--
	if ds.RecIndex < 0 {
		ds.RecIndex++
		return false
	}
	return true
}

func (ds *DataSet) FieldCount() int {
	if ds.Records == nil || ds.recordCount == 0 {
		return 0
	}
	i := 0
	if ds.RecIndex >= 0 && ds.RecIndex < ds.recordCount {
		i = ds.RecIndex
	}
	return len(ds.Records[i])
}

//从首记录开始查找
func (ds *DataSet) Locate(KeyFields string, Values ...interface{}) bool {
	ds.findKey = KeyFields
	ds.findValue = Values
	return ds.find(0, KeyFields, Values...)
}

//从指定记录开始向下查找
func (ds *DataSet) find(recno int, KeyFields string, Values ...interface{}) bool {
	keys := strings.Split(KeyFields, ";")
	n := len(keys)
	if n > len(Values) {
		n = len(Values)
	}
	for i := recno; i < ds.recordCount; i++ {
		count := 0
		for k := 0; k < n; k++ {
			if ds.Row(i).ValueAsString(keys[k]) == skit.String(Values[k]) {
				count++
			} else {
				break
			}
		}
		if count == n {
			ds.RecIndex = i
			return true
		}
	}
	return false
}

//从当前记录向下查找（含当前记录）
func (ds *DataSet) Find(KeyFields string, Values ...interface{}) bool {
	ds.findKey = KeyFields
	ds.findValue = Values
	return ds.find(ds.RecIndex, KeyFields, Values...)
}

//向后查找
func (ds *DataSet) FindNext() bool {
	return ds.find(ds.RecIndex+1, ds.findKey, ds.findValue...)
}

//向前查找
func (ds *DataSet) FindPrior() bool {
	keys := strings.Split(ds.findKey, ";")
	n := len(keys)
	if n > len(ds.findValue) {
		n = len(ds.findValue)
	}
	for i := ds.RecIndex - 1; i >= 0; i-- {
		count := 0
		for k := 0; k < n; k++ {
			if skit.String(ds.Records[i][keys[k]]) == skit.String(ds.findValue[k]) {
				count++
			} else {
				break
			}
		}
		if count == n {
			ds.RecIndex = i
			return true
		}
	}
	return false
}
