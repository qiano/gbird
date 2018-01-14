package base

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"

	"time"
)

//ObjID 实体主键类型
type ObjID bson.ObjectId

//HandlerFunc 处理方法
type HandlerFunc gin.HandlerFunc

//Base 模型基础字段
type Base struct {
	Creater    string    //创建人
	CreateTime time.Time //创建时间
	Updater    string    //创建人
	UpdateTime time.Time //创建时间
	IsDelete   bool      //是否已删除
}

//User 用户
type User struct {
	ID       string `json:"_id"`
	Name     string
	UserName string
	Roles    []interface{}
	IsActive bool `json:"Is_Active"`
}
