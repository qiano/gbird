package gbird

import (
	"github.com/gin-gonic/gin"

)

//H h
type H gin.H

//Context 上下文
type Context struct {
	*gin.Context
}

//Use use
func (a *App) Use(middleware ...func(*Context)) {
	a.Engine.Use(BirdToGin(middleware...)...)
}

//BirdToGin 类型转换
func BirdToGin(handlers ...func(c *Context)) []gin.HandlerFunc {
	ginHandlers := make([]gin.HandlerFunc, 0, 0)
	for _, handler := range handlers {
		ginHandlers = append(ginHandlers, func(ginc *gin.Context) {
			handler(&Context{Context: ginc})
		})
	}
	return ginHandlers
}

//GinToBird 类型转换
func GinToBird(handler func(c *gin.Context)) func(*Context) {
	return func(gc *Context) {
		handler(gc.Context)
	}
}

//POST post
func (a *App) POST(relativePath string, handlers ...func(c *Context)) {
	a.Engine.POST(relativePath, BirdToGin(handlers...)...)
}

//GET get
func (a *App) GET(relativePath string, handlers ...func(c *Context)) {
	a.Engine.GET(relativePath, BirdToGin(handlers...)...)
}


