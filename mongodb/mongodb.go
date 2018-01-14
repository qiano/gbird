package mongodb

import (
	"errors"
	"fmt"
	"gbird/base"
	"gbird/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"time"
)

//GlobalMgoSession 全局mongo连接
var GlobalMgoSession *mgo.Session
var mongodbstr = config.Config["mongodbHost"]

//DbName 数据库名
var DbName = config.Config["mongodbDbName"]

//Connect 连接Mongo数据库
func init() {
	globalMgoSession, err := mgo.DialWithTimeout(mongodbstr, 10*time.Second)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("连接成功："+mongodbstr, DbName)
	GlobalMgoSession = globalMgoSession
	GlobalMgoSession.SetMode(mgo.Monotonic, true)
	//default is 4096
	GlobalMgoSession.SetPoolLimit(300)
}

//Use 使用Collection
func Use(colname string, f func(c *mgo.Collection)) {
	session := GlobalMgoSession.Clone()
	defer session.Close()
	col := session.DB(DbName).C(colname)
	f(col)
}

//Insert 新增
func Insert(model interface{}, robj interface{}, user base.User) (err error) {
	col, err := getCollection(model)
	if err != nil {
		return
	}
	Use(col, func(c *mgo.Collection) {
		refobj := reflect.ValueOf(robj).Elem()
		typeOfT := refobj.Type()
		for i := 0; i < refobj.NumField(); i++ {
			bstr := typeOfT.Field(i).Tag.Get("bson")
			if bstr == "_id" {
				refobj.Field(i).Set(reflect.ValueOf(bson.NewObjectId().Hex()))
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
func Remove(model interface{}, qjson string, user base.User) (err error) {
	col, err := getCollection(model)
	if err != nil {
		return
	}
	Use(col, func(c *mgo.Collection) {
		err = Update(model, qjson, `{"base.isdelete":"true"}`, user)
	})
	return
}

//Query 查询
func Query(model interface{}, qjson string, page, pageSize int, sort string, containsDeleted bool) (datas interface{}, total int, err error) {
	col, err := getCollection(model)
	if err != nil {
		return
	}
	qi, err := toQueryBson(qjson, containsDeleted)
	if err != nil {
		return
	}

	slice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(model)), 0, 0)
	temps := reflect.New(slice.Type())
	temps.Elem().Set(slice)
	Use(col, func(c *mgo.Collection) {
		if sort == "" {
			sort = "-updatetime -createtime"
		}
		qe := c.Find(qi).Sort(sort)
		if total, err = qe.Count(); err != nil {
			return
		}
		qe.Skip((page - 1) * pageSize).Limit(pageSize).All(temps.Interface())
	})
	return temps.Interface(), total, nil
}

//Update 更新
func Update(model interface{}, qjson, ujson string, user base.User) (err error) {
	col, err := getCollection(model)
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
	Use(col, func(c *mgo.Collection) {
		if err = c.Update(q, up); err != nil {
		}
	})
	return
}

//Count 计数
func Count(model interface{}, qjson string, containsDeleted bool) (count int, err error) {
	col, err := getCollection(model)
	if err != nil {
		return 0, err
	}
	var b bson.M
	b, err = toQueryBson(qjson, containsDeleted)
	if err != nil {
		return 0, err
	}
	Use(col, func(c *mgo.Collection) {
		q := c.Find(b)
		count, err = q.Count()
	})
	return
}

//GetCollection 获取模型对应的集合
func getCollection(model interface{}) (string, error) {
	col := ""
	t := reflect.TypeOf(model)
	for i := 0; i < t.NumField(); i++ {
		col = t.Field(i).Tag.Get("collection")
		if col != "" {
			break
		}
	}
	if col == "" {
		return col, errors.New("model:" + t.String() + ",未设置collection")
	}
	return col, nil
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

func getMongoID(rins interface{}) (bson.ObjectId, error) {
	refobj := reflect.ValueOf(rins).Elem()
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
	return "", errors.New("模型：" + typeOfT.String() + ",未找到 bson: _id 设置")
}

//DBRef 关联字段
func DBRef(model interface{}, rins interface{}) (ref *mgo.DBRef, err error) {
	col, err := getCollection(model)
	if err != nil {
		return nil, err
	}
	id, err := getMongoID(rins)
	if err != nil {
		return nil, err
	}
	return &mgo.DBRef{Collection: col, Id: id, Database: DbName}, nil
}
