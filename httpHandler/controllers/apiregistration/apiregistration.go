// Package apiregistration wires controller route declarations into Gin router groups.
//
// # URL assembly
//
// For every entry in a controller's GetApisMap(), the final Gin path is:
//
//	config.ServiceLBName + config.ApiPrefix + "/" + Apis.ApiName
//
// For example, with ServiceLBName = "/api/v1", ApiPrefix = "/users", and
// ApiName = "profile/:id", the registered path is /api/v1/users/profile/:id.
//
// # Tier resolution
//
// The map key (customtypes.RouterNames) must match one of the tier names
// registered in GINServer.Setup().  If no matching group is found the route
// is skipped and RegisterApis returns an error for that tier so the caller
// can surface the misconfiguration clearly at startup.
package apiregistration

import (
	"fmt"
	"strings"
	"sync"

	"github.com/glodb/keel/httpHandler/baserouter"
	"github.com/glodb/keel/internal/openapi"
	"github.com/glodb/keel/models/genericmodels"
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

// RegisterApis registers every route in apiMap with Gin and records it in the
// internal route registry used by middleware for method-matching.
//
// It returns one error per tier whose name is not found in the active
// BaseRouterHandler (i.e. was not declared in GINServer.Setup()).  Routes in
// valid tiers are still registered even when other tiers fail, so callers
// receive the full list of problems rather than stopping at the first one.
func (a *ApiRegistration) RegisterApis(apiMap map[customtypes.RouterNames][]genericmodels.Apis) []error {
	registryMu.Lock()
	defer registryMu.Unlock()

	var errs []error

	for name, apis := range apiMap {
		router, err := baserouter.GetActive().GetRouter(name)
		if err != nil {
			routeErr := fmt.Errorf(
				"RegisterApis: tier %q not found — did you pass it to GINServer.Setup()? "+
					"Valid tiers are the keys of the middlewares map you supplied to Setup(). "+
					"Routes for this tier were NOT registered: %w",
				name, err,
			)
			logger.Log().Error("Route registration failed — unknown tier",
				logger.StringField("tier", string(name)),
				logger.ErrorField("error", routeErr),
			)
			errs = append(errs, routeErr)
			continue
		}

		for _, api := range apis {
			// Strip path parameters to get the base route name used as the
			// registry key (e.g. "users/:id" → "users").
			routeName := strings.Split(api.ApiName, "/:")[0]

			// Final Gin path = ServiceLBName + ApiPrefix + "/" + ApiName
			// e.g. "/api/v1" + "/users" + "/" + "profile/:id" = /api/v1/users/profile/:id
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

			if api.Doc != nil {
				openapi.GetInstance().RegisterEndpoint(string(api.ApiMethod), apiName, api.Doc)
			}
		}
	}

	return errs
}
