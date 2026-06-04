package basemodels

type MigrationModels interface {
	GetQuery() map[string]interface{}
	GetUpdate() map[string]interface{}
	MapData(map[string]interface{})
	CleanUp()
}
