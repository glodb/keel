package docinterface

import "github.com/glodb/keel/settings/openapi"

type DocInterface interface {
	RegisterDocs(*openapi.ControllerHelper)
}
