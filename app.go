package gbird

import (
	"fmt"
	"gbird/config"
	"gbird/logger"
	"gbird/model"
	"gbird/mongodb"
	"gbird/util"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/tommy351/gin-sessions"
	"gopkg.in/mgo.v2"
	"io"
	"mime/multipart"
	"os"
	"time"
)

//Module 接口
type Module interface {
	Register(a *App)
}

//App 应用实例
type App struct {
	*gin.Engine
	TaskManager *cron.Cron
	Name        string
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
	r.GET("/", func(c *gin.Context) {
		c.String(200, name+"  server")
	})
	r.GET("/api/metadata", func(c *gin.Context) {
		m, _ := c.GetQuery("model")
		if m != "" {
			Ret(c, gin.H{"data": model.Metadatas[m]}, nil, 0)
		} else {
			Ret(c, gin.H{"data": model.Metadatas}, nil, 0)
		}
	})
	app.TaskManager.Start()
	return app
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
