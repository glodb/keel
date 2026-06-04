package baserouter

import "github.com/gin-gonic/gin"

type BaseRouter interface {
	SetRouter(name string, router *gin.RouterGroup)
	GetBaseRouter(secret string) *gin.RouterGroup
	GetOpenRouter() *gin.RouterGroup
	GetLoginRouter() *gin.RouterGroup
}