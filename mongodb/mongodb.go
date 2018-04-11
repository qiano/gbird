package mongodb

import (
	"errors"
	"gbird/logger"
	"gbird/model"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"time"
)

//Use 使用Mongo数据库
func Use(mongodbstr, dbName string) {
	DbName = dbName
	if mongodbstr == "" {
		logger.Infoln("未启用 Mongodb 数据库")
		return
	}
	globalMgoSession, err := mgo.DialWithTimeout(mongodbstr, 10*time.Second)
	if err != nil {
		logger.Errorln("Mongodb：", err)
		panic(err)
	}
	logger.Infoln("Mondodb连接成功：" + mongodbstr + "  " + DbName)
	GlobalMgoSession = globalMgoSession
	GlobalMgoSession.SetMode(mgo.Monotonic, true)
	//default is 4096
	GlobalMgoSession.SetPoolLimit(300)

}

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
	// _, bval := model.GetValue(robj, "Base")
	// if bval == nil {
	model.SetValue(robj, "Base", model.Base{
		Creater:    userid,
		CreateTime: time.Now(),
		Updater:    userid,
		UpdateTime: time.Now(),
	})
	// }
	UseCol(col, func(c *mgo.Collection) {
		err = c.Insert(&robj)
	})
	return
}

//Remove 删除
func Remove(robj interface{}, qi bson.M, userid string, batch bool) (info *mgo.ChangeInfo, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	UseCol(col, func(c *mgo.Collection) {
		info, err = Update(robj, qi, bson.M{"base.isdelete": true}, userid, batch)
	})
	return
}

//Query 查询
func Query(robj interface{}, q bson.M, page, pageSize int, sort string, fields string, containsDeleted bool) (datas interface{}, total int, err error) {
	col, err := getCollection(robj)
	if err != nil {
		logger.Fatalln(err)
		return
	}
	qi, err := toQueryBson(robj, q, containsDeleted)
	if err != nil {
		logger.Fatalln(err)
		return
	}
	fd, err := ToBson(fields)
	if err != nil {
		logger.Fatalln(err)
		return
	}
	if sort == "" {
		sort = "-base.updatetime"
	}
	refobj := reflect.ValueOf(robj).Elem()
	t := refobj.Type()
	slice := reflect.MakeSlice(reflect.SliceOf(t), 0, 0)
	temps := reflect.New(slice.Type())
	temps.Elem().Set(slice)
	UseCol(col, func(c *mgo.Collection) {
		qe := c.Find(qi).Sort(sort).Select(fd)
		if total, err = qe.Count(); err != nil {
			logger.Fatalln(err)
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
	UseCol(col, func(c *mgo.Collection) {
		err = c.FindId(id).One(robj)
	})
	return robj, err

}

//FindOne 查找一个
func FindOne(robj interface{}, qi bson.M) (interface{}, error) {
	data, total, err := Query(robj, qi, 1, 1, "", "", false)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, errors.New("未找到匹配的数据")
	}
	arr := model.ToSlice(data)
	return arr[0], nil
}

//FindAll 查找所有
func FindAll(robj interface{}, qi bson.M, sort string) ([]interface{}, error) {
	data, total, err := Query(robj, qi, 0, 0, "", sort, false)
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
func Update(robj interface{}, q, u bson.M, userid string, batch bool) (info *mgo.ChangeInfo, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	qi, err := toQueryBson(robj, q, false)
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
			if info, err = c.UpdateAll(qi, up); err != nil {
			}
		} else {
			if err = c.Update(qi, up); err != nil {
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

//UpsertID UpsertId
func UpsertID(robj interface{}) (info *mgo.ChangeInfo, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return
	}
	id, err := model.GetID(robj)

	UseCol(col, func(c *mgo.Collection) {
		info, err = c.UpsertId(id, robj)
	})
	return
}

//Count 计数
func Count(robj interface{}, qi bson.M) (count int, err error) {
	col, err := getCollection(robj)
	if err != nil {
		return 0, err
	}
	var b bson.M
	b, err = toQueryBson(robj, qi, false)
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
	id, err := model.GetID(robj)
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

//GetCollection 获取模型对应的集合
func getCollection(robj interface{}) (string, error) {
	return model.MTagVal(robj, "collection")
}
