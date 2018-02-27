package mongodb

import (
	"errors"
	"gbird/model"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strings"
	"time"
	"gbird/module/logger"
)

//GlobalMgoSession 全局mongo连接
var GlobalMgoSession *mgo.Session

//DbName 数据库名
var DbName string

//UseCol 使用Collection
func UseCol(colname string, f func(c *mgo.Collection)) {
	session := GlobalMgoSession.Clone()
	defer session.Close()
	col := session.DB(DbName).C(colname)
	f(col)

}

//Insert 新增
func Insert(robj interface{}, userid string) (err error) {
	if ok, err := ModelValidation(robj); !ok {
		logger.Fatalln(err)
		return err
	}
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	//引用关系处理
	fs, err := model.GetFieldsWithTag(robj, "ref")
	if err != nil {
		return
	}
	for _, f := range fs {
		_, val := model.GetValue(robj, f)
		ref := val.(mgo.DBRef)
		id := ref.Id.(string)
		c, er := model.FTagVal(robj, f, "ref")
		if er != nil {
			return er
		}
		ref.Id = bson.ObjectIdHex(id)
		ref.Collection = c
		ref.Database = DbName
		model.SetValue(robj, f, ref)
	}

	model.SetValue(robj, "ID", bson.NewObjectId())
	model.SetValue(robj, "Base", model.Base{
		Creater:    userid,
		CreateTime: time.Now(),
		Updater:    userid,
		UpdateTime: time.Now(),
	})
	UseCol(col, func(c *mgo.Collection) {
		err = c.Insert(&robj)
	})
	return
}

//Remove 删除
func Remove(robj interface{}, qjson string, userid string, batch bool) (info *mgo.ChangeInfo, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	UseCol(col, func(c *mgo.Collection) {
		info, err = Update(robj, qjson, `{"base.isdelete":"true"}`, userid, batch)
	})
	return
}

//Query 查询
func Query(robj interface{}, qjson string, page, pageSize int, sort string, fields string, containsDeleted bool) (datas interface{}, total int, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	qi, err := toQueryBson(robj, qjson, containsDeleted)
	if err != nil {
		return
	}
	fd, err := toBson(fields)
	if err != nil {
		return
	}
	if sort == "" {
		sort = "-updatetime -createtime"
	}
	refobj := reflect.ValueOf(robj).Elem()
	t := refobj.Type()
	slice := reflect.MakeSlice(reflect.SliceOf(t), 0, 0)
	temps := reflect.New(slice.Type())
	temps.Elem().Set(slice)
	UseCol(col, func(c *mgo.Collection) {
		qe := c.Find(qi).Sort(sort).Select(fd)
		if total, err = qe.Count(); err != nil {
			return
		}
		if page == 0 {
			qe.All(temps.Interface())
		} else {
			qe.Skip((page - 1) * pageSize).Limit(pageSize).All(temps.Interface())
		}
	})
	return temps.Interface(), total, nil
}

//FindID ID查找
func FindID(robj interface{}, id bson.ObjectId) (interface{}, error) {
	col, err := getCollection(robj)
	if err != nil {
		return nil, err
	}
	t := reflect.ValueOf(robj).Elem().Type()
	temp := reflect.New(t).Interface()
	UseCol(col, func(c *mgo.Collection) {
		err = c.FindId(id).One(temp)
	})
	return temp, err

}

//FindOne 查找一个
func FindOne(robj interface{}, qjson, sort string) (interface{}, error) {
	data, total, err := Query(robj, qjson, 1, 1, "", sort, false)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, errors.New("一个都没有")
	}
	arr := model.ToSlice(data)
	return arr[0], nil
}

//FindAll 查找所有
func FindAll(robj interface{}, qjson, sort string) ([]interface{}, error) {
	data, total, err := Query(robj, qjson, 0, 0, "", sort, false)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return make([]interface{}, 0, 0), nil
	}
	arr := model.ToSlice(data)
	return arr, nil
}

//Update 更新记录
func Update(robj interface{}, qjson, ujson string, userid string, batch bool) (info *mgo.ChangeInfo, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	q, err := toQueryBson(robj, qjson, false)
	if err != nil {
		return
	}
	u, err := toBson(ujson)
	if err != nil {
		return
	}
	err = UpdateValidation(robj, u)
	if err != nil {
		return
	}
	up := bson.M{"$set": u}
	temp := up["$set"].(bson.M)
	temp["base.updatetime"] = time.Now()
	temp["base.updater"] = userid
	UseCol(col, func(c *mgo.Collection) {

		if batch {
			if info, err = c.UpdateAll(q, up); err != nil {
			}
		} else {
			if err = c.Update(q, up); err != nil {
				info.Updated = 1
			}
		}
	})
	return
}

//UpdateID 更新
func UpdateID(robj interface{}, id bson.ObjectId, data bson.M, userid string) (err error) {
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	err = UpdateValidation(robj, data)
	if err != nil {
		return
	}
	d := bson.M{"$set": data}
	d = toLower(d)
	temp := d["$set"].(bson.M)
	temp["base.updatetime"] = time.Now()
	temp["base.updater"] = userid
	UseCol(col, func(c *mgo.Collection) {
		err = c.UpdateId(id, d)
	})
	return
}

//Count 计数
func Count(robj interface{}, qjson string, containsDeleted bool) (count int, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return 0, err
	}
	var b bson.M
	b, err = toQueryBson(robj, qjson, containsDeleted)
	if err != nil {
		return 0, err
	}
	UseCol(col, func(c *mgo.Collection) {
		q := c.Find(b)
		count, err = q.Count()
	})
	return
}

//DBRef 关联字段
func DBRef(robj interface{}) (ref *mgo.DBRef, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return nil, err
	}
	id, err := getMongoID(robj)
	if err != nil {
		return nil, err
	}
	return &mgo.DBRef{Collection: col, Id: id, Database: DbName}, nil
}

//FindRef 查找关联实体
func FindRef(robj interface{}, ref *mgo.DBRef) (interface{}, error) {
	session := GlobalMgoSession.Clone()
	defer session.Close()
	t := reflect.ValueOf(robj).Elem().Type()
	temp := reflect.New(t).Interface()
	err := session.DB(DbName).FindRef(ref).One(temp)
	return temp, err
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

func toBson(json string) (bson.M, error) {
	if len(json) == 0 {
		return nil, nil
	}
	var qi bson.M
	if err := bson.UnmarshalJSON([]byte(json), &qi); err != nil {
		return nil, errors.New("json=" + json + ",查询mongodb 查询 json错误 " + err.Error())
	}

	return toLower(qi), nil
}

func toQueryBson(robj interface{}, qjson string, containsDeleted bool) (bson.M, error) {
	qi, err := toBson(qjson)
	if err != nil {
		return qi, err
	}
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

func getMongoID(robj interface{}) (bson.ObjectId, error) {
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

//GetCollection 获取模型对应的集合
func getCollection(robj interface{}) (string, error) {
	return model.MTagVal(robj, "collection")
}
