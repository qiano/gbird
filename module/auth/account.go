package auth

import (
	"errors"
	"gbird"
	"gbird/model"
	"gbird/mongodb"
	"gbird/util"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

//Register 注册
func Register(app *gbird.App) {
	app.Register(&model.Permission{}, nil, nil)
	app.Register(&model.Role{}, nil, nil)
	app.Register(&model.User{}, nil, nil)
	app.POST("/api/auth/login", loginHandler)
	app.POST("/api/auth/register", registerHandler)
	app.POST("/api/auth/verify", func(c *gbird.Context) {
		data, err := verify(c)
		if err != nil {
			c.RetError(err)
		} else {
			c.Ret(data)
		}
	})
	app.TaskManager.AddFunc("@every 3h3m", func() {
		temp := make([]string, 0, 0)
		for i := 0; i < len(blackTokenList); i++ {
			_, err := validation(blackTokenList[i])
			if err == nil {
				temp = append(temp, blackTokenList[i])
			}
		}
		blackTokenList = temp
	})
}

//登陆
func loginHandler(c *gbird.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	u, err := mongodb.FindOne(&model.User{}, bson.M{"username": username, "password": password, "isactive": true})
	if err != nil {
		c.RetError(errors.New("帐号或密码错误 " + err.Error()))
		return
	}
	uu := u.(model.User)
	ps := GetPermissions(uu.Roles)
	tokenString, err := CreateJWTToken("pp", map[string]interface{}{"user": &uu, "permissions": ps})
	if err != nil {
		c.RetError(err)
		return
	}
	c.Ret(gin.H{"token": tokenString, "user": u, "permissions": ps})
}

//失效token
var blackTokenList = make([]string, 0, 0)
func logoutHandler(c *gbird.Context) {
	tk := GetTokenString(c.Context)
	if tk != "" {
		blackTokenList = append(blackTokenList, tk)
	}
}

//注册
func registerHandler(c *gbird.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	err := mongodb.Insert(&model.User{UserName: username, Password: util.Md5(password), IsActive: true}, "")
	if err != nil {
		c.RetError(err)
		return
	}
	c.Ret("添加成功")
}

//GetPermissions 获取用户所有权限
func GetPermissions(roles []string) []string {
	ps := make([]string, 0, 0)
	for i := 0; i < len(roles); i++ {
		r, err := mongodb.FindOne(&model.Role{}, bson.M{"code": roles[i]})
		if err != nil {
			continue
		}
		rr := r.(model.Role)
		ps = append(ps, rr.Permissions...)
	}
	return ps
}

//GetCurUser 获取当前用户ID和名称
func GetCurUser(c *gbird.Context) (map[string]interface{}, error) {
	_, data := GetTokenData(c)
	if data != nil {
		dd := data.(map[string]interface{})

		return dd["user"].(map[string]interface{}), nil
	}
	return nil, errors.New("未找到当前用户")
}

//GetCurPermission 获取当前用户权限信息
func GetCurPermission(c *gbird.Context) ([]string, error) {
	_, data := GetTokenData(c)
	if data != nil {
		dd := data.(map[string]interface{})

		return dd["permissions"].([]string), nil
	}
	return nil, errors.New("未找到当前用户")
}
