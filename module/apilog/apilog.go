package apilog

import (
	"bytes"
	"gbird"
	"gbird/logger"
	m "gbird/mongodb"
	"gbird/router"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"strings"
)

//Register 模型注册
func Register(app *gbird.App) {
	router.Register(app, &APILog{}, nil, nil)
}

//APILog api調用日誌
type APILog struct {
	ID                bson.ObjectId `bson:"_id"  collection:"apilog" router:"apilog 0 0 0 0"`
	RequestURL        string        //请求路径
	RequestMethod     string        //调用方式
	RequestDesc       string        //描述
	RequestHeaders    http.Header   //请求头
	UserAgent         string
	QueryStringParams string
	RequestBody       string //请求体
	IP                string //IP
	UserID            string //用户ID
	UserName          string //用户名
	gbird.Base
}

//Middleware API日志中间件
func Middleware(getDesc func(string) string) func(*gbird.Context) {
	logger.Infoln("API日志：开启")
	return gbird.GinToBird(func(c *gin.Context) {
		rbody, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		bf := bytes.NewBuffer(rbody)
		c.Request.Body = ioutil.NopCloser(bf)

		desc := getDesc(c.Request.URL.Path)
		log := &APILog{
			UserAgent:         c.Request.UserAgent(),
			RequestURL:        strings.ToLower(c.Request.URL.Path),
			RequestBody:       (string)(rbody),
			RequestHeaders:    c.Request.Header,
			RequestMethod:     strings.ToUpper(c.Request.Method),
			IP:                c.ClientIP(),
			QueryStringParams: c.Request.URL.RawQuery,
			RequestDesc:       desc}
		var uid string
		user, err := gbird.GetCurUser(&gbird.Context{Context: c})
		if err == nil {
			uid = user.UserID()
			log.UserID = user.UserID()
			log.UserName = user.DisplayName()
		}
		m.Insert(log, uid)
		c.Next()
	})
}
