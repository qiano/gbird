package model

import (
	"gopkg.in/mgo.v2/bson"
)

//Permission 权限
type Permission struct {
	ID         bson.ObjectId `bson:"_id"  collection:"permission" router:"permission 1111"`
	Code       string        `required:"true"`
	Desc       string        `required:"true"`
	ParentCode string
	Data       map[string]string
	Type       string `required:"true" enum:"菜单,按钮,API,其他"`
	Base
}

//Role 角色
type Role struct {
	ID          bson.ObjectId `bson:"_id"  collection:"role" router:"role 1111"`
	Code        string        `required:"true"`
	Name        string        `required:"true"`
	Type        string        `required:"true" enum:"业务角色,管理角色" default:"业务角色"`
	Permissions []string
	Base
}

//User 用户
type User struct {
	ID          bson.ObjectId `bson:"_id"  collection:"user"`
	UserName    string
	Password    string
	IsActive    bool
	Avatar      string
	Name        string
	Mobile      string
	Email       string
	Roles       []string
	Base
}
