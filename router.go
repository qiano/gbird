package gbird

import (
	"encoding/json"
	"gbird/module/auth"
	"gbird/model"
	m "gbird/mongodb"
	"gbird/module/logger"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"math"
	"reflect"
	"strconv"
	"strings"
)

//RegisterOptins 模型注册选项
type RegisterOptins struct {
	BeforeHandler func(c *gin.Context, data interface{}) error
	AfterHandler  func(c *gin.Context, data *gin.H, err error) error
	RouterEnabled int
	GET           int
	GETID         int
	POST          int
	PUT           int
	DELETE        int
}

//Register 模型注册
func (r *App) Register(robj interface{}, options *RegisterOptins) {
	model.RegisterMetadata(robj)
	rname, err := model.MTagVal(robj, "urlname")
	if err != nil || rname == "" {
		logger.Infoln(reflect.TypeOf(robj).String() + "未指定 urlname ，不生成默认路由")
		return
	}
	//初始化options
	if options == nil {
		options = &RegisterOptins{
			BeforeHandler: nil,
			AfterHandler:  nil,
			RouterEnabled: 0,
			GET:           0,
			GETID:         0,
			POST:          0,
			PUT:           0,
			DELETE:        0,
		}
	}
	if options.RouterEnabled < 0 {
		return
	}
	grp := r.Group("/api/" + rname)
	if options.GET >= 0 {
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
			if options.BeforeHandler != nil {
				options.BeforeHandler(c, nil)
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

			if options.AfterHandler != nil {
				err = options.AfterHandler(c, &retData, err)
			}
			Ret(c, retData, err, 500)
		})
	}
	if options.GETID >= 0 {
		//ID查询
		grp.GET("/id", func(c *gin.Context) {
			val, _ := c.GetQuery("val")
			if options.BeforeHandler != nil {
				err := options.BeforeHandler(c, val)
				if err != nil {
					Ret(c, nil, err, 500)
					return
				}
			}
			data, err := m.FindID(robj, bson.ObjectIdHex(val))
			retdata := gin.H{"data": data}
			if options.AfterHandler != nil {
				err = options.AfterHandler(c, &retdata, err)
			}
			Ret(c, retdata, err, 500)
		})
	}
	if options.POST >= 0 {
		//新增
		grp.POST("", func(c *gin.Context) {
			body, _ := ioutil.ReadAll(c.Request.Body)
			data := (string)(body)
			objType := reflect.TypeOf(robj).Elem()
			obj := reflect.New(objType).Interface()
			json.Unmarshal([]byte(data), &obj)
			var uid string
			if auth.GetCurUser != nil {
				user, err := auth.GetCurUser(c)
				if err == nil {
					uid = user.UserID()
				}
			}
			if options.BeforeHandler != nil {
				err := options.BeforeHandler(c, obj)
				if err != nil {
					Ret(c, nil, err, 500)
					return
				}
			}
			err = m.Insert(obj, uid)

			retdata := gin.H{"data": obj}
			if options.AfterHandler != nil {
				err = options.AfterHandler(c, &retdata, err)
			}
			Ret(c, retdata, err, 500)
		})
	}
	if options.PUT >= 0 {
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
			if auth.GetCurUser != nil {
				user, err := auth.GetCurUser(c)
				if err == nil {
					uid = user.UserID()
				}
			}
			if options.BeforeHandler != nil {
				err := options.BeforeHandler(c, nil)
				if err != nil {
					Ret(c, nil, err, 500)
					return
				}
			}
			info, err := m.Update(robj, cond, doc, uid, b)
			retdata := gin.H{"data": info}
			if options.AfterHandler != nil {
				err = options.AfterHandler(c, &retdata, err)
			}
			Ret(c, retdata, err, 500)
		})
	}
	if options.DELETE >= 0 {
		//删除
		grp.DELETE("", func(c *gin.Context) {
			cond := c.PostForm("cond")
			multi := c.PostForm("multi")
			b, err := strconv.ParseBool(multi)
			if err != nil {
				b = false
			}
			var uid string
			if auth.GetCurUser != nil {
				user, err := auth.GetCurUser(c)
				if err == nil {
					uid = user.UserID()
				}
			}
			if options.BeforeHandler != nil {
				err := options.BeforeHandler(c, nil)
				if err != nil {
					Ret(c, nil, err, 500)
					return
				}
			}
			info, err := m.Remove(robj, cond, uid, b)
			retdata := gin.H{"data": info}
			if options.AfterHandler != nil {
				err = options.AfterHandler(c, &retdata, err)
			}
			Ret(c, retdata, err, 500)
		})
	}
}

//Ret 返回值
func Ret(c *gin.Context, data gin.H, err error, code int) {
	if err != nil {
		c.JSON(200, gin.H{"errcode": code, "errmsg": err.Error()})
	} else {
		c.JSON(200, data)
	}
}
