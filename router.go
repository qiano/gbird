package gbird

import (
	"encoding/json"
	"errors"
	"gbird/auth"
	"gbird/base"
	"math"
	// "gbird/logger"
	m "gbird/mongodb"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

//Register 模型注册
func (r *App) Register(robj interface{}, before func(c *gin.Context), after func(*gin.Context, interface{}, error) error) {
	base.RegisterMetadata(robj)
	rname, _, err := base.FindTag(robj, "urlname", "")
	if err != nil {
		panic(err)
	}
	soles, err := base.GetFieldsWithTag(robj, "sole")
	if err != nil {
		panic(err)
	}
	grp := r.Group("/api/" + rname)

	//查询
	grp.GET("/", func(c *gin.Context) {
		if before != nil {
			before(c)
		}
		r, _ := c.GetQuery("range")
		if r == "" {
			r = "page"
		}
		sort, _ := c.GetQuery("sort")
		if sort == "" {
			sort = "-updatetime -createtime"
		}
		cond, _ := c.GetQuery("cond")
		fileds, _ := c.GetQuery("fields")
		idx, size := 0, 0
		if strings.ToLower(r) == "page" {
			pageIndex, _ := c.GetQuery("page")
			pageSize, _ := c.GetQuery("size")
			idx, _ = strconv.Atoi(pageIndex)
			if idx == 0 {
				idx = 1
			}
			size, err = strconv.Atoi(pageSize)
			if err != nil {
				size = 10
			}
		} else if strings.ToLower(r) == "one" {
			size = 1
			idx = 1
		}
		datas, total, err := m.Query(robj, cond, idx, size, sort, fileds, false)
		tp := 0.0
		if size != 0 {
			tp = math.Ceil((float64)(total) / (float64)(size))
		}

		if after != nil {
			err = after(c, &datas, err)
		}
		retData := gin.H{
			"range":        r,
			"sort":         sort,
			"size":         size,
			"list":         datas,
			"totalrecords": total,
			"totalpages":   tp,
			"page":         idx}
		Ret(c, retData, err, 500)
	})
	//ID查询
	grp.GET("/id", func(c *gin.Context) {
		if before != nil {
			before(c)
		}
		val, _ := c.GetQuery("val")
		data, err := m.FindID(robj, val)
		if after != nil {
			err = after(c, &data, err)
		}
		Ret(c, data, err, 500)
	})
	//新增
	grp.POST("/", func(c *gin.Context) {
		if before != nil {
			before(c)
		}
		body, _ := ioutil.ReadAll(c.Request.Body)
		data := (string)(body)
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
				u := auth.CurUser(c)
				err := m.Insert(obj, u)
				if after != nil {
					err = after(c, nil, err)
				}
				Ret(c, obj, err, 500)
			} else {
				if after != nil {
					err = after(c, nil, err)
				}
				Ret(c, nil, errors.New("数据已存在，查询条件："+exist), 500)
			}
		}

	})
	//修改
	grp.PUT("/", func(c *gin.Context) {
		if before != nil {
			before(c)
		}
		cond := c.PostForm("cond")
		doc := c.PostForm("doc")
		multi := c.PostForm("multi")
		b, err := strconv.ParseBool(multi)
		if err != nil {
			b = false
		}
		user := auth.CurUser(c)
		info, err := m.Update(robj, cond, doc, user, b)
		if after != nil {
			err = after(c, info, err)
		}
		Ret(c, info, err, 500)
	})
	//删除
	grp.DELETE("/", func(c *gin.Context) {
		if before != nil {
			before(c)
		}
		cond := c.PostForm("cond")
		multi := c.PostForm("multi")
		b, err := strconv.ParseBool(multi)
		if err != nil {
			b = false
		}
		user := auth.CurUser(c)
		info, err := m.Remove(robj, cond, user, b)
		if after != nil {
			err = after(c, info, err)
		}
		Ret(c, info, err, 500)
	})
}

//Ret 返回值
func Ret(c *gin.Context, data interface{}, err error, code int) {
	if err != nil {
		c.JSON(200, gin.H{"errcode": code, "errmsg": err.Error()})
	} else {
		c.JSON(200, gin.H{"data": data})
	}
}
