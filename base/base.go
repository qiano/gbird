package base

import (
	"errors"
	"fmt"
	"gbird/logger"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//Base 模型基础字段
type Base struct {
	Creater    string    //创建人
	CreateTime time.Time //创建时间
	Updater    string    //创建人
	UpdateTime time.Time //创建时间
	IsDelete   bool      //是否已删除
}

//Metadatas 模型元数据
var Metadatas map[string]map[string]FieldInfo

//FieldInfo 字段元数据
type FieldInfo struct {
	// Name string            //字段名
	Tags map[string]string //Tags
	Type string            //类型
	Kind string            //类型种类
}

func init() {
	Metadatas = make(map[string]map[string]FieldInfo)
}

//RegisterMetadata 将模型注册到源数据信息中
func RegisterMetadata(robj interface{}) {
	fields := make(map[string]FieldInfo)
	tags := []string{"bson", "collection", "urlname", "sole", "required", "default", "desc", "display", "ref"}
	refobj := reflect.ValueOf(robj).Elem()
	t := refobj.Type()
	for i := 0; i < refobj.NumField(); i++ {
		f := t.Field(i)
		field := new(FieldInfo)
		field.Type = f.Type.Name()
		field.Kind = f.Type.Kind().String()
		field.Tags = make(map[string]string)
		for _, tag := range tags {
			if v := f.Tag.Get(tag); v != "" {
				field.Tags[tag] = v
			}
		}
		fields[strings.ToLower(f.Name)] = *field
	}
	Metadatas[getKey(robj)] = fields
	logger.Infoln(getKey(robj) + " 模型元数据注册")
}

func getKey(robj interface{}) string {
	t := reflect.ValueOf(robj).Elem().Type()
	return strings.ToLower(t.PkgPath() + "/" + t.Name())
}

//Metadata 读取模型元数据
func Metadata(robj interface{}) (map[string]FieldInfo, error) {
	key := getKey(robj)
	v, ok := Metadatas[key]
	if !ok {
		return nil, errors.New("model:" + key + ",未读取到模型元数据")
	}
	return v, nil
}

//FieldMetadata 读取字段元数据
func FieldMetadata(robj interface{}, field string) (f FieldInfo, err error) {
	field = strings.ToLower(field)
	model, err := Metadata(robj)
	if err != nil {
		return f, err
	}
	if _, ok := model[field]; !ok {
		return f, errors.New("model:" + getKey(robj) + ",field:" + field + ",未读取到字段元数据")
	}
	return model[field], nil
}

//GetTag 读取TAG
func GetTag(robj interface{}, field, tag string) (string, error) {
	fieldmd, err := FieldMetadata(robj, field)
	if err != nil {
		return "", err
	}
	tag = strings.ToLower(tag)
	if _, ok := fieldmd.Tags[tag]; !ok {
		return "", errors.New("model:" + getKey(robj) + ",field:" + field + ",tag:" + tag + ",未读取到TAG数据")
	}
	return fieldmd.Tags[tag], nil
}

//GetTypeKind 读取字段的类型和Kind
func GetTypeKind(robj interface{}, field string) (string, string) {
	fieldmd, err := FieldMetadata(robj, field)
	fmt.Println(fieldmd)
	if err != nil {
		return "", ""
	}
	return fieldmd.Type, fieldmd.Kind
}

//FindTag 查找TAG值，或指定tag值的field
func FindTag(robj interface{}, tag, value string) (tagval string, field string, err error) {
	model, err := Metadata(robj)
	if err != nil {
		return
	}
	tag = strings.ToLower(tag)
	for key, val := range model {
		if v, ok := val.Tags[tag]; ok {
			if value == "" {
				return v, key, nil
			} else if value == v {
				return v, key, nil
			}
		}
	}
	return "", "", nil
}

//GetFieldsWithTag 获取有指定TAG的字段名
func GetFieldsWithTag(robj interface{}, tag string) (fs []string, err error) {
	model, err := Metadata(robj)
	if err != nil {
		return
	}
	tag = strings.ToLower(tag)
	for key, val := range model {
		if _, ok := val.Tags[tag]; ok {
			fs = append(fs, key)
		}
	}
	return
}

//ToSlice to slice
func ToSlice(arr interface{}) []interface{} {
	v := reflect.ValueOf(arr)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		panic("toslice arr not slice")
	}
	l := v.Len()
	ret := make([]interface{}, l)
	for i := 0; i < l; i++ {
		ret[i] = v.Index(i).Interface()
	}
	return ret
}

//SetValue 设置值
func SetValue(robj interface{}, field string, val interface{}) {
	refobj := reflect.ValueOf(robj).Elem()
	f := refobj.FieldByName(field)
	f.Set(reflect.ValueOf(val))
}

//GetValue 获取值
func GetValue(robj interface{}, field string) (string, interface{}) {
	refobj := reflect.ValueOf(robj).Elem()
	f := refobj.FieldByName(field)
	k := f.Kind()
	if k == reflect.Int || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64 || k == reflect.Int8 {
		v := f.Int()
		if v == 0 {
			return "", nil
		}
		return strconv.Itoa((int)(v)), nil
	} else if k == reflect.String {
		return f.String(), nil
	} else if k == reflect.Float64 || k == reflect.Float32 {
		return fmt.Sprintf("%.6f", f.Float()), nil
	} else if k == reflect.Struct {
		return "", f.Interface()
	}
	return "", nil
}
