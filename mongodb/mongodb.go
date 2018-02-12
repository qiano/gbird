package mongodb

import (
	"errors"
	"fmt"
	"gbird/base"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strconv"
	"strings"
	"time"
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
		return err
	}
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	//引用关系处理
	fs, err := base.GetFieldsWithTag(robj, "ref")
	if err != nil {
		return
	}
	for _, f := range fs {
		_, val := base.GetValue(robj, f)
		ref := val.(mgo.DBRef)
		id := ref.Id.(string)
		c, er := base.GetTag(robj, f, "ref")
		if er != nil {
			return er
		}
		ref.Id = bson.ObjectIdHex(id)
		ref.Collection = c
		ref.Database = DbName
		base.SetValue(robj, f, ref)
	}

	base.SetValue(robj, "ID", bson.NewObjectId())
	base.SetValue(robj, "Base", base.Base{
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
	arr := base.ToSlice(data)
	return arr[0], nil
}

//FindAll 查找所有
func FindAll(robj interface{}, qjson, sort string) ([]interface{}, error) {
	data, total, err := Query(robj, qjson, 0, 0, "", sort, false)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, errors.New("一个都没有")
	}
	arr := base.ToSlice(data)
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
	d := bson.M{"$set": data}
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
		fmt.Println(key, val, reflect.TypeOf(val), reflect.TypeOf(val).Kind())
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
	fmt.Println(ret)
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
		ty, kind := base.GetTypeKind(robj, key)
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
	tval, _, err := base.FindTag(robj, "collection", "")
	return tval, err
}

//ModelValidation 模型验证
func ModelValidation(robj interface{}) (bool, error) {
	//结构验证
	refobj := reflect.ValueOf(robj).Elem()
	t := refobj.Type()
	for i := 0; i < refobj.NumField(); i++ {
		def := t.Field(i).Tag.Get("default")
		if def != "" {
			if t.Field(i).Type.Kind() == reflect.Int && refobj.Field(i).Int() == 0 {
				val, err := strconv.Atoi(def)
				if err != nil {
					return false, errors.New("model:" + t.String() + ",字段" + t.Field(i).Name + "，TAG:default,默认值设置异常")
				}
				refobj.Field(i).SetInt((int64)(val))
			}
			if t.Field(i).Type.Kind() == reflect.String && refobj.Field(i).String() == "" {
				refobj.Field(i).SetString(def)
			}
		}
		req := t.Field(i).Tag.Get("required")
		if req == "true" {
			val, o := base.GetValue(robj, t.Field(i).Name)
			if val == "" && o == nil {
				return false, errors.New("model:" + t.String() + ",字段" + t.Field(i).Name + "，TAG:required,值为空")
			}
		}
	}

	//唯一性验证
	soles, _, err := base.FindTag(robj, "sole", "")
	if err != nil {
		panic(err)
	}
	if len(soles) > 0 {
		exists := []string{}
		groups := strings.Split(soles, "|")
		for _, g := range groups {
			temps := []string{}
			for _, val := range strings.Split(g, ",") {
				v, _ := base.GetValue(robj, val)
				if v != "" {
					temps = append(temps, `"`+strings.ToLower(val)+`":"`+v+`"`)
				}
			}
			if len(temps) > 0 {
				exists = append(exists, `{`+strings.Join(temps, ",")+`}`)
			}
		}
		qjson := exists[0]
		if len(exists) > 1 {
			qjson = `{"$or":[` + strings.Join(exists, ",") + `]}`
		}
		if count, err := Count(robj, qjson, false); err == nil && count > 0 {
			return false, errors.New("数据已存���，查询条件：" + qjson)
		}
	}
	return true, nil
}
