package gbird

import (
	"encoding/json"
	"gbird/auth"
	"gbird/base"
	m "gbird/mongodb"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"math"
	"reflect"
	"strconv"
	"strings"
)

//Register 模型注册
func (r *App) Register(robj interface{}, before func(c *gin.Context, data interface{}) error, after func(*gin.Context, interface{}, error) error) {
	base.RegisterMetadata(robj)
	rname, _, err := base.FindTag(robj, "urlname", "")
	if err != nil {
		panic(err)
	}

	grp := r.Group("/api/" + rname)

	//查询
	grp.GET("", func(c *gin.Context) {
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
		if before != nil {
			before(c, nil)
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
		val, _ := c.GetQuery("val")
		if before != nil {
			err := before(c, val)
			if err != nil {
				Ret(c, nil, err, 500)
				return
			}
		}
		data, err := m.FindID(robj, bson.ObjectIdHex(val))
		if after != nil {
			err = after(c, &data, err)
		}
		Ret(c, data, err, 500)
	})
	//新增
	grp.POST("", func(c *gin.Context) {
		body, _ := ioutil.ReadAll(c.Request.Body)
		data := (string)(body)
		objType := reflect.TypeOf(robj).Elem()
		obj := reflect.New(objType).Interface()
		json.Unmarshal([]byte(data), &obj)
		var uid string
		if auth.GetCurUserIDName != nil {
			uid,_ = auth.GetCurUserIDName(c)
		}
		if before != nil {
			err := before(c, obj)
			if err != nil {
				Ret(c, nil, err, 500)
				return
			}
		}
		err := m.Insert(obj, uid)

		if after != nil {
			err = after(c, nil, err)
		}
		Ret(c, obj, err, 500)
	})
	//修改
	grp.PUT("", func(c *gin.Context) {
		cond := c.PostForm("cond")
		doc := c.PostForm("doc")
		multi := c.PostForm("multi")
		b, err := strconv.ParseBool(multi)
		if err != nil {
			b = false
		}
		var uid string
		if auth.GetCurUserIDName != nil {
			uid,_ = auth.GetCurUserIDName(c)
		}
		if before != nil {
			err := before(c, nil)
			if err != nil {
				Ret(c, nil, err, 500)
				return
			}
		}
		info, err := m.Update(robj, cond, doc, uid, b)
		if after != nil {
			err = after(c, info, err)
		}
		Ret(c, info, err, 500)
	})
	//删除
	grp.DELETE("", func(c *gin.Context) {
		cond := c.PostForm("cond")
		multi := c.PostForm("multi")
		b, err := strconv.ParseBool(multi)
		if err != nil {
			b = false
		}
		var uid string
		if auth.GetCurUserIDName != nil {
			uid,_ = auth.GetCurUserIDName(c)
		}
		if before != nil {
			err:=before(c, nil)
			if err != nil {
				Ret(c, nil, err, 500)
				return
			}
		}
		info, err := m.Remove(robj, cond, uid, b)
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
