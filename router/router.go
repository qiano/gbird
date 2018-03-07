package router

import (
	"encoding/json"
	"errors"
	"gbird"
	"gbird/logger"
	"gbird/model"
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
func Register(r *gbird.App, robj interface{}, beforeHandler func(c *gbird.Context, data interface{}) error, afterHandler func(c *gbird.Context, data *gbird.H, err error) error) {
	model.RegisterMetadata(robj)

	tagval, err := model.MTagVal(robj, "router")
	if err != nil {
		logger.Infoln(reflect.TypeOf(robj).String() + "未指定 router ")
		return
	}
	rs := strings.Split(tagval, " ")
	rname := rs[0]
	if len(rs) != 5 && len(rs) != 1 {
		panic(errors.New(reflect.TypeOf(robj).String() + "router设置异常"))
	}
	var (
		get    = true
		post   = true
		put    = true
		delete = true
	)
	if len(rs) == 5 {
		get = rs[1] == "1"
		post = rs[2] == "1"
		put = rs[3] == "1"
		delete = rs[4] == "1"
	}
	grp := r.Engine.Group("/api/" + rname)
	if get {
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
			if beforeHandler != nil {
				beforeHandler(&gbird.Context{Context: c}, nil)
			}
			datas, total, err := m.Query(robj, cond, idx, size, sort, fileds, false)
			tp := 0.0
			if size != 0 {
				tp = math.Ceil((float64)(total) / (float64)(size))
			}

			retData := gin.H{"data": gin.H{
				"range":        r,
				"sort":         sort,
				"size":         size,
				"list":         datas,
				"totalrecords": total,
				"totalpages":   tp,
				"page":         idx}}
			h := (gbird.H)(retData)
			if afterHandler != nil {
				err = afterHandler(&gbird.Context{Context: c}, &h, err)
			}
			gbird.Ret(&gbird.Context{Context: c}, h, err, 500)
		})

		//ID查询
		grp.GET("/id", func(c *gin.Context) {
			val, _ := c.GetQuery("val")
			if beforeHandler != nil {
				err := beforeHandler(&gbird.Context{Context: c}, val)
				if err != nil {

					gbird.Ret(&gbird.Context{Context: c}, nil, err, 500)
					return
				}
			}
			data, err := m.FindID(robj, bson.ObjectIdHex(val))
			retdata := gbird.H{"data": data}
			if afterHandler != nil {
				err = afterHandler(&gbird.Context{Context: c}, &retdata, err)
			}
			gbird.Ret(&gbird.Context{Context: c}, retdata, err, 500)
		})
	}
	if post {
		//新增
		grp.POST("", func(c *gin.Context) {
			body, _ := ioutil.ReadAll(c.Request.Body)
			data := (string)(body)
			objType := reflect.TypeOf(robj).Elem()
			obj := reflect.New(objType).Interface()
			json.Unmarshal([]byte(data), &obj)
			var uid string
			if gbird.GetCurUser != nil {
				user, err := gbird.GetCurUser(&gbird.Context{Context: c})
				if err == nil {
					uid = user.UserID()
				}
			}
			if beforeHandler != nil {
				err := beforeHandler(&gbird.Context{Context: c}, obj)
				if err != nil {
					gbird.Ret(&gbird.Context{Context: c}, nil, err, 500)
					return
				}
			}
			err = m.Insert(obj, uid)

			retdata := gbird.H{"data": obj}
			if afterHandler != nil {
				err = afterHandler(&gbird.Context{Context: c}, &retdata, err)
			}
			gbird.Ret(&gbird.Context{Context: c}, retdata, err, 500)
		})
	}
	if put {
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
			if gbird.GetCurUser != nil {
				user, err := gbird.GetCurUser(&gbird.Context{Context: c})
				if err == nil {
					uid = user.UserID()
				}
			}
			if beforeHandler != nil {
				err := beforeHandler(&gbird.Context{Context: c}, nil)
				if err != nil {
					gbird.Ret(&gbird.Context{Context: c}, nil, err, 500)
					return
				}
			}
			info, err := m.Update(robj, cond, doc, uid, b)
			retdata := gbird.H{"data": info, "cond": cond, "multi": b}
			if afterHandler != nil {
				err = afterHandler(&gbird.Context{Context: c}, &retdata, err)
			}
			gbird.Ret(&gbird.Context{Context: c}, retdata, err, 500)
		})
	}
	if delete {
		//删除
		grp.DELETE("", func(c *gin.Context) {
			cond := c.PostForm("cond")
			multi := c.PostForm("multi")
			b, err := strconv.ParseBool(multi)
			if err != nil {
				b = false
			}
			var uid string
			if gbird.GetCurUser != nil {
				user, err := gbird.GetCurUser(&gbird.Context{Context: c})
				if err == nil {
					uid = user.UserID()
				}
			}
			if beforeHandler != nil {
				err := beforeHandler(&gbird.Context{Context: c}, nil)
				if err != nil {
					gbird.Ret(&gbird.Context{Context: c}, nil, err, 500)
					return
				}
			}
			info, err := m.Remove(robj, cond, uid, b)
			retdata := gbird.H{"data": info, "cond": cond, "multi": b}
			if afterHandler != nil {
				err = afterHandler(&gbird.Context{Context: c}, &retdata, err)
			}
			gbird.Ret(&gbird.Context{Context: c}, retdata, err, 500)
		})
	}
}
