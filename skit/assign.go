package skit

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//自动类型转换赋值（注：第1个为目标参数，须传入指针方能赋值成功）
func SetValue(dPtr interface{}, s interface{}) {
	if reflect.ValueOf(dPtr).Kind() != reflect.Ptr {
		panic(fmt.Errorf("SetValue Check value error not Ptr"))
	}
	if s == nil {
		return
	}
	dv := reflect.ValueOf(dPtr).Elem()
	dt := dv.Type()
	sv := reflect.ValueOf(s)
	if reflect.ValueOf(s).Kind() == reflect.Ptr {
		sv = reflect.ValueOf(s).Elem()
	}
	st := sv.Type()
	if dt.Kind() == st.Kind() && dt.Name() == st.Name() {
		dv.Set(sv)
	} else {
		AssignVal(dt, st, dv, sv)
	}
}

//获取结构体指定名称Field字段的值（注：o结构体可以是指针也可是实参）
func GetField(o interface{}, name string) *Metadata {
	r := Metadata{}
	v := reflect.ValueOf(o)
	t := v.Type()
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		v = v.Elem()
		t = v.Type()
	} else if t.Kind() != reflect.Struct {
		//fmt.Errorf("GetField Check type error not Struct")
		return &r
	}
	b := false
	if v, t, b = findField(v, t, name); b {
		return r.Value(v.FieldByName(name).Interface())
	}
	//fmt.Errorf("GetField not find field")
	return &r
}

//将结构体Field字段的值赋给指定变量（注：o结构体可以是指针也可是实参, d目标变量必须是指针）
func GetFieldValue(o interface{}, name string, dPtr interface{}) error {
	v := reflect.ValueOf(o)
	t := v.Type()
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		v = v.Elem()
		t = v.Type()
	} else if t.Kind() != reflect.Struct {
		return fmt.Errorf("GetFieldValue Check type error not Struct")
	}
	b := false
	if v, t, b = findField(v, t, name); b {
		SetValue(dPtr, v.FieldByName(name).Interface())
		return nil
	}
	return fmt.Errorf("GetFieldValue not find field")
}

//设置结构体指定名称Field字段的值，可自动类型转换(注：结构体只能以指针传入，否则赋值失败)
func SetFieldValue(oPtr interface{}, name string, val interface{}) error {
	if reflect.ValueOf(oPtr).Kind() != reflect.Ptr {
		return fmt.Errorf("SetFieldValue Check value error not Ptr")
	}
	v := reflect.ValueOf(oPtr).Elem()
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("SetFieldValue Check type error not Struct")
	}
	b := false
	if v, t, b = findField(v, t, name); b {
		sv := reflect.ValueOf(val)
		st := sv.Type()
		if st.Kind() == reflect.Ptr {
			sv = sv.Elem()
			st = sv.Type()
		}
		AssignVal(v.FieldByName(name).Type(), st, v.FieldByName(name), sv)
		return nil
	}

	return fmt.Errorf("SetFieldValue not find field")
}

//同SetFieldValue，字段名字忽略大小写
func SetFieldVariant(oPtr interface{}, name string, val interface{}) error {
	if reflect.ValueOf(oPtr).Kind() != reflect.Ptr {
		return fmt.Errorf("SetFieldVariant Check value error not Ptr")
	}
	v := reflect.ValueOf(oPtr).Elem()
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("SetFieldVariant Check type error not Struct")
	}
	b := false
	if v, t, b = findField2(v, t, name); b {
		sv := reflect.ValueOf(val)
		st := sv.Type()
		if st.Kind() == reflect.Ptr {
			sv = sv.Elem()
			st = sv.Type()
		}
		AssignVal(v.FieldByName(name).Type(), st, v.FieldByName(name), sv)
		return nil
	}
	return fmt.Errorf("SetFieldVariant not find field")
}

//用于递归处理
func findField(v reflect.Value, t reflect.Type, name string) (reflect.Value, reflect.Type, bool) {
	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).Name == name {
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
		if t.Field(i).Type.Kind() == reflect.Ptr { //如果不是Interface or Pointer,调用Elem时就会panic: reflect: call of reflect.Value.Elem on struct Value
			if v.Field(i).IsNil() { //panic: reflect: call of reflect.Value.Type on zero Value
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

//用于递归处理（字段名忽略大小写）
func findField2(v reflect.Value, t reflect.Type, name string) (reflect.Value, reflect.Type, bool) {
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

//赋值
func AssignVal(dt, st reflect.Type, dv, sv reflect.Value) {
	switch dt.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch st.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dv.SetInt(sv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			dv.SetInt(int64(sv.Uint()))
		case reflect.Float32, reflect.Float64:
			dv.SetInt(int64(sv.Float()))
		case reflect.Bool:
			if sv.Bool() {
				dv.SetInt(1)
			} else {
				dv.SetInt(0)
			}
		case reflect.String:
			i, err := strconv.ParseInt(sv.String(), 10, 64)
			if err != nil {
				dv.SetInt(0)
			} else {
				dv.SetInt(i)
			}
		default:
			//dv.Set(sv)
			dv.SetInt(Int64(sv))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch st.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dv.SetUint(uint64(sv.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			dv.SetUint(uint64(sv.Uint()))
		case reflect.Float32, reflect.Float64:
			dv.SetUint(uint64(sv.Float()))
		case reflect.Bool:
			if sv.Bool() {
				dv.SetUint(1)
			} else {
				dv.SetUint(0)
			}
		case reflect.String:
			i, err := strconv.ParseUint(sv.String(), 10, 64)
			if err != nil {
				dv.SetUint(0)
			} else {
				dv.SetUint(i)
			}
		default:
			//dv.Set(sv)
			dv.SetUint(Uint64(sv))
		}
	case reflect.Float32, reflect.Float64:
		switch st.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dv.SetFloat(float64(sv.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			dv.SetFloat(float64(sv.Uint()))
		case reflect.Float32, reflect.Float64:
			dv.SetFloat(float64(sv.Float()))
		case reflect.Bool:
			if sv.Bool() {
				dv.SetFloat(1)
			} else {
				dv.SetFloat(0)
			}
		case reflect.String:
			i, err := strconv.ParseFloat(sv.String(), 64)
			if err != nil {
				dv.SetFloat(0)
			} else {
				dv.SetFloat(i)
			}
		default:
			switch sv.Interface().(type) {
			case []byte:
				dv.SetFloat(Float64(string(sv.Bytes())))
			default:
				//dv.Set(sv)
				dv.SetFloat(Float64(sv))
			}
		}
	case reflect.String:
		dv.SetString(String(sv))
	case reflect.Bool:
		switch st.Kind() {
		case reflect.Bool:
			dv.SetBool(sv.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if sv.Int() > 0 {
				dv.SetBool(true)
			} else {
				dv.SetBool(false)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if sv.Uint() > 0 {
				dv.SetBool(true)
			} else {
				dv.SetBool(false)
			}
		case reflect.Float32, reflect.Float64:
			if sv.Float() > 0 {
				dv.SetBool(true)
			} else {
				dv.SetBool(false)
			}
		default:
			s := String(sv)
			if strings.EqualFold(s, "true") || strings.EqualFold(s, "1") {
				dv.SetBool(true)
			} else {
				dv.SetBool(false)
			}
		}
	default:
		if dt.Name() == "Time" {
			var t time.Time
			if st.Name() == "DateTime" {
				if sv.Kind() == reflect.Ptr {
					t = time.Time(*sv.Interface().(*DateTime))
				} else {
					t = time.Time(sv.Interface().(DateTime))
				}
			} else {
				t = StrToDateTime(String(sv))
			}
			dv.Set(reflect.ValueOf(t))
		} else if dt.Name() == "DateTime" {
			var t DateTime
			if st.Name() == "Time" {
				if sv.Kind() == reflect.Ptr {
					t = DateTime(*sv.Interface().(*time.Time))
				} else {
					t = DateTime(sv.Interface().(time.Time))
				}
			} else {
				t = DateTime(StrToDateTime(String(sv)))
			}
			dv.Set(reflect.ValueOf(t))
			return
		} else {
			dv.Set(sv)
		}
	}
}
