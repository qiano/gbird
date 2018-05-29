package gbird

import (
	// "encoding/json"
	// "fmt"
	"gbird/config"
	"gbird/logger"
	"github.com/gin-gonic/gin"
	"github.com/tommy351/gin-sessions"
	// "io/ioutil"
	// "net/http"
	"strings"
)

//UserInterface 用户接口
type UserInterface interface {
	DisplayName() string
	UserID() string
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

//白名单处理
func whitelist(c *gin.Context) bool {
	cip := c.ClientIP()
	origin := c.Request.Header.Get("origin")
	for _, ip := range strings.Split(config.Config["whitelist"], ",") {
		if ip == "*" {
			return true
		}
		if ip == "" {
			break
		}
		if origin == ip || cip == ip {
			return true
		}
	}
	return false
}

func getToken(c *gin.Context) string {
	token := c.Request.Header.Get("token")
	if token == "" {
		token = c.Query("token")
	}
	if token == "" {
		token = c.Param("token")
	}
	if token == "" {
		token = c.PostForm("token")
	}
	return token
}

//Middleware 权限中间件
func Middleware(authFn func(*Context) bool, verifyTokenFn func(string) (User, error)) func(*Context) {
	logger.Infoln("帐户权限验证：开启")
	return GinToBird(
		func(c *gin.Context) {
			if !whitelist(c) && !authFn(&Context{Context: c}) && verifyTokenFn != nil {
				token := getToken(c)
				if token != "" {
					// res, err := http.Get(verifyURL + "?token=" + token)
					// if err != nil {
					// 	fmt.Println(err.Error())
					// 	return
					// }
					// body, _ := ioutil.ReadAll(res.Body)
					// defer res.Body.Close()
					// str := string(body)
					// if strings.Contains(str, "errcode") {
					// 	c.Status(res.StatusCode)
					// 	for key, vals := range res.Header {
					// 		if key == "Access-Control-Allow-Origin" {
					// 			continue
					// 		}
					// 		for _, val := range vals {
					// 			c.Writer.Header().Add(key, val)
					// 		}
					// 	}
					// 	c.Writer.Write(body)
					// 	c.Abort()
					// 	return
					// }
					// json.Unmarshal(body, &wu)
					u, err := verifyTokenFn(token)
					if err == nil {
						ss := sessions.Get(c)
						ss.Set("user", &u)
						ss.Save()
					} else {
						c.AbortWithStatusJSON(200, gin.H{"errcode": 0, "errmsg": err.Error()})
						return
					}
					c.Next()
					return
				}
				logger.Fatalln(c.Request.URL.Path + " no token")
				c.AbortWithStatusJSON(200, gin.H{"errcode": 0, "errmsg": "no token"})
				return
			}
			c.Next()
		})
}
