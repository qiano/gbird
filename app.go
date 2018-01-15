package gbird

import (
	"encoding/json"
	"errors"
	"gbird/base"
	"gbird/logger"
	m "gbird/mongodb"
	"github.com/gin-gonic/gin"
	"github.com/tommy351/gin-sessions"
	"reflect"
	"strconv"
	"strings"
)

//App 应用实例
type App struct {
	*gin.Engine
}

//NewApp 创建实例
func NewApp(name string) *App {

	var store = sessions.NewCookieStore([]byte(name))
	r := gin.Default()
	r.Static("/assets", "./assets")
	r.Use(sessions.Middleware(name+"session", store))
	// r.Use(mw.CORSMiddleware())
	// r.Use(mw.AgencyMiddleware())
	// r.Use(mw.AuthMiddleware())
	r.GET("/", func(c *gin.Context) {
		c.String(200, "apihub module server")
	})
	return &App{Engine: r}
}

//RegisterModel 模型注册
func (r *App) RegisterModel(model interface{}, refmodel interface{}, routers ...*base.Router) {
	rname, err := getRouterName(model)
	if err != nil {
		panic(err)
	}
	soles := getSoles(model)
	grp := r.Group("/api/" + rname)

	grp.GET("/query", func(c *gin.Context) {
		sort, _ := c.GetQuery("sort")
		pageIndex, _ := c.GetQuery("page")
		pageSize, _ := c.GetQuery("size")
		cond, _ := c.GetQuery("cond")
		idx, _ := strconv.Atoi(pageIndex)
		if idx == 0 {
			idx = 1
		}
		size, _ := strconv.Atoi(pageSize)
		datas, total, err := m.Query(model, cond, idx, size, sort, false)
		if err != nil {
			c.JSON(200, gin.H{"code": 1, "error": err})
		} else {
			c.JSON(200, gin.H{"code": 0, "data": gin.H{
				"list":  datas,
				"total": total,
				"page":  idx}})
		}
	})
	logger.Infoln("路由注册：GET" + " /api/" + rname + "/query")

	grp.POST("/insert", func(c *gin.Context) {
		data := c.PostForm("data")
		objType := reflect.TypeOf(refmodel).Elem()
		obj := reflect.New(objType).Interface()
		json.Unmarshal([]byte(data), &obj)

		temps := []string{}
		for _, val := range soles {
			field := reflect.ValueOf(obj).Elem().FieldByName(val)
			v := field.Interface().(string)
			temps = append(temps, `"`+strings.ToLower(val)+`":"`+v+`"`)
		}
		var exist string
		if len(temps) > 0 {
			exist = `{` + strings.Join(temps, ",") + `}`
		}

		if count, err := m.Count(model, exist, false); err == nil {
			if count == 0 {
				ss := sessions.Get(c)
				user := ss.Get("user")
				var u base.User
				if user != nil {
					u = user.(base.User)
				}
				err := m.Insert(model, obj, u)
				if err != nil {
					c.JSON(500, gin.H{"errmsg": err.Error()})
				} else {
					c.JSON(200, gin.H{"data": obj})
				}
			} else {
				c.JSON(500, gin.H{"errmsg": "数据已存在，查询条件：" + exist})

			}
		}

	})
	logger.Infoln("路由注册：POST" + " /api/" + rname + "/insert")

	grp.POST("/delete", func(c *gin.Context) {
		data := c.PostForm("cond")
		batch := c.PostForm("batch")
		b, err := strconv.ParseBool(batch)
		if err != nil {
			b = false
		}
		ss := sessions.Get(c)
		user := ss.Get("user")
		var u base.User
		if user != nil {
			u = user.(base.User)
		}
		info, err := m.Remove(model, data, u, b)
		if err != nil {
			c.JSON(500, gin.H{"errmsg": err.Error(), "data": info})
		} else {
			c.JSON(200, gin.H{"data": info})
		}
	})
	logger.Infoln("路由注册：POST" + " /api/" + rname + "/delete")

	for _, router := range routers {
		method := strings.ToUpper(router.Method)
		grp.Handle(method, router.RelativePath, func(c *gin.Context) {
			ctx := &base.Context{Context: *c}
			router.HandlerFunc(ctx)
		})
		logger.Infoln("路由注册：" + method + " /api/" + rname + "/" + router.RelativePath)
	}

}

//RegisterRouter 路由注册
func (r *App) RegisterRouter(routers ...*base.Router) {
	for _, router := range routers {
		method := strings.ToUpper(router.Method)
		r.Handle(method, router.RelativePath, func(c *gin.Context) {
			ctx := &base.Context{Context: *c}
			router.HandlerFunc(ctx)
		})
		logger.Infoln("路由注册：" + method + " " + router.RelativePath)
	}
}

func getRouterName(model interface{}) (string, error) {
	col := ""
	t := reflect.TypeOf(model)
	for i := 0; i < t.NumField(); i++ {
		col = t.Field(i).Tag.Get("uname")
		if col != "" {
			break
		}
	}
	if col == "" {
		return col, errors.New("model:" + t.String() + ",未设置路由名称")
	}
	return col, nil
}

func getSoles(model interface{}) []string {
	soles := []string{}
	t := reflect.TypeOf(model)
	for i := 0; i < t.NumField(); i++ {
		val := t.Field(i).Tag.Get("sole")
		if val != "" && val == "true" {
			soles = append(soles, t.Field(i).Name)
		}
	}
	return soles
}
