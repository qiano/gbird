package gbird

import (
	"encoding/json"
	"errors"
	"gbird/auth"
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
func (r *App) Register(robj interface{}) {
	rname, err := getRouterName(robj)
	if err != nil {
		panic(err)
	}
	soles := getSoles(robj)
	grp := r.Group("/api/" + rname)

	grp.GET("/", func(c *gin.Context) {
		r, _ := c.GetQuery("range")
		if r == "" {
			r = "page"
		}
		sort, _ := c.GetQuery("sort")
		if sort == "" {
			sort = "-updatetime -createtime"
		}
		cond, _ := c.GetQuery("cond")
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
		datas, total, err := m.Query(robj, cond, idx, size, sort, false)
		tp := 0.0
		if size != 0 {
			tp = math.Ceil((float64)(total) / (float64)(size))
		}
		Ret(c, gin.H{
			"range":        r,
			"sort":         sort,
			"size":         size,
			"list":         datas,
			"totalrecords": total,
			"totalpages":   tp,
			"page":         idx}, err, 500)

		// if err != nil {
		// 	c.JSON(200, gin.H{"errcode": 500, "errmsg": err})
		// } else {
		// 	c.JSON(200, gin.H{"data": gin.H{
		// 		"range":        r,
		// 		"sort":         sort,
		// 		"size":         size,
		// 		"list":         datas,
		// 		"totalrecords": total,
		// 		"totalpages":   total / size,
		// 		"page":         idx}})
		// }
	})
	// logger.Infoln("路由注册：GET" + "  /api/" + rname)

	grp.GET("/id", func(c *gin.Context) {
		val, _ := c.GetQuery("val")
		data, err := m.FindID(robj, val)
		Ret(c, data, err, 500)
		// if err != nil {
		// 	c.JSON(200, gin.H{"errcode": 500, "errmsg": err})
		// } else {
		// 	c.JSON(200, gin.H{"code": 0, "data": data})
		// }
	})

	// logger.Infoln("路由注册：GET" + "  /api/" + rname + "/:id")

	grp.POST("/", func(c *gin.Context) {
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
				Ret(c, obj, err, 500)
				// if err != nil {
				// 	c.JSON(500, gin.H{"errmsg": err.Error()})
				// } else {
				// 	c.JSON(200, gin.H{"data": obj})
				// }
			} else {
				Ret(c, nil, errors.New("数据已存在，查询条件："+exist), 500)
				// c.JSON(200, gin.H{"errmsg": "数据已存在，查询条件：" + exist})
			}
		}

	})
	// logger.Infoln("路由注册：POST" + " /api/" + rname + "/insert")

	grp.PUT("/", func(c *gin.Context) {
		cond := c.PostForm("cond")
		doc := c.PostForm("doc")
		multi := c.PostForm("multi")
		b, err := strconv.ParseBool(multi)
		if err != nil {
			b = false
		}
		user := auth.CurUser(c)
		info, err := m.Update(robj, cond, doc, user, b)
		Ret(c, info, err, 500)
		// if err != nil {
		// 	c.JSON(500, gin.H{"errmsg": err.Error()})
		// } else {
		// 	c.JSON(200, gin.H{"data": info})
		// }
	})
	// logger.Infoln("路由注册：PUT" + " /api/" + rname)

	grp.DELETE("/", func(c *gin.Context) {
		cond := c.PostForm("cond")
		multi := c.PostForm("multi")
		b, err := strconv.ParseBool(multi)
		if err != nil {
			b = false
		}
		user := auth.CurUser(c)
		info, err := m.Remove(robj, cond, user, b)
		Ret(c, info, err, 500)
		// if err != nil {
		// 	c.JSON(500, gin.H{"errmsg": err.Error(), "data": info})
		// } else {
		// 	c.JSON(200, gin.H{"data": info})
		// }
	})
	// logger.Infoln("路由注册：DELETE" + " /api/" + rname)

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

//Ret 返回值
func Ret(c *gin.Context, data interface{}, err error, code int) {
	if err != nil {
		c.JSON(200, gin.H{"errcode": code, "errmsg": err})
	} else {
		c.JSON(200, gin.H{"data": data})
	}
}

//ToSlice to slice
func ToSlice(arr interface{}) []interface{} {
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		panic("toslice arr not slice")
	}
	l := v.Len()
	ret := make([]interface{}, l)
	for i := 0; i < l; i++ {
		ret[i] = v.Index(i).Interface()
	}
	return ret
}
