package mongodb

import (
	"errors"
	"gbird/auth"
	"gbird/base"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
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
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	UseCol(col, func(c *mgo.Collection) {
		refobj := reflect.ValueOf(robj).Elem()
		typeOfT := refobj.Type()
		for i := 0; i < refobj.NumField(); i++ {
			bstr := typeOfT.Field(i).Tag.Get("bson")
			if bstr == "_id" {
				refobj.Field(i).Set(reflect.ValueOf(bson.NewObjectId()))
			}
			if v, ok := refobj.Field(i).Interface().(base.Base); ok {
				temp := time.Now()
				v.CreateTime = temp
				v.UpdateTime = temp
				v.Creater = user.ID
				v.Updater = user.ID
				refobj.Field(i).Set(reflect.ValueOf(v))
			}
		}
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
func Query(robj interface{}, qjson string, page, pageSize int, sort string, containsDeleted bool) (datas interface{}, total int, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	qi, err := toQueryBson(qjson, containsDeleted)
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
		qe := c.Find(qi).Sort(sort)
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

//Update 更新单条记录
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
	col := ""
	refobj := reflect.ValueOf(robj).Elem()
	t := refobj.Type()
	for i := 0; i < refobj.NumField(); i++ {
		col = t.Field(i).Tag.Get("collection")
		if col != "" {
			break
		}
	}
	if col == "" {
		return col, errors.New("model:" + t.String() + ",未设置TAG:collection")
	}
	return col, nil
}
