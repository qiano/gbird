package auth

import (
	"errors"
	"gbird"
	"gbird/model"
	"gbird/mongodb"
	"gbird/util"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Register 注册
func Register(app *gbird.App) {
	app.Register(&Permission{}, nil, nil)
	app.Register(&Role{}, nil, nil)
	app.Register(&User{}, nil, nil)
	app.POST("/api/auth/login", loginHandler)
	app.POST("/api/auth/register", registerHandler)
	app.GET("/api/auth/refresh", func(c *gbird.Context) {
		tokenstring := getToken(c.Context)
		newtoken, err := refresh(tokenstring)
		if err != nil {
			c.RetError(err)
		} else {
			c.Ret(gin.H{"token": newtoken})
		}
	})
	app.POST("/api/auth/verify", func(c *gbird.Context) {
		data, err := verify(c)
		if err != nil {
			c.RetError(err)
		} else {
			c.Ret(data)
		}
	})
}

//Permission 权限
type Permission struct {
	ID   bson.ObjectId `bson:"_id"  collection:"permission"`
	Name string
	PID  bson.ObjectId
	Type string `enum:"一级菜单,二级菜单,三级菜单,按钮,其他"`
	model.Base
}

//Role 角色
type Role struct {
	ID          bson.ObjectId `bson:"_id"  collection:"role"`
	Name        string
	Type        string      `enum:"业务角色,管理角色" default:"业务角色"`
	Permissoins []mgo.DBRef `ref:"permission"`
	model.Base
}

//User 用户
type User struct {
	ID       bson.ObjectId `bson:"_id"  collection:"user"`
	UserName string
	Password string
	IsActive bool
	Avatar   string
	Name     string
	Mobile   string
	Email    string
	Roles    []mgo.DBRef `ref:"role"`
	model.Base
}

//登陆
func loginHandler(c *gbird.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	u, err := mongodb.FindOne(&User{}, bson.M{"username": username, "password": password, "isactive": true})
	if err != nil {
		c.RetError(errors.New("帐号或密码错误 " + err.Error()))
		return
	}
	uu := u.(User)
	tokenString, err := CreateJWTToken("pp", &uu)
	if err != nil {
		c.RetError(err)
		return
	}
	c.Ret(gin.H{"token": tokenString, "user": u})
}

//注册
func registerHandler(c *gbird.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	err := mongodb.Insert(&User{UserName: username, Password: util.Md5(password), IsActive: true}, "")
	if err != nil {
		c.RetError(err)
		return
	}
	c.Ret("添加成功")
}
