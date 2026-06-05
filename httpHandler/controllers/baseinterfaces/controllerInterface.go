package baseinterfaces

import (
	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/models/genericmodels"
	"github.com/glodb/keel/settings/customtypes"
)

// Controller is the interface every keel controller must implement.
//
// # Routing overview
//
// GetApisMap returns routes grouped by "tier" — a named Gin router group
// created in GINServer.Setup().  The tier key must exactly match one of the
// keys you passed to Setup(); a mismatch causes that route to be skipped and
// an error to be returned from RegisterApis.
//
// The final URL registered with Gin is assembled as:
//
//	config.ServiceLBName + config.ApiPrefix + "/" + Apis.ApiName
//
// Example: ServiceLBName="/api/v1", ApiPrefix="/users", ApiName="profile/:id"
// registers the path  /api/v1/users/profile/:id  inside the chosen tier group.
type Controller interface {
	BaseFunctionsInterface
	BaseControllerFactory
	SetDependencies(BaseFunctionsInterface, BaseControllerFactory)
	GetCollectionName() basetypes.CollectionName

	// GetApisMap returns the routes this controller owns, keyed by tier name.
	// See the Controller doc and customtypes.RouterNames for the full contract.
	GetApisMap() map[customtypes.RouterNames][]genericmodels.Apis

	// RegisterApis registers the provided routes with Gin.  It returns one
	// error per route whose tier name did not match any group registered in
	// GINServer.Setup(); all other routes are still registered normally.
	RegisterApis(apiMap map[customtypes.RouterNames][]genericmodels.Apis) []error

	Initialize() error
	GetDBName() basetypes.DBName
}
