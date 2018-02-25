package gbird

import (
	"fmt"
	"gbird/model"
	mw "gbird/middleware"
	"gbird/mongodb"
	"gbird/util"
	"gbird/util/config"
	"gbird/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/tommy351/gin-sessions"
	"gopkg.in/mgo.v2"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"time"
)

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
	r.Use(mw.CORSMiddleware())
	app := &App{Engine: r, Name: name, TaskManager: cron.New()}
	r.GET("/", func(c *gin.Context) {
		c.String(200, name+" module server")
	})
	r.GET("/api/metadata", func(c *gin.Context) {
		m, _ := c.GetQuery("model")
		if m != "" {
			Ret(c, gin.H{"data": model.Metadatas[m]}, nil, 0)
		} else {
			Ret(c, gin.H{"data": model.Metadatas}, nil, 0)
		}
	})
	return app
}

//Router 路由注册
func (a *App) Router(method, path string, f func(*gin.Context)) {
	m := strings.ToUpper(method)
	a.Engine.Handle(m, path, f)
	// logger.Infoln("路由注册：" + m + " " + path)
}

//ModelRouter 注册模型下的路由
func (a *App) ModelRouter(robj interface{}, method, path string, f func(*gin.Context)) {
	rname, err := model.MTagVal(robj, "urlname")
	if err != nil {
		panic(err)
	}
	grp := a.Group("/api/" + rname)
	m := strings.ToUpper(method)
	grp.Handle(m, path, f)
	// logger.Infoln("路由注册：" + m + " /api/" + rname + path)
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

//SaveUploadedFile 保存上传的文件
func SaveUploadedFile(file *multipart.FileHeader, flag string) (path string, err error) {
	basedir := util.GetCurrDir()
	savedir := "/assets/" + flag + "/"
	if !util.IsExistFileOrDir(basedir + savedir) {
		os.MkdirAll(basedir+savedir, 0777) //创建文件夹
	}
	tnow := fmt.Sprintf("%d", time.Now().UnixNano())
	path = basedir + savedir + tnow + "_" + flag + "_" + file.Filename
	src, err := file.Open()
	if err != nil {
		return
	}
	defer src.Close()

	out, err := os.Create(path)
	if err != nil {
		return
	}
	defer out.Close()
	io.Copy(out, src)

	return
}
