package middleware

import (
	"fmt"
	"gbird/logger"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
)

//AgencyMiddleware 代理中间件
func AgencyMiddleware(getMap func(string) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.ToLower(c.Request.Method) == "options" {
			c.Next()
			return
		}
		if target := getMap(c.Request.URL.Path); len(target) > 0 {
			if c.Request.URL.RawQuery != "" {
				target = target + "?" + c.Request.URL.RawQuery
			}
			logger.Infoln(c.Request.RequestURI, "->", target)

			req, err := http.NewRequest(c.Request.Method, target, c.Request.Body)
			req.Header = c.Request.Header
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				logger.Fatalln(err)
				fmt.Println(err)
				c.JSON(500, gin.H{"errcode": 0, "errmsg": err.Error()})
				return
			}
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
			c.Status(resp.StatusCode)
			for key, vals := range resp.Header {
				if key == "Access-Control-Allow-Origin" {
					continue
				}
				for _, val := range vals {
					c.Writer.Header().Add(key, val)
				}
			}

			c.Writer.Write(body)
			defer resp.Body.Close()

			return
		}
		c.Next()
	}
}