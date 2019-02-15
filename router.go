package gbird

import (
	"encoding/json"
	"errors"
	"github.com/qiano/gbird/logger"
	"github.com/qiano/gbird/model"
	m "github.com/qiano/gbird/mongodb"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"math"
	"reflect"
	"strconv"
	"strings"
)

//Register 模型注册
func (r *App) Register(robj interface{}, beforeHandler func(c *Context, data interface{}) error, afterHandler func(c *Context, data interface{}, err error) (interface{}, error)) {
	model.RegisterMetadata(robj)
	tagval, err := model.MTagVal(robj, "router")
	if err != nil {
		logger.Infoln(reflect.TypeOf(robj).String() + "未指定 router ")
		return
	}
	rs := strings.Split(tagval, " ")
	rname := rs[0]
	if len(rs) != 2 && len(rs) != 1 {
		panic(errors.New(reflect.TypeOf(robj).String() + "router设置异常"))
	}
	var (
		get    = false
		post   = false
		put    = false
		delete = false
	)
	if len(rs) == 2 {
		sets := rs[1]
		if len(sets) == 4 {
			get = sets[0:1] == "1"
			post = sets[1:2] == "1"
			put = sets[2:3] == "1"
			delete = sets[3:4] == "1"
		}
	}
	path := "/api/" + rname
	grp := r.Engine.Group(path)
	if get {
		//查询
		RegisterAPIPermission(path, "GET")
		grp.GET("", func(c *gin.Context) {
			r, _ := c.GetQuery("range")
			if r == "" {
				r = "page"
			}
			sort, _ := c.GetQuery("sort")
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
				beforeHandler(&Context{Context: c}, nil)
			}
			gc := Context{Context: c}

			qi, err := m.ToBson(cond)
			if err != nil {
				gc.RetError(err)
				return
			}
			datas, total, err := m.Query(robj, qi, idx, size, sort, fileds, false)
			tp := 0.0
			if size != 0 {
				tp = math.Ceil((float64)(total) / (float64)(size))
			}

			retData := H{
				"range":        r,
				"sort":         sort,
				"size":         size,
				"list":         datas,
				"totalrecords": total,
				"totalpages":   tp,
				"page":         idx}
			var ret interface{} = retData
			if afterHandler != nil {
				ret, err = afterHandler(&Context{Context: c}, &retData, err)
			}
			if err != nil {
				gc.RetError(err)
			} else {
				gc.Ret(ret)
			}
		})

		//ID查询
		RegisterAPIPermission(path+"/id", "GET")
		grp.GET("/id", func(c *gin.Context) {
			gc := Context{Context: c}
			val, _ := c.GetQuery("val")
			if beforeHandler != nil {
				err := beforeHandler(&Context{Context: c}, val)
				if err != nil {
					gc.RetError(err)
					return
				}
			}
			data, err := m.FindID(robj, bson.ObjectIdHex(val))
			var ret interface{} = data
			if afterHandler != nil {
				ret, err = afterHandler(&Context{Context: c}, &data, err)
			}
			if err != nil {
				gc.RetError(err)
			} else {
				gc.Ret(ret)
			}
		})
	}
	if post {
		//新增
		RegisterAPIPermission(path, "POST")
		grp.POST("", func(c *gin.Context) {
			body, _ := ioutil.ReadAll(c.Request.Body)
			data := (string)(body)
			objType := reflect.TypeOf(robj).Elem()
			obj := reflect.New(objType).Interface()
			json.Unmarshal([]byte(data), &obj)
			var uid string
			if GetCurUser != nil {
				user, err := GetCurUser(&Context{Context: c})
				if err == nil {
					uid = user["ID"].(string)
				}
			}
			if beforeHandler != nil {
				err := beforeHandler(&Context{Context: c}, obj)
				if err != nil {
					gc := Context{Context: c}
					gc.RetError(err)
					return
				}
			}
			err = m.Insert(obj, uid)
			if afterHandler != nil {
				obj, err = afterHandler(&Context{Context: c}, obj, err)
			}
			gc := Context{Context: c}
			if err != nil {
				gc.RetError(err)
			} else {
				gc.Ret(obj)
			}
		})
	}
	if put {
		//修改
		RegisterAPIPermission(path, "PUT")
		grp.PUT("", func(c *gin.Context) {
			gc := Context{Context: c}
			cond := c.PostForm("cond")
			doc := c.PostForm("doc")
			multi := c.PostForm("multi")
			b, err := strconv.ParseBool(multi)
			if err != nil {
				b = false
			}
			var uid string
			if GetCurUser != nil {
				user, err := GetCurUser(&Context{Context: c})
				if err == nil {
					uid = user["ID"].(string)
				}
			}
			if beforeHandler != nil {
				err := beforeHandler(&Context{Context: c}, nil)
				if err != nil {
					gc.RetError(err)
					return
				}
			}
			qu, err := m.ToBson(cond)
			if err != nil {
				gc.RetError(err)
				return
			}
			do, err := m.ToBson(doc)
			if err != nil {
				gc.RetError(err)
				return
			}
			info, err := m.Update(robj, qu, do, uid, b)
			retdata := H{"result": info, "cond": cond, "multi": b}
			var ret interface{} = retdata
			if afterHandler != nil {
				ret, err = afterHandler(&Context{Context: c}, &retdata, err)
			}
			if err != nil {
				gc.RetError(err)
			} else {
				gc.Ret(ret)
			}
		})
	}
	if delete {
		//删除
		RegisterAPIPermission(path, "DELETE")
		grp.DELETE("", func(c *gin.Context) {
			gc := Context{Context: c}
			cond := c.PostForm("cond")
			multi := c.PostForm("multi")
			b, err := strconv.ParseBool(multi)
			if err != nil {
				b = false
			}
			var uid string
			if GetCurUser != nil {
				user, err := GetCurUser(&Context{Context: c})
				if err == nil {
					uid = user["ID"].(string)
				}
			}
			if beforeHandler != nil {
				err := beforeHandler(&Context{Context: c}, nil)
				if err != nil {
					gc.RetError(err)
					return
				}
			}
			qu, err := m.ToBson(cond)
			if err != nil {
				gc.RetError(err)
				return
			}
			info, err := m.Remove(robj, qu, uid, b)
			retdata := H{"data": info, "cond": cond, "multi": b}
			var ret interface{} = retdata
			if afterHandler != nil {
				ret, err = afterHandler(&Context{Context: c}, &retdata, err)
			}
			if err != nil {
				gc.RetError(err)
			} else {
				gc.Ret(ret)
			}
		})
	}
}
