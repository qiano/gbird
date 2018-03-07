package gbird

import (
	"errors"
	"fmt"
	"gbird/logger"
	"gbird/model"
	"gbird/util"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/tommy351/gin-sessions"
	"io"
	"mime/multipart"
	"os"
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
	r.Use(CORSMiddleware())
	app := &App{Engine: r, Name: name, TaskManager: cron.New()}
	r.GET("/", func(c *gin.Context) {
		c.String(200, name+"  server")
	})
	r.GET("/api/metadata", func(c *gin.Context) {
		m, _ := c.GetQuery("model")
		if m != "" {
			Ret(&Context{c}, H{"data": model.Metadatas[m]}, nil, 0)
		} else {
			Ret(&Context{c}, H{"data": model.Metadatas}, nil, 0)
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

//Base 模型基础字段
type Base struct {
	Creater    string    //创建人
	CreateTime time.Time //创建时间
	Updater    string    //创建人
	UpdateTime time.Time //创建时间
	IsDelete   bool      //是否已删除
}

//Ret 返回值
func Ret(c *Context, data H, err error, code int) {
	if err != nil {
		c.JSON(200, gin.H{"errcode": code, "errmsg": err.Error()})
	} else {
		c.JSON(200, data)
	}
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
