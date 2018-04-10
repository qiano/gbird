package gbird

import (
	"errors"
	"gbird/logger"
	"gbird/model"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/tommy351/gin-sessions"
	"strconv"
)

//H h
type H gin.H

//Context 上下文
type Context struct {
	*gin.Context
}

//App 应用实例
type App struct {
	*gin.Engine
	TaskManager *cron.Cron
	Name        string
}

//NewApp 创建实例
func NewApp(name string) *App {
	logger.Infoln("应用启动：" + name)
	var store = sessions.NewCookieStore([]byte(name))
	r := gin.Default()
	r.Static("/assets", "./assets")
	r.Static("/doc", "./doc")
	r.Use(sessions.Middleware(name+"session", store))
	r.Use(CORSMiddleware())
	app := &App{Engine: r, Name: name, TaskManager: cron.New()}
	r.GET("/isok", func(c *gin.Context) {
		c.String(200, name+"  server")
	})
	r.GET("/api/metadata", func(c *gin.Context) {
		m, _ := c.GetQuery("model")
		if m != "" {
			c.JSON(200, H{"data": model.Metadatas[m]})
		} else {
			c.JSON(200, H{"data": model.Metadatas})
		}
	})
	app.TaskManager.Start()
	return app
}

//CORSMiddleware 跨域
func CORSMiddleware() gin.HandlerFunc {
	logger.Infoln("跨域：开启")
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("origin")
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("vary", "Origin")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

//Ret 返回数据
func (c *Context) Ret(datas ...interface{}) {
	r := H{"data": datas[0]}
	if len(datas) > 1 {
		for i := 1; i < len(datas); i++ {
			r["data"+strconv.Itoa(i)] = datas[i]
		}
	}
	c.JSON(200, r)
}

//RetError 返回错误
func (c *Context) RetError(err error) {
	logger.Fatalln(err)
	c.JSON(200, H{"errmsg": err.Error()})
}

//GetCurUser 获取当前用户ID和名称
var GetCurUser = func(r *Context) (UserInterface, error) {
	ss := sessions.Get(r.Context)
	user := ss.Get("user")
	if user != nil {
		u := user.(*User)
		return u, nil
	}
	return nil, errors.New("未找到当前用户")
}

//GetSession getsession
func GetSession(c *Context) sessions.Session {
	ss := sessions.Get(c.Context)
	return ss
}

//Use use
func (a *App) Use(middleware ...func(*Context)) {
	a.Engine.Use(BirdToGin(middleware...)...)
}

//POST post
func (a *App) POST(relativePath string, handlers ...func(c *Context)) {
	a.Engine.POST(relativePath, BirdToGin(handlers...)...)
}

//GET get
func (a *App) GET(relativePath string, handlers ...func(c *Context)) {
	a.Engine.GET(relativePath, BirdToGin(handlers...)...)
}

//PUT put
func (a *App) PUT(relativePath string, handlers ...func(c *Context)) {
	a.Engine.PUT(relativePath, BirdToGin(handlers...)...)
}

//DELETE delete
func (a *App) DELETE(relativePath string, handlers ...func(c *Context)) {
	a.Engine.DELETE(relativePath, BirdToGin(handlers...)...)
}

//BirdToGin 类型转换
func BirdToGin(handlers ...func(c *Context)) []gin.HandlerFunc {
	ginHandlers := make([]gin.HandlerFunc, 0, 0)
	for _, handler := range handlers {
		ginHandlers = append(ginHandlers, func(ginc *gin.Context) {
			handler(&Context{Context: ginc})
		})
	}
	return ginHandlers
}

//GinToBird 类型转换
func GinToBird(handler func(c *gin.Context)) func(*Context) {
	return func(gc *Context) {
		handler(gc.Context)
	}
}
