package basefunctions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/glodb/keel/app/models/dbmodels/keelmodels"
	"github.com/glodb/keel/app/models/genericmodels"
	"github.com/glodb/keel/database/baseconnections"
	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/httpHandler/controllers/baseinterfaces"
	"github.com/glodb/keel/settings/customtypes"

	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
)

type PSqlFunctions struct {
}

func (u *PSqlFunctions) GetFunctions() baseinterfaces.BaseFunctionsInterface {
	return u
}

func (u *PSqlFunctions) EnsureIndex(ctx context.Context, controller baseinterfaces.Controller, data interface{}, unique bool, partialFilterExpression *customtypes.M) error {
	conn := baseconnections.DBConnection().GetConnection(basetypes.PSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	query := `CREATE TABLE IF NOT EXISTS ` + string(controller.GetCollectionName()) + ` (`
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()

	if dataType.Kind() != reflect.Struct {
		return errors.New("required a struct for data")
	}

	columns := ""

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		tagValue := field.Tag.Get("db")

		if tagValue == "" {
			continue
		}
		tags := strings.Split(tagValue, ",")

		if columns != "" {
			columns += ","
		}

		columns += strings.Join(tags, " ")
	}

	query += columns + ");"
	_, err := conn.Exec(query)
	return err
}

func (u *PSqlFunctions) IsNull(value interface{}) bool {
	switch v := value.(type) {
	case time.Time:
		// Check if it's a non-zero time value
		if !v.IsZero() {
			return false
		} else {
			return true
		}
	case string:
		// Check if it's a non-empty string
		if v != "" {
			return false
		} else {
			return true
		}
	default:
		return false
	}
}

func (u *PSqlFunctions) Add(ctx context.Context, controller baseinterfaces.Controller, data interface{}, scan bool) (int64, error) {
	conn := baseconnections.DBConnection().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	query := "INSERT INTO " + string(controller.GetCollectionName())

	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()

	if dataType.Kind() != reflect.Struct {
		return -1, errors.New("required a struct for data")
	}

	var columns []string
	var placeholders []string
	values := make([]interface{}, 0)

	placeholderCount := 1

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		tagVal := field.Tag.Get("db")
		if strings.Contains(strings.ToUpper(tagVal), "SERIAL") {
			continue
		}
		tag := strings.Split(tagVal, " ")[0]

		if tag == "" {
			continue
		}

		value := dataValue.Field(i).Interface()

		if value != nil && !u.IsNull(value) {
			if strings.Contains(tagVal, "[]") {
				values = append(values, pq.Array(value))

			} else {
				values = append(values, value)
			}

			columns = append(columns, tag)
			placeholders = append(placeholders, "$"+strconv.FormatInt(int64(placeholderCount), 10))
			placeholderCount++
		}
	}

	query += "(" + strings.Join(columns, ", ") + ")"
	query += " VALUES(" + strings.Join(placeholders, ", ") + ")"
	if scan {
		query += " RETURNING id"
	}

	var insertedID int64
	row := conn.QueryRow(
		query,
		values...,
	)
	if row.Err() != nil {
		return -1, row.Err()
	}
	var err error
	if scan {
		err = row.Scan(&insertedID)
	}
	if err != nil {
		return -1, err
	}
	return insertedID, err
}

func (u *PSqlFunctions) AddMany(ctx context.Context, controller baseinterfaces.Controller, dataArray []interface{}, scan bool) ([]int64, error) {
	conn := baseconnections.DBConnection().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	query := "INSERT INTO " + string(controller.GetCollectionName())

	var columns []string
	values := make([]interface{}, 0)
	placeholderCount := 1

	for i, data := range dataArray {
		dataValue := reflect.ValueOf(data)
		dataType := dataValue.Type()

		if dataType.Kind() != reflect.Struct {
			return nil, errors.New("required a struct for data")
		}

		var placeholders []string

		for i := 0; i < dataType.NumField(); i++ {
			field := dataType.Field(i)
			if strings.Contains(strings.ToUpper(field.Tag.Get("db")), "SERIAL") {
				continue
			}
			tag := strings.Split(field.Tag.Get("db"), " ")[0]

			if tag == "" {
				continue
			}

			value := dataValue.Field(i).Interface()
			values = append(values, value)

			columns = append(columns, tag)
			placeholders = append(placeholders, "$"+strconv.FormatInt(int64(placeholderCount), 10))
			placeholderCount++
		}
		if i == 0 {
			query += "(" + strings.Join(columns, ", ") + ")"
			query += " VALUES(" + strings.Join(placeholders, ", ") + ")"
		} else {
			query += ", (" + strings.Join(placeholders, ", ") + ")"
		}
	}

	query += " RETURNING id"

	var insertedID []int64
	row := conn.QueryRow(
		query,
		values...,
	)
	if row.Err() != nil {
		return nil, row.Err()
	}
	var err error
	if scan {
		err = row.Scan(&insertedID)
	}
	if err != nil {
		return nil, err
	}
	return insertedID, nil
	// return insertedID, err
}

func (u *PSqlFunctions) SqlFind(ctx context.Context, controller baseinterfaces.Controller, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool) (*sql.Rows, error) {
	conn := baseconnections.DBConnection().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	query := "SELECT * FROM " + string(controller.GetCollectionName())

	if keys != "" {
		query = "SELECT " + keys + " FROM " + string(controller.GetCollectionName())
	}

	whereClause := ""
	values := make([]interface{}, 0)

	placeholderCount := 1

	for key, val := range condition {
		if whereClause != "" {
			if !useOr {
				whereClause += " AND "
			} else {
				whereClause += " OR "
			}
		} else {
			whereClause += " WHERE "
			if addParenthesis {
				whereClause = whereClause + "("
			}
		}
		whereClause += key + "= $" + strconv.FormatInt(int64(placeholderCount), 10) + " "
		placeholderCount++
		values = append(values, val)
	}
	if addParenthesis {
		whereClause = whereClause + ")"
	}

	query += whereClause + appendQuery
	rows, err := conn.Query(query, values...)

	return rows, err
}

func (u *PSqlFunctions) SqlPaginate(ctx context.Context, controller baseinterfaces.Controller, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool, pageSize int, page int) (*sql.Rows, int64, error) {
	conn := baseconnections.DBConnection().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	query := "SELECT * FROM " + string(controller.GetCollectionName())

	if keys != "" {
		query = "SELECT " + keys + " FROM " + string(controller.GetCollectionName())
	}

	whereClause := ""
	values := make([]interface{}, 0)

	placeholderCount := 1

	for key, val := range condition {
		if whereClause != "" {
			if !useOr {
				whereClause += " AND "
			} else {
				whereClause += " OR "
			}
		} else {
			whereClause += " WHERE "
			if addParenthesis {
				whereClause = whereClause + "("
			}
		}
		whereClause += key + "= $" + strconv.FormatInt(int64(placeholderCount), 10) + " "
		placeholderCount++
		values = append(values, val)
	}
	if addParenthesis {
		whereClause = whereClause + ")"
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", string(controller.GetCollectionName())) + whereClause

	var count int64
	row, err := conn.Query(countQuery, values...)
	if err != nil {
		return nil, -1, err
	}

	countQueryRows := 0
	for row.Next() {
		countQueryRows++
		err = row.Scan(&count)
		if err != nil {
			return nil, -1, err
		}
	}

	query += whereClause + appendQuery

	query = fmt.Sprintf(query+" LIMIT %d", pageSize)

	// If there is a skip offset specified, append the OFFSET clause to the query.
	skip := (page - 1) * pageSize
	if skip > 0 {
		query = fmt.Sprintf(query+" OFFSET %d", skip)
	}

	rows, err := conn.Query(query, values...)

	return rows, count, err
}

func (u *PSqlFunctions) Count(ctx context.Context, controller baseinterfaces.Controller, condition map[string]interface{}, useOr bool) (int64, error) {
	conn := baseconnections.DBConnection().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	whereClause := ""
	values := make([]interface{}, 0)

	placeholderCount := 1

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
		whereClause += key + "= $" + strconv.FormatInt(int64(placeholderCount), 10) + " "
		placeholderCount++
		values = append(values, val)
	}
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", string(controller.GetCollectionName())) + whereClause

	var count int64
	row, err := conn.Query(countQuery, values...)
	if err != nil {
		return -1, err
	}

	countQueryRows := 0
	for row.Next() {
		countQueryRows++
		err = row.Scan(&count)
		if err != nil {
			return -1, err
		}
	}
	return count, nil
}

func (u *PSqlFunctions) Update(ctx context.Context, controller baseinterfaces.Controller, query string, data []interface{}, upsert bool) error {
	if !strings.Contains(query, "UPDATE") {
		return errors.New("format of query seems in correct")
	}
	conn := baseconnections.DBConnection().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	_, err := conn.Exec(query, data...)
	return err
}

func (u *PSqlFunctions) RawQuery(ctx context.Context, controller baseinterfaces.Controller, query string, data []interface{}) (*sql.Rows, error) {
	conn := baseconnections.DBConnection().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	rows, err := conn.Query(query, data...)
	return rows, err
}

func (u *PSqlFunctions) Delete(ctx context.Context, controller baseinterfaces.Controller, condition map[string]interface{}, useOr bool, addParenthesis bool) error {
	if len(condition) == 0 {
		return errors.New("delete can't run with out conditions")
	}

	conn := baseconnections.DBConnection().GetConnection(basetypes.PSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	query := "DELETE FROM " + string(controller.GetCollectionName())

	whereClause := ""
	values := make([]interface{}, 0)

	placeholderCount := 1

	for key, val := range condition {
		if whereClause != "" {
			if !useOr {
				whereClause += " AND "
			} else {
				whereClause += " OR "
			}
		} else {
			whereClause += " WHERE "
			if addParenthesis {
				whereClause = whereClause + "("
			}
		}
		whereClause += key + "= $" + strconv.FormatInt(int64(placeholderCount), 10) + " "
		placeholderCount++
		values = append(values, val)
	}
	if addParenthesis {
		whereClause = whereClause + ")"
	}

	query += whereClause
	_, err := conn.Exec(query, values...)
	return err
}

func (u *PSqlFunctions) UpdateOne(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, data customtypes.M, upsert bool) error {
	return errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) UpdateMany(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, data customtypes.M, upsert bool) (int64, error) {
	return 0, errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) FindOneAndUpdate(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, data customtypes.M, upsert bool, returnNew bool, result interface{}, option baseinterfaces.FunctionOptions) error {
	return errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) DeleteOne(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M) error {
	return errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) SoftDeleteOne(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, deletedBy keelmodels.Session) error {
	return errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) DeleteMany(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M) error {
	return errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) SoftDeleteMany(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, deletedBy keelmodels.Session) error {
	return errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) BulkWrite(ctx context.Context, controller baseinterfaces.Controller, writers []mongo.WriteModel) error {
	return errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) Distinct(ctx context.Context, controller baseinterfaces.Controller, key string, filter interface{}) ([]interface{}, error) {
	return nil, errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) Aggregate(ctx context.Context, controller baseinterfaces.Controller, result interface{}, basebaseFunctions baseinterfaces.FunctionOptions, pipelines ...customtypes.M) error {
	return errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) AggregatePaginate(ctx context.Context, controller baseinterfaces.Controller, option baseinterfaces.FunctionOptions, pipelines ...customtypes.M) (genericmodels.PaginationResults, error) {
	return genericmodels.PaginationResults{}, errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) FindCursor(context.Context, baseinterfaces.Controller, customtypes.M, baseinterfaces.FunctionOptions, func(result customtypes.M, err error)) {
	return
}
func (u *PSqlFunctions) FindOne(ctx context.Context, controller baseinterfaces.Controller, query interface{}, result interface{}, option baseinterfaces.FunctionOptions) error {
	return errors.New("unimplemented for reational db")
}
func (u *PSqlFunctions) Find(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, results interface{}, option baseinterfaces.FunctionOptions) error {
	return errors.New("unimplemented for reational db")
}

func (u *PSqlFunctions) Paginate(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, results interface{}, option baseinterfaces.FunctionOptions) (genericmodels.PaginationResults, error) {
	return genericmodels.PaginationResults{}, errors.New("unimplemented for reational db")
}

func (u *PSqlFunctions) FindOneAndDelete(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, result interface{}, option baseinterfaces.FunctionOptions) error {
	return nil
}
