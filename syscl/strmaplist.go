package syscl

import (
	"fmt"
	"strings"
	"sync"
)

//Map模拟delphi的TStringList，支持Name(key)\Values\Object

type StringList struct {
	Delimiter string

	lock     sync.RWMutex
	dataMap  map[string]*element
	dataList []*element
}

func NewStrMapList(sls ...string) *StringList {
	if len(sls) > 0 && sls[0] != "" {
		return &StringList{dataMap: make(map[string]*element), Delimiter: sls[0]}
	} else {
		return &StringList{dataMap: make(map[string]*element), Delimiter: DefaultDelimiter /*","*/}
	}
}

func SplitToMapList(s, sep string) *StringList {
	m := NewStrMapList(sep)
	m.Split(s, sep)
	return m
}

type element struct {
	Key    string
	Value  string
	Object interface{}
}

func (this *StringList) Split(s string, sep ...string) int {
	this.lock.Lock()
	if len(sep) > 0 {
		this.Delimiter = sep[0]
	}
	v := strings.Split(s, this.Delimiter)
	for i := 0; i < len(v); i++ {
		t := strings.Split(v[i], "=")
		if len(t) > 1 {
			this.Add(t[0], t[1])
		} else {
			this.Add(v[i])
		}
	}
	this.lock.Unlock()
	return len(this.dataMap)
}

func (this *StringList) Text(sep ...string) (ret string) {
	s := this.Delimiter
	if len(sep) > 0 {
		s = sep[0]
	}
	for i := 0; i < this.Count(); i++ {
		if i == 0 {
			ret = this.Strings(i)
		} else {
			ret += s + this.Strings(i)
		}
	}
	return
}

func (this *StringList) String() (ret string) {
	return this.Text()
}

func (this *StringList) Exists(key string) bool {
	_, ok := this.dataMap[key]
	return ok
}

func (this *StringList) Count() int {
	return len(this.dataMap)
}

func (this *StringList) Clear() {
	this.lock.Lock()
	//重建Map
	this.dataMap = make(map[string]*element)
	this.dataList = append([]*element{})
	this.lock.Unlock()
}

func (this *StringList) IndexOf(key string) int {
	for i := 0; i < this.Count(); i++ {
		if this.dataMap[key] == this.dataList[i] {
			return i
		}
	}
	return -1
}

//key,value,object 或 key,object(非string)
func (this *StringList) Add(key string, value ...interface{}) int {
	var k, v string
	var o interface{}
	this.lock.Lock()
	defer this.lock.Unlock()
	if len(value) >= 2 {
		k = key
		//v = value[0].(string)
		switch value[0].(type) {
		case string:
			v = value[0].(string)
		default:
			v = fmt.Sprintf("%v", value[0])
		}
		o = value[1]
	} else {
		if strings.Contains(key, "=") {
			kv := strings.Split(key, "=")
			k = kv[0]
			v = kv[1]
			if len(value) > 0 {
				o = value[0]
			}
		} else {
			k = key
			if len(value) > 0 {
				switch value[0].(type) {
				case string:
					v = value[0].(string)
				default:
					o = value[0]
				}
			}
		}
	}
	i := this.IndexOf(k)
	if i != -1 {
		this.Lines(i).Value = v
		this.Lines(i).Object = o
		return i
	}
	keyer := &element{Key: k, Value: v, Object: o}
	this.dataMap[keyer.Key] = keyer
	this.dataList = append(this.dataList, keyer)
	return this.Count() - 1
}

func (this *StringList) Insert(index int, key string, value ...interface{}) int {
	var k, v string
	var o interface{}
	this.lock.Lock()
	defer this.lock.Unlock()
	if len(value) >= 2 {
		k = key
		//v = value[0].(string)
		switch value[0].(type) {
		case string:
			v = value[0].(string)
		default:
			v = fmt.Sprintf("%v", value[0])
		}
		o = value[1]
	} else {
		if strings.Contains(key, "=") {
			kv := strings.Split(key, "=")
			k = kv[0]
			v = kv[1]
			if len(value) > 0 {
				o = value[0]
			}
		} else {
			k = key
			if len(value) > 0 {
				switch value[0].(type) {
				case string:
					v = value[0].(string)
				default:
					o = value[0]
				}
			}
		}
	}
	i := this.IndexOf(k)
	if i != -1 {
		this.Lines(i).Value = v
		this.Lines(i).Object = o
		return i
	}
	keyer := &element{Key: k, Value: v, Object: o}
	this.dataMap[keyer.Key] = keyer
	//this.dataList = append(this.dataList, keyer)
	rear := append([]*element{}, (this.dataList)[index:]...)
	this.dataList = append(append((this.dataList)[:index], keyer), rear...)
	return this.Count() - 1
}

func (this *StringList) Delete(index int) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if index > -1 && index < this.Count() {
		delete(this.dataMap, this.dataList[index].Key)
		this.dataList = append(this.dataList[:index], this.dataList[index+1:]...)
	}
}

func (this *StringList) Remove(key string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	index := this.IndexOf(key)
	if index > -1 {
		delete(this.dataMap, key)
		this.dataList = append(this.dataList[:index], this.dataList[index+1:]...)
	}
}

func (this *StringList) Strings(index int) string {
	if this.dataList[index].Value == "" {
		return this.dataList[index].Key
	}
	return this.dataList[index].Key + "=" + this.dataList[index].Value
}

func (this *StringList) Lines(index int) *element {
	return this.dataList[index]
}

func (this *StringList) Walk(do func(i int, k string, v string, o interface{})) {
	for i, v := range this.dataList {
		do(i, v.Key, v.Value, v.Object)
	}
}

func (this *StringList) Names(index int) string {
	return this.dataList[index].Key
}

func (this *StringList) ValueFromIndex(index int) string {
	return this.dataList[index].Value
}

func (this *StringList) ObjectFromIndex(index int) interface{} {
	return this.dataList[index].Object
}

func (this *StringList) Values(key string) string {
	return this.dataMap[key].Value
}

func (this *StringList) Objects(key string) interface{} {
	return this.dataMap[key].Object
}
