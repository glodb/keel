package baseinterfaces

import (
	"context"
	"database/sql"

	"github.com/glodb/keel/models/genericmodels"
	"github.com/glodb/keel/settings/customtypes"

	"go.mongodb.org/mongo-driver/mongo"
)

/*
* flyweight interface to separate
* different types of connections
* with db functionality
 */
type BaseFunctionsInterface interface {
	GetFunctions() BaseFunctionsInterface
	EnsureIndex(ctx context.Context, controller Controller, data interface{}, unique bool, partialFilterExpression *customtypes.M) error
	Add(ctx context.Context, controller Controller, data interface{}, scan bool) (int64, error)
	AddMany(ctx context.Context, controller Controller, data []interface{}, scan bool) ([]int64, error)
	SqlFind(ctx context.Context, controller Controller, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool) (*sql.Rows, error)
	Update(ctx context.Context, controller Controller, query string, data []interface{}, upsert bool) error
	Delete(ctx context.Context, controller Controller, condition map[string]interface{}, useOr bool, addParenthesis bool) error
	SqlPaginate(ctx context.Context, controller Controller, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool, pageSize int, page int) (*sql.Rows, int64, error)
	Count(ctx context.Context, controller Controller, condition map[string]interface{}, useOr bool) (int64, error)
	RawQuery(ctx context.Context, controller Controller, query string, data []interface{}) (*sql.Rows, error)

	//For mongo mainly
	UpdateOne(ctx context.Context, controller Controller, query customtypes.M, data customtypes.M, upsert bool) error
	UpdateMany(ctx context.Context, controller Controller, query customtypes.M, data customtypes.M, upsert bool) (int64, error)
	FindOneAndUpdate(ctx context.Context, controller Controller, query customtypes.M, data customtypes.M, upsert bool, returnNew bool, result interface{}, option FunctionOptions) error
	FindOneAndDelete(ctx context.Context, controller Controller, query customtypes.M, result interface{}, option FunctionOptions) error
	DeleteOne(ctx context.Context, controller Controller, query customtypes.M) error
	DeleteMany(ctx context.Context, controller Controller, query customtypes.M) error
	SoftDeleteOne(ctx context.Context, controller Controller, query customtypes.M, deletedByUserId string, deletedByUserName string) error
	SoftDeleteMany(ctx context.Context, controller Controller, query customtypes.M, deletedByUserId string, deletedByUserName string) error
	BulkWrite(ctx context.Context, controller Controller, writers []mongo.WriteModel) error
	Distinct(ctx context.Context, controller Controller, key string, filter interface{}) ([]interface{}, error)
	Aggregate(ctx context.Context, controller Controller, result interface{}, basebaseFunctions FunctionOptions, pipelines ...customtypes.M) error
	AggregatePaginate(ctx context.Context, controller Controller, option FunctionOptions, pipelines ...customtypes.M) (genericmodels.PaginationResults, error)
	FindOne(ctx context.Context, controller Controller, query interface{}, result interface{}, option FunctionOptions) error
	Find(ctx context.Context, controller Controller, query customtypes.M, results interface{}, option FunctionOptions) error
	FindCursor(context.Context, Controller, customtypes.M, FunctionOptions, func(result customtypes.M, err error))
	Paginate(ctx context.Context, controller Controller, query customtypes.M, results interface{}, option FunctionOptions) (genericmodels.PaginationResults, error)
}
