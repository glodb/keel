package controllers

import (
	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/httpHandler/controllers/apiregistration"
	"github.com/glodb/keel/httpHandler/controllers/baseinterfaces"
	"github.com/glodb/keel/settings/configmanager"
)

type BaseController struct {
	baseinterfaces.BaseControllerFactory
	baseinterfaces.BaseFunctionsInterface
	apiregistration.ApiRegistration
}

func (bc *BaseController) GetDBName() basetypes.DBName {
	return basetypes.DBName(configmanager.GetInstance().Mongo.DBName)
}

func (b *BaseController) SetDependencies(f baseinterfaces.BaseFunctionsInterface, c baseinterfaces.BaseControllerFactory) {
	b.BaseFunctionsInterface = f
	b.BaseControllerFactory = c
}
