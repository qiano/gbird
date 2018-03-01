package agency

import (
	"gbird/logger"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
)

//Middleware 代理中间件
func Middleware(getMap func(*gin.Context) string) gin.HandlerFunc {
	logger.Infoln("代理功能：开启")
	return func(c *gin.Context) {
		if strings.ToLower(c.Request.Method) == "options" {
			c.AbortWithStatus(204)
			return
		}
		if target := getMap(c); len(target) > 0 {
			if c.Request.URL.RawQuery != "" {
				target = target + "?" + c.Request.URL.RawQuery
			}
			logger.Infoln(c.Request.Method+" "+c.Request.RequestURI, " --> ", target)
			req, err := http.NewRequest(c.Request.Method, target, c.Request.Body)
			req.Header = c.Request.Header
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				logger.Fatalln(err)
				c.AbortWithStatusJSON(500, gin.H{"errcode": 0, "errmsg": err.Error()})
				return
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			for key, vals := range resp.Header {
				if key == "Access-Control-Allow-Origin" ||
					key == "Access-Control-Allow-Credentials" ||
					key == "Access-Control-Allow-Headers" ||
					key == "Access-Control-Allow-Methods" ||
					key == "Vary" {
					continue
				}
				for _, val := range vals {
					c.Writer.Header().Add(key, val)
				}
			}
			c.Status(resp.StatusCode)
			c.Writer.Write(body)
			c.Abort()
			return
		}
		c.Next()
	}
}
