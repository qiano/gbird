package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"gbird/logger"
	"github.com/gin-gonic/gin"
	"github.com/tommy351/gin-sessions"
	"io/ioutil"
	"net/http"
	"strings"
)

//UserInterface 用户接口
type UserInterface interface {
	DisplayName() string
	UserID() string
}

//GetCurUser 获取当前用户ID和名称
var GetCurUser = func(r *gin.Context) (UserInterface, error) {
	ss := sessions.Get(r)
	user := ss.Get("user")
	if user != nil {
		u := user.(*User)
		return u, nil
	}
	return nil, errors.New("未找到当前用户")
}

//User 用户
type User struct {
	ID       string `json:"_id"`
	Name     string
	UserName string
	Roles    []interface{}
	IsActive bool `json:"Is_Active"`
}

//DisplayName  显示名称
func (u *User) DisplayName() string {
	return u.Name
}

//UserID  用户ID
func (u *User) UserID() string {
	return u.ID
}

//Middleware 权限中间件
func Middleware(verifyURL string, needAuth func(*gin.Context) bool) gin.HandlerFunc {
	logger.Infoln("帐户权限验证：开启")
	return func(c *gin.Context) {
		cip := c.ClientIP()
		if cip != "127.0.0.1" && needAuth(c) {
			if token := c.Request.Header.Get("token"); token != "" {
				res, err := http.Get(verifyURL + "?token=" + token)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				body, _ := ioutil.ReadAll(res.Body)
				defer res.Body.Close()
				str := string(body)
				var wu struct{ Data User }
				json.Unmarshal(body, &wu)
				ss := sessions.Get(c)
				ss.Set("user", &wu.Data)
				ss.Save()
				if strings.Contains(str, "errcode") {
					c.Status(res.StatusCode)
					for key, vals := range res.Header {
						if key == "Access-Control-Allow-Origin" {
							continue
						}
						for _, val := range vals {
							c.Writer.Header().Add(key, val)
						}
					}
					c.Writer.Write(body)
					c.Abort()
					return
				}
				c.Next()
				return
			}
			c.AbortWithStatusJSON(200, gin.H{"errorcode": 0, "errormsg": "no token"})
			return
		}
		c.Next()

	}
}
