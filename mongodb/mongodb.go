package mongodb

import (
	"errors"
	// "fmt"
	"gbird/auth"
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
func Insert(robj interface{}, user auth.User) (err error) {
	if ok, err := ModelValidation(robj); !ok {
		return err
	}
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	base.SetValue(robj, "ID", bson.NewObjectId())
	base.SetValue(robj, "Base", base.Base{
		Creater:    user.ID,
		CreateTime: time.Now(),
		Updater:    user.ID,
		UpdateTime: time.Now(),
	})
	UseCol(col, func(c *mgo.Collection) {
		err = c.Insert(&robj)
	})
	return
}

//Remove 删除
func Remove(robj interface{}, qjson string, user auth.User, batch bool) (info *mgo.ChangeInfo, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	UseCol(col, func(c *mgo.Collection) {
		info, err = Update(robj, qjson, `{"base.isdelete":"true"}`, user, batch)
	})
	return
}

//Query 查询
func Query(robj interface{}, qjson string, page, pageSize int, sort string, fields string, containsDeleted bool) (datas interface{}, total int, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	qi, err := toQueryBson(qjson, containsDeleted)
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
func FindID(robj interface{}, id string) (interface{}, error) {
	col, err := getCollection(robj)
	if err != nil {
		return nil, err
	}
	t := reflect.ValueOf(robj).Type()
	temp := reflect.New(t).Interface()
	UseCol(col, func(c *mgo.Collection) {
		err = c.FindId(bson.ObjectIdHex(id)).One(temp)
	})
	return temp, err

}

//Update 更新记录
func Update(robj interface{}, qjson, ujson string, user auth.User, batch bool) (info *mgo.ChangeInfo, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	q, err := toQueryBson(qjson, false)
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
	temp["base.updater"] = user.ID
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

//Count 计数
func Count(robj interface{}, qjson string, containsDeleted bool) (count int, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return 0, err
	}
	var b bson.M
	b, err = toQueryBson(qjson, containsDeleted)
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

func toBson(json string) (bson.M, error) {
	if len(json) == 0 {
		return nil, nil
	}
	var qi bson.M
	if err := bson.UnmarshalJSON([]byte(json), &qi); err != nil {
		return nil, errors.New("json=" + json + ",查询mongodb 查询 json错误 " + err.Error())
	}
	return qi, nil
}

func toQueryBson(qjson string, containsDeleted bool) (bson.M, error) {
	qi, err := toBson(qjson)
	if err != nil {
		return qi, err
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
			if t.Field(i).Type.Kind() == reflect.Int && refobj.Field(i).Int() == 0 {
				return false, errors.New("model:" + t.String() + ",字段" + t.Field(i).Name + "，TAG:required,值为空")
			} else if t.Field(i).Type.Kind() == reflect.String && refobj.Field(i).String() == "" {
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
				field := reflect.ValueOf(robj).Elem().FieldByName(val)
				v := field.Interface().(string)
				temps = append(temps, `"`+strings.ToLower(val)+`":"`+v+`"`)
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
			return false, errors.New("数据已存在，查询条件：" + qjson)
		}
	}
	return true, nil
}
