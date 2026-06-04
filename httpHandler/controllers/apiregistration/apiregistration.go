package apiregistration

import (
	"strings"
	"sync"

	"github.com/glodb/keel/models/genericmodels"
	"github.com/glodb/keel/httpHandler/baserouter"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/customtypes"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/utilsdatatypes"
)

var (
	registryMu sync.RWMutex
	registry   = map[string]*utilsdatatypes.Set[string]{}
)

// RegisteredApis returns a snapshot of the registered API route → HTTP-method sets.
// Callers (e.g. keel-code's apimiddleware) should call this once during startup and
// cache the result, or call it per-request (the map is safe for concurrent reads after
// all RegisterApis calls have completed).
func RegisteredApis() map[string]*utilsdatatypes.Set[string] {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make(map[string]*utilsdatatypes.Set[string], len(registry))
	for k, v := range registry {
		out[k] = v
	}
	return out
}

type ApiRegistration struct{}

func (a *ApiRegistration) RegisterApis(apiMap map[customtypes.RouterNames][]genericmodels.Apis) {
	registryMu.Lock()
	defer registryMu.Unlock()

	for name, apis := range apiMap {
		router, err := baserouter.GetInstance().GetRouter(name)
		if err != nil {
			logger.Log().Error("Failed to get router", logger.ErrorField("error", err), logger.StringField("name", string(name)))
			continue
		}
		for _, api := range apis {
			routeName := strings.Split(api.ApiName, "/:")[0]
			configApiName := configmanager.GetInstance().ApiPrefix + "/" + routeName
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

			if data, ok := registry[apiName]; ok {
				data.Add(string(api.ApiMethod))
			} else {
				s := utilsdatatypes.NewSet[string]()
				s.Add(string(api.ApiMethod))
				registry[configApiName] = s
			}
		}
	}
}
