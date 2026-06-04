package baseinterfaces

import (
	"github.com/glodb/keel/app/models/genericmodels"
	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/settings/customtypes"
)

type Controller interface {
	BaseFunctionsInterface
	BaseControllerFactory
	SetDependencies(BaseFunctionsInterface)
	GetCollectionName() basetypes.CollectionName
	GetApisMap() map[customtypes.RouterNames][]genericmodels.Apis
	RegisterApis(apiMap map[customtypes.RouterNames][]genericmodels.Apis)
	Initialize() error
	GetDBName() basetypes.DBName
}
