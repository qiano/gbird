package base

import (
	"github.com/gin-gonic/gin"
	"github.com/tommy351/gin-sessions"
	"gopkg.in/mgo.v2/bson"
	"time"
)

//ObjID 实体主键类型
type ObjID bson.ObjectId

//Context 上下文
type Context struct{ gin.Context }

//HandlerFunc 处理方法
type HandlerFunc gin.HandlerFunc

//Router 自定义路由
type Router struct {
	Method       string
	RelativePath string
	HandlerFunc  func(*Context)
}

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

//CurUser 获取当前用户信息
func (r *Context) CurUser() User {
	ss := sessions.Get(&(r.Context))
	user := ss.Get("user")
	return user.(User)
}
