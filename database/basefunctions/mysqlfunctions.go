package basefunctions

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"

	"github.com/glodb/keel/database/baseconnections"
	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/httpHandler/controllers/baseinterfaces"
	"github.com/glodb/keel/models/genericmodels"
	"github.com/glodb/keel/settings/customtypes"
	"github.com/glodb/keel/settings/logger"

	"go.mongodb.org/mongo-driver/mongo"
)

type MySqlFunctions struct {
}

func (u *MySqlFunctions) GetFunctions() baseinterfaces.BaseFunctionsInterface {
	return u
}

func (u *MySqlFunctions) EnsureIndex(ctx context.Context, controller baseinterfaces.Controller, data interface{}, unique bool, partialFilterExpression *customtypes.M) error {
	conn := baseconnections.DBConnection().GetConnection(basetypes.MYSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	query := `CREATE TABLE IF NOT EXISTS ` + string(controller.GetCollectionName()) + ` (`
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()

	if dataType.Kind() != reflect.Struct {
		return errors.New("Required a struct for data")
	}

	columns := ""

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		tags := strings.Split(field.Tag.Get("db"), ",")

		if columns != "" {
			columns += ","
		}

		columns += strings.Join(tags, " ")
	}

	query += columns + ");"
	_, err := conn.Exec(query)
	return err
}

func (u *MySqlFunctions) AddMany(ctx context.Context, controller baseinterfaces.Controller, data []interface{}, scan bool) ([]int64, error) {
	return nil, errors.New("unimplemented exception")
}

func (u *MySqlFunctions) SqlPaginate(ctx context.Context, controller baseinterfaces.Controller, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool, pageSize int, page int) (*sql.Rows, int64, error) {
	return nil, -1, errors.New("unimplemented exception")
}

func (u *MySqlFunctions) Count(ctx context.Context, controller baseinterfaces.Controller, condition map[string]interface{}, useOr bool) (int64, error) {
	return -1, nil
}

func (u *MySqlFunctions) Add(ctx context.Context, controller baseinterfaces.Controller, data interface{}, scan bool) (int64, error) {
	conn := baseconnections.DBConnection().GetConnection(basetypes.MYSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	query := "INSERT INTO " + string(controller.GetCollectionName())

	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()

	if dataType.Kind() != reflect.Struct {
		return -1, errors.New("Required a struct for data")
	}

	var columns []string
	var placeholders []string
	values := make([]interface{}, 0)

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		tag := strings.Split(field.Tag.Get("db"), ",")[0]

		if tag == "" {
			continue
		}

		value := dataValue.Field(i).Interface()
		values = append(values, value)

		columns = append(columns, tag)
		placeholders = append(placeholders, "?")
	}

	query += "(" + strings.Join(columns, ", ") + ")"
	query += " VALUES(" + strings.Join(placeholders, ", ") + ")"

	_, err := conn.Exec(query, values...)
	return 0, err
}
func (u *MySqlFunctions) SqlFind(ctx context.Context, controller baseinterfaces.Controller, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool) (*sql.Rows, error) {
	conn := baseconnections.DBConnection().GetConnection(basetypes.MYSQL).GetDB(basetypes.MYSQL).(*sql.DB)

	query := "SELECT * FROM " + string(controller.GetCollectionName())

	whereClause := ""
	values := make([]interface{}, 0)

	for key, val := range condition {
		if whereClause != "" {
			if !useOr {
				whereClause += " AND "
			} else {
				whereClause += " OR "
			}
		} else {
			whereClause += " WHERE "
		}
		whereClause += key + "= ? "
		values = append(values, val)
	}

	if addParenthesis {
		whereClause = "(" + whereClause + ")"
	}

	query += whereClause + appendQuery
	rows, err := conn.Query(query, values...)

	return rows, err
}
func (u *MySqlFunctions) Update(ctx context.Context, controller baseinterfaces.Controller, query string, data []interface{}, upsert bool) error {
	conn := baseconnections.DBConnection().GetConnection(basetypes.MYSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	_, err := conn.Exec(query, data...)
	return err
}

func (u *MySqlFunctions) RawQuery(ctx context.Context, controller baseinterfaces.Controller, query string, data []interface{}) (*sql.Rows, error) {
	return nil, nil
}
func (u *MySqlFunctions) Delete(ctx context.Context, controller baseinterfaces.Controller, condition map[string]interface{}, useOr bool, addParenthesis bool) error {
	logger.Log().Error("Unimplemented DeleteOne MySql")
	return nil
}

func (u *MySqlFunctions) UpdateOne(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, data customtypes.M, upsert bool) error {
	return errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) UpdateMany(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, data customtypes.M, upsert bool) (int64, error) {
	return 0, errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) FindOneAndUpdate(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, data customtypes.M, upsert bool, returnNew bool, result interface{}, option baseinterfaces.FunctionOptions) error {
	return errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) DeleteOne(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M) error {
	return errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) SoftDeleteOne(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, deletedByUserId string, deletedByUserName string) error {
	return errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) DeleteMany(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M) error {
	return errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) SoftDeleteMany(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, deletedByUserId string, deletedByUserName string) error {
	return errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) BulkWrite(ctx context.Context, controller baseinterfaces.Controller, writers []mongo.WriteModel) error {
	return errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) Distinct(ctx context.Context, controller baseinterfaces.Controller, key string, filter interface{}) ([]interface{}, error) {
	return nil, errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) Aggregate(ctx context.Context, controller baseinterfaces.Controller, result interface{}, basebaseFunctions baseinterfaces.FunctionOptions, pipelines ...customtypes.M) error {
	return errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) AggregatePaginate(ctx context.Context, controller baseinterfaces.Controller, option baseinterfaces.FunctionOptions, pipelines ...customtypes.M) (genericmodels.PaginationResults, error) {
	return genericmodels.PaginationResults{}, errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) FindCursor(context.Context, baseinterfaces.Controller, customtypes.M, baseinterfaces.FunctionOptions, func(result customtypes.M, err error)) {
	return
}

func (u *MySqlFunctions) FindOne(ctx context.Context, controller baseinterfaces.Controller, query interface{}, result interface{}, option baseinterfaces.FunctionOptions) error {
	return errors.New("unimplemented for reational db")
}
func (u *MySqlFunctions) Find(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, results interface{}, option baseinterfaces.FunctionOptions) error {
	return errors.New("unimplemented for reational db")
}

func (u *MySqlFunctions) Paginate(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, results interface{}, option baseinterfaces.FunctionOptions) (genericmodels.PaginationResults, error) {
	return genericmodels.PaginationResults{}, errors.New("unimplemented for reational db")
}

func (u *MySqlFunctions) FindOneAndDelete(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, result interface{}, option baseinterfaces.FunctionOptions) error {
	return nil
}
