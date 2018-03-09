package mongodb

import (
	"errors"
	"gbird/model"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strings"
)

//ToBson 字符串转Bson.M
func ToBson(json string) (bson.M, error) {
	if len(json) == 0 {
		return nil, nil
	}
	var qi bson.M
	if err := bson.UnmarshalJSON([]byte(json), &qi); err != nil {
		return nil, errors.New("json=" + json + ",查询mongodb 查询 json错误 " + err.Error())
	}

	return toLower(qi), nil
}
func toLower(q map[string]interface{}) bson.M {
	ret := make(bson.M)
	for key, val := range q {
		if reflect.TypeOf(val) == reflect.TypeOf(q) && val != nil {
			val = toLower(val.(map[string]interface{}))
		} else if reflect.TypeOf(val).Kind() == reflect.Slice {
			arr := make([]map[string]interface{}, 0, 0)
			temp := val.([]interface{})
			for i := 0; i < len(temp); i++ {
				v := temp[i].(map[string]interface{})
				v = toLower(v)
				arr = append(arr, v)
			}
			val = arr

		}
		ret[strings.ToLower(key)] = val
	}
	return ret
}

//ToQueryBson 为查询条件附加默认查询配置
func toQueryBson(robj interface{}, qi bson.M, containsDeleted bool) (bson.M, error) {
	//objectid处理
	for key, val := range qi {
		ty, kind := model.GetTypeKind(robj, key)
		if ty == "ObjectId" && kind == "string" {
			qi[key] = bson.ObjectIdHex(val.(string))
		}
	}

	if !containsDeleted {
		if qi == nil {
			qi = bson.M{"base.isdelete": false}
		} else {
			if _, ok := qi["base.isdelete"]; !ok {
				qi["base.isdelete"] = false
			}
		}
	}
	return qi, nil
}
