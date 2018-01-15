package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	// "apihub/common/logger"
	"encoding/json"
	"gbird/base"
	"github.com/tommy351/gin-sessions"
	"io/ioutil"
	"net/http"
	"strings"
)

type wrapuser struct {
	Data base.User
}

//AuthMiddleware 权限中间件
func AuthMiddleware(verifyURL string, needAuth func(string) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if needAuth(c.Request.URL.Path) {
			if token := c.Request.Header.Get("token"); token != "" {
				res, err := http.Get(verifyURL + "?token=" + token)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				body, _ := ioutil.ReadAll(res.Body)
				defer res.Body.Close()
				str := string(body)
				var wu wrapuser
				json.Unmarshal(body, &wu)
				ss := sessions.Get(c)
				ss.Set("user", wu.Data)
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
