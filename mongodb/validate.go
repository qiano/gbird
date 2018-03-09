package mongodb

import (
	"errors"
	"gbird/logger"
	"gbird/model"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strconv"
	"strings"
)

//ModelValidation 模型验证
func ModelValidation(robj interface{}) (bool, error) {
	//值验证
	refobj := reflect.ValueOf(robj).Elem()
	t := refobj.Type()
	for i := 0; i < refobj.NumField(); i++ {
		err := FieldValidation(robj, t.Field(i).Name)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

//FieldValidation 字段验证
func FieldValidation(robj interface{}, fieldname string) error {
	//值验证
	refobj := reflect.ValueOf(robj).Elem()
	t := refobj.Type()
	field, _ := t.FieldByName(fieldname)
	def := field.Tag.Get("default")
	if def != "" {
		if field.Type.Kind() == reflect.Int && refobj.FieldByName(fieldname).Int() == 0 {
			val, err := strconv.Atoi(def)
			if err != nil {
				return errors.New("model:" + t.String() + ",字段" + fieldname + "，TAG:default,默认值设置异常")
			}
			refobj.FieldByName(fieldname).SetInt((int64)(val))
		}
		if field.Type.Kind() == reflect.String && refobj.FieldByName(fieldname).String() == "" {
			refobj.FieldByName(fieldname).SetString(def)
		}
	}
	req := field.Tag.Get("required")
	if req == "true" {
		val, o := model.GetValue(robj, fieldname)
		if val == "" && o == nil {
			return errors.New("model:" + t.String() + ",字段" + fieldname + "，TAG:required,值为空")
		}
	}
	//枚举值验证
	enums, err := model.GetEnum(robj, fieldname)
	if err != nil {
		return nil
	}
	v, _ := model.GetValue(robj, fieldname)
	_, ok := enums[v]
	if !ok {
		return errors.New("model:" + t.String() + ",字段" + fieldname + "，TAG:enums,非法值")
	}
	return nil
}

//SoleValidation 唯一性验证
func SoleValidation(robj interface{}) (bool, error) {
	tval, err := model.MTagVal(robj, "sole")
	if err != nil {
		logger.Fatalln(err)
	}
	if tval == "" {
		return true, nil
	}
	qbson := bson.M{}
	for _, val := range strings.Split(tval, " ") {
		v, _ := model.GetValue(robj, val)
		if v != "" {
			qbson[val] = v
		}
	}
	if count, err := Count(robj, qbson, false); err == nil && count > 0 {
		return false, errors.New("唯一性验证结果：数据已存在，查询字段：" + tval)
	}
	return true, nil
}

//UpdateValidation  模型更新值验证
func UpdateValidation(robj interface{}, u map[string]interface{}) (err error) {
	for key, val := range u {

		if reflect.TypeOf(val) == reflect.TypeOf(u) && val != nil {
			err = UpdateValidation(robj, val.(map[string]interface{}))
			if err != nil {
				return
			}
		} else if reflect.TypeOf(val).Kind() == reflect.Slice {
			temp := model.ToSlice(val)
			for i := 0; i < len(temp); i++ {
				if reflect.TypeOf(temp[i]).Kind() == reflect.Struct {
					return
				}
				v := temp[i].(map[string]interface{})
				err = UpdateValidation(robj, v)
				if err != nil {
					return
				}
			}
		}
		err = FieldValidation(robj, key)
		if err != nil {
			return
		}
	}
	return nil
}
