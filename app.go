package gbird

import (
	"gbird/config"
	"gbird/logger"
	mw "gbird/middleware"
	"gbird/mongodb"
	"github.com/gin-gonic/gin"
	"github.com/tommy351/gin-sessions"
	"gopkg.in/mgo.v2"
	"strings"
	"time"
)

//App 应用实例
type App struct {
	*gin.Engine
	Name string
}

//NewApp 创建实例
func NewApp(name string) *App {
	logger.Infoln("应用启动：" + name)
	var store = sessions.NewCookieStore([]byte(name))
	r := gin.Default()
	r.Static("/assets", "./assets")
	r.Use(sessions.Middleware(name+"session", store))
	r.Use(mw.CORSMiddleware())
	app := &App{Engine: r, Name: name}

	r.POST("/", func(c *gin.Context) {
		c.String(200, name+" module server")
	})
	return app
}

//Router 路由注册
func (a *App) Router(method, path string, f func(*gin.Context)) {
	m := strings.ToUpper(method)
	a.Engine.Handle(m, path, f)
	logger.Infoln("路由注册：" + m + " " + path)
}

//ModelRouter 注册模型下的路由
func (a *App) ModelRouter(robj interface{}, method, path string, f func(*gin.Context)) {
	rname, err := getRouterName(robj)
	if err != nil {
		panic(err)
	}
	grp := a.Group("/api/" + rname)
	m := strings.ToUpper(method)
	grp.Handle(m, path, f)
	logger.Infoln("路由注册：" + m + " /api/" + rname + path)
}

//UseMongodb 使用Mongo数据库
func (a *App) UseMongodb() {
	mongodbstr := config.Config["mongodbHost"]
	mongodb.DbName = config.Config["mongodbDbName"]
	if mongodbstr == "" {
		logger.Infoln("未启用 Mongodb 数据库")
		return
	}
	globalMgoSession, err := mgo.DialWithTimeout(mongodbstr, 10*time.Second)
	if err != nil {
		logger.Errorln("Mongodb：", err)
		panic(err)
	}
	logger.Infoln("Mondodb连接成功：" + mongodbstr + "  " + mongodb.DbName)
	mongodb.GlobalMgoSession = globalMgoSession
	mongodb.GlobalMgoSession.SetMode(mgo.Monotonic, true)
	//default is 4096
	mongodb.GlobalMgoSession.SetPoolLimit(300)

}


