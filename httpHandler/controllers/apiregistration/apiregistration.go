package apiregistration

import (
	"strings"

	"github.com/glodb/keel/app/models/genericmodels"
	"github.com/glodb/keel/httpHandler/baserouter"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/customtypes"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/utilsdatatypes"
)

type ApiRegistration struct{}

func (a *ApiRegistration) RegisterApis(apiMap map[customtypes.RouterNames][]genericmodels.Apis) {
	for name, apis := range apiMap {
		router, err := baserouter.GetInstance().GetRouter(name)
		if err != nil {
			logger.Log().Error("Failed to get router", logger.ErrorField("error", err), logger.StringField("name", string(name)))
			continue
		}
		for _, api := range apis {
			name := strings.Split(api.ApiName, "/:")[0]
			connfigApiName := configmanager.GetInstance().ApiPrefix + "/" + name
			apiName := configmanager.GetInstance().ApiPrefix + "/" + api.ApiName
			apiLBName := configmanager.GetInstance().ServiceLBName + apiName
			switch api.ApiMethod {
			case customtypes.POST:
				router.POST(apiLBName, api.Method)
			case customtypes.GET:
				router.GET(apiLBName, api.Method)
			case customtypes.PUT:
				router.PUT(apiLBName, api.Method)
			case customtypes.DELETE:
				router.DELETE(apiLBName, api.Method)
			}
			if data, ok := configmanager.GetInstance().Apis[apiName]; ok {
				data.Add(string(api.ApiMethod))
			} else {
				configmanager.GetInstance().Apis[connfigApiName] = utilsdatatypes.NewSet[string]()
				configmanager.GetInstance().Apis[connfigApiName].Add(string(api.ApiMethod))
			}
		}
	}
}
