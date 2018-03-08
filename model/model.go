package model

import (
	"errors"
	"fmt"
	"gbird/logger"
	"gopkg.in/mgo.v2/bson"
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
	Tags map[string]string //Tags
	Type string            //类型
	Kind string            //类型种类
}

func init() {
	Metadatas = make(map[string]map[string]FieldInfo)
}

var mtags = []string{"collection", "router"}                                          //模型标签
var ftags = []string{"bson", "required", "default", "desc", "display", "ref", "enum"} //字段标签

//RegisterMetadata 将模型注册到源数据信息中
func RegisterMetadata(robj interface{}) {
	key := getKey(robj)
	_, ok := Metadatas[key]
	if ok {
		return
	}

	fields := make(map[string]FieldInfo)
	refobj := reflect.ValueOf(robj).Elem()
	t := refobj.Type()
	for i := 0; i < refobj.NumField(); i++ {
		f := t.Field(i)
		field := new(FieldInfo)
		field.Type = f.Type.Name()
		field.Kind = f.Type.Kind().String()
		field.Tags = make(map[string]string)
		for _, tag := range mtags {
			if v := f.Tag.Get(tag); v != "" {
				field.Tags[tag] = v
			}
		}
		for _, tag := range ftags {
			if v := f.Tag.Get(tag); v != "" {
				field.Tags[tag] = v
			}
		}
		fields[strings.ToLower(f.Name)] = *field
	}
	Metadatas[key] = fields
	logger.Infoln(key + " 模型元数据注册")
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
		RegisterMetadata(robj)
		return Metadata(robj)
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

//FTagVal 字段TAG值
func FTagVal(robj interface{}, field, tag string) (string, error) {
	for _, val := range ftags {
		if tag == val {
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
	}
	return "", errors.New("该方法只支持模型标签：" + strings.Join(ftags, ","))
}

//MTagVal 获取模型标签的值，只支持标签：collection, router
func MTagVal(robj interface{}, tag string) (string, error) {
	for _, val := range mtags {
		if tag == val {
			model, err := Metadata(robj)
			if err != nil {
				return "", err
			}
			tag = strings.ToLower(tag)
			for _, val := range model {
				if v, ok := val.Tags[tag]; ok {
					if v != "" {
						return v, nil
					}
				}
			}
			return "", errors.New("model:" + getKey(robj) + ",未设置标签:" + tag)
		}
	}
	return "", errors.New("该方法只支持模型标签：" + strings.Join(mtags, ","))
}

//GetTypeKind 读取字段的类型和Kind
func GetTypeKind(robj interface{}, field string) (string, string) {
	fieldmd, err := FieldMetadata(robj, field)
	if err != nil {
		return "", ""
	}
	return fieldmd.Type, fieldmd.Kind
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

//GetID 获取ID
func GetID(robj interface{}) (bson.ObjectId, error) {
	refobj := reflect.ValueOf(robj).Elem()
	typeOfT := refobj.Type()
	for i := 0; i < refobj.NumField(); i++ {
		bstr := typeOfT.Field(i).Tag.Get("bson")
		if bstr == "_id" {
			v := refobj.Field(i).String()
			if len(v) == 0 {
				return "", errors.New("模型：" + typeOfT.String() + ",对象 bson: _id 值为空")
			}
			return (bson.ObjectId)(v), nil
		}
	}
	return "", errors.New("模型：" + typeOfT.String() + ",未设置TAG： bson: _id 设置")
}

//GetEnum  获取枚举值
func GetEnum(robj interface{}, field string) (map[string]string, error) {
	val, err := FTagVal(robj, field, "enum")
	if err != nil {
		return nil, err
	}
	rets := make(map[string]string)
	kvs := strings.Split(val, ",")
	for j := 0; j < len(kvs); j++ {
		ss := strings.Split(kvs[j], ":")
		rets[ss[0]] = ss[1]
	}
	return rets, nil
}

//GetEnumDesc 枚举值描述
func GetEnumDesc(robj interface{}, fieldname, code string) string {
	enums, err := GetEnum(robj, fieldname)
	if err != nil {
		panic(err)
	}
	return enums[code]
}
