package basemiddlewares

import "github.com/gin-gonic/gin"

type Middleware interface {
	GetHandlerFunc() gin.HandlerFunc
}
