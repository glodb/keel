package genericmodels

import (
	"github.com/glodb/keel/settings/customtypes"

	"github.com/gin-gonic/gin"
)

type Apis struct {
	ApiName   string               `json:"apiName"`
	ApiMethod customtypes.ApiTypes `json:"apiMethod"`
	Method    gin.HandlerFunc      `json:"method"`
}
