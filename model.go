package gbird

import (
	"encoding/json"
	"errors"
	"gbird/base"
	"gbird/logger"
	m "gbird/mongodb"
	"github.com/gin-gonic/gin"
	"reflect"
	"strconv"
	"strings"
)

//Register 模型注册
func (r *App) Register(robj interface{}) {
	rname, err := getRouterName(robj)
	if err != nil {
		panic(err)
	}
	soles := getSoles(robj)
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
		datas, total, err := m.Query(robj, cond, idx, size, sort, false)
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
		objType := reflect.TypeOf(robj).Elem()
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

		if count, err := m.Count(robj, exist, false); err == nil {
			if count == 0 {
				ctx := base.Context{Context: c}
				u := ctx.CurUser()
				err := m.Insert(obj, u)
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
		ctx := base.Context{Context: c}
		u := ctx.CurUser()
		info, err := m.Remove(robj, data, u, b)
		if err != nil {
			c.JSON(500, gin.H{"errmsg": err.Error(), "data": info})
		} else {
			c.JSON(200, gin.H{"data": info})
		}
	})
	logger.Infoln("路由注册：POST" + " /api/" + rname + "/delete")
}

func getRouterName(robj interface{}) (string, error) {
	col := ""
	refobj := reflect.ValueOf(robj).Elem()
	t := refobj.Type()
	for i := 0; i < refobj.NumField(); i++ {
		col = t.Field(i).Tag.Get("urlname")
		if col != "" {
			break
		}
	}
	if col == "" {
		return col, errors.New("model:" + t.String() + ",未设置路由名称")
	}

	return col, nil
}

func getSoles(robj interface{}) []string {
	soles := []string{}

	refobj := reflect.ValueOf(robj).Elem()
	t := refobj.Type()
	for i := 0; i < refobj.NumField(); i++ {
		val := t.Field(i).Tag.Get("sole")
		if val != "" && val == "true" {
			soles = append(soles, t.Field(i).Name)
		}
	}
	return soles
}
