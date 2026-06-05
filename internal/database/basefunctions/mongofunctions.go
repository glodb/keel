package basefunctions

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"github.com/glodb/keel/database/baseconnections"
	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/httpHandler/controllers/baseinterfaces"
	"github.com/glodb/keel/models/genericmodels"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/customtypes"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBFunctions struct {
}

func (u *MongoDBFunctions) GetFunctions() baseinterfaces.BaseFunctionsInterface {
	return u
}

func (u *MongoDBFunctions) EnsureIndex(ctx context.Context, controller baseinterfaces.Controller, data interface{}, unique bool, partialFilterExpression *customtypes.M) error {
	logger.Log().Debug("MongoDB-EnsureIndex-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("data", data))
	start := time.Now()
	dictData, _ := utils.GetInstance().ToBsonDict(data)

	idxOpts := options.Index().
		SetUnique(unique)

	if partialFilterExpression != nil {
		partialDictData, _ := utils.GetInstance().ToBsonDict(*partialFilterExpression)
		idxOpts.SetPartialFilterExpression(partialDictData)
	}

	indexModel := mongo.IndexModel{Keys: dictData, Options: idxOpts}

	// SetPartialFilterExpression(bson.D{
	//     {"sportsId", bson.D{{"$exists", true}}},
	// })

	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	_, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		logger.Log().Debug("MongoDB-EnsureIndex-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-EnsureIndex-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return err
}

func (u *MongoDBFunctions) Add(ctx context.Context, controller baseinterfaces.Controller, data interface{}, scan bool) (int64, error) {
	logger.Log().Debug("MongoDB-Add-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("data", data))
	start := time.Now()
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	_, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).InsertOne(ctx, data)
	if err != nil {
		logger.Log().Debug("MongoDB-Add-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-Add-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return -1, err
}

func (u *MongoDBFunctions) AddMany(ctx context.Context, controller baseinterfaces.Controller, dataArray []interface{}, scan bool) ([]int64, error) {
	logger.Log().Debug("MongoDB-AddMany-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("data", dataArray))
	start := time.Now()
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	_, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).InsertMany(ctx, dataArray)
	if err != nil {
		logger.Log().Debug("MongoDB-AddMany-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-AddMany-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return []int64{}, err
}

func (u *MongoDBFunctions) Count(ctx context.Context, controller baseinterfaces.Controller, condition map[string]interface{}, useOr bool) (int64, error) {
	logger.Log().Debug("MongoDB-Count-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("condition", condition))
	start := time.Now()
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	count, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).CountDocuments(ctx, condition)
	if err != nil {
		logger.Log().Debug("MongoDB-Count-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-Count-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return count, err
}

func (u *MongoDBFunctions) FindOne(ctx context.Context, controller baseinterfaces.Controller, query interface{}, result interface{}, option baseinterfaces.FunctionOptions) error {
	logger.Log().Debug("MongoDB-FindOne-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query))
	start := time.Now()
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)

	opts := options.FindOne()
	if option.IsProjectionSet() {
		opts.SetProjection(option.GetProjection())
	}

	if option.IsSortSet() {
		opts.SetSort(option.GetSort())
	}

	err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).FindOne(ctx, query, opts).Decode(result)
	if err != nil {
		logger.Log().Debug("MongoDB-FindOne-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-FindOne-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return err
}

func (u *MongoDBFunctions) FindOneAndUpdate(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, update customtypes.M, upsert bool, returnNew bool, result interface{}, option baseinterfaces.FunctionOptions) error {
	logger.Log().Debug("MongoDB-FindOneAndUpdate-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query), logger.AnyField("update", update))
	start := time.Now()
	returnType := options.Before
	if returnNew {
		returnType = options.After
	}
	opts := options.FindOneAndUpdate().SetUpsert(upsert).SetReturnDocument(returnType)
	if option.IsProjectionSet() {
		opts.SetProjection(option.GetProjection())
	}

	if option.IsSortSet() {
		opts.SetSort(option.GetSort())
	}
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).FindOneAndUpdate(ctx, query, update, opts).Decode(result)
	if err != nil {
		logger.Log().Debug("MongoDB-FindOneAndUpdate-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-FindOneAndUpdate-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return err
}

func (u *MongoDBFunctions) UpdateOne(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, data customtypes.M, upsert bool) error {
	logger.Log().Debug("MongoDB-UpdateOne-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query), logger.AnyField("data", data))
	start := time.Now()
	opts := options.Update().SetUpsert(upsert)
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	updateResults, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).UpdateOne(ctx, query, data, opts)
	if err == nil {
		if updateResults.ModifiedCount == 0 && updateResults.UpsertedCount == 0 {
			return errors.New("no record updated")
		}
	}
	if err != nil {
		logger.Log().Debug("MongoDB-UpdateOne-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-UpdateOne-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return err
}

func (u *MongoDBFunctions) UpdateMany(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, data customtypes.M, upsert bool) (int64, error) {
	logger.Log().Debug("MongoDB-UpdateMany-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query), logger.AnyField("data", data))
	start := time.Now()
	opts := options.Update().SetUpsert(upsert)
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	updateResults, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).UpdateMany(ctx, query, data, opts)

	if err == nil {
		if updateResults.ModifiedCount == 0 && updateResults.UpsertedCount == 0 {
			return 0, errors.New("no record updated")
		}
		return updateResults.ModifiedCount, err
	}
	if err != nil {
		logger.Log().Debug("MongoDB-UpdateMany-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-UpdateMany-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return 0, err
}

func (u *MongoDBFunctions) DeleteOne(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M) error {
	logger.Log().Debug("MongoDB-DeleteOne-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query))
	start := time.Now()
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	deleteResult, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).DeleteOne(ctx, query)
	if deleteResult.DeletedCount == 0 {
		return errors.New("no record deleted")
	}
	if err != nil {
		logger.Log().Debug("MongoDB-DeleteOne-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-DeleteOne-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return err
}

func (u *MongoDBFunctions) FindOneAndDelete(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, result interface{}, option baseinterfaces.FunctionOptions) error {
	logger.Log().Debug("MongoDB-FindOneAndDelete-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query))
	start := time.Now()
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).FindOneAndDelete(ctx, query, &options.FindOneAndDeleteOptions{}).Decode(result)
	if err != nil {
		logger.Log().Debug("MongoDB-FindOneAndDelete-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-FindOneAndDelete-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return err
}

func (u *MongoDBFunctions) SoftDeleteOne(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, deletedByUserId string, deletedByUserName string) error {
	logger.Log().Debug("MongoDB-SoftDeleteOne-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query))
	start := time.Now()
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	// conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).FindOneAndDelete()

	deletedData := customtypes.M{}
	err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).FindOneAndDelete(ctx, query, &options.FindOneAndDeleteOptions{}).Decode(&deletedData)
	if err != nil {
		return err
	}

	deletedData[configmanager.GetInstance().SoftDeletionKey] = time.Now()
	deletedData[configmanager.GetInstance().DeletedByKey] = deletedByUserId
	_, err = conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())+configmanager.GetInstance().SoftDeleteCollectionPrefix).InsertOne(ctx, deletedData)

	if err != nil {
		//Rollback
		conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).InsertOne(ctx, deletedData)
	}

	if err != nil {
		logger.Log().Debug("MongoDB-SoftDeleteOne-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-SoftDeleteOne-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return err
}

func (u *MongoDBFunctions) DeleteMany(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M) error {
	logger.Log().Debug("MongoDB-DeleteMany-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query))
	start := time.Now()
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	deleteResult, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).DeleteMany(ctx, query)
	if err != nil {
		return err
	}
	if deleteResult.DeletedCount == 0 {
		return errors.New("no record deleted")
	}
	if err != nil {
		logger.Log().Debug("MongoDB-DeleteMany-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-DeleteMany-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return err
}

func (u *MongoDBFunctions) SoftDeleteMany(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, deletedByUserId string, deletedByUserName string) error {
	logger.Log().Debug("MongoDB-SoftDeleteMany-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query))
	start := time.Now()
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)

	cursor, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).Find(context.Background(), query)
	if err != nil {
		logger.Log().Debug("MongoDB-Error finding documents", logger.ErrorField("error", err))
		return err
	}
	defer cursor.Close(context.Background())

	var documentsToInsert []interface{}
	for cursor.Next(context.Background()) {
		var document customtypes.M
		if err := cursor.Decode(&document); err != nil {
			return err
		}
		document[configmanager.GetInstance().SoftDeletionKey] = time.Now()
		document[configmanager.GetInstance().DeletedByKey] = deletedByUserId
		documentsToInsert = append(documentsToInsert, document)
	}
	if _, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())+configmanager.GetInstance().SoftDeleteCollectionPrefix).InsertMany(context.Background(), documentsToInsert); err != nil {
		return err
	}

	_, err = conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).DeleteMany(context.Background(), query)
	if err != nil {
		conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).InsertMany(context.Background(), documentsToInsert)
		return err
	}
	if err != nil {
		logger.Log().Debug("MongoDB-SoftDeleteMany-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-BulkWrite-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return err
}

func (u *MongoDBFunctions) BulkWrite(ctx context.Context, controller baseinterfaces.Controller, writers []mongo.WriteModel) error {
	logger.Log().Debug("MongoDB-BulkWrite-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("writers", writers))
	start := time.Now()
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	_, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).BulkWrite(ctx, writers, &options.BulkWriteOptions{})
	if err != nil {
		logger.Log().Debug("MongoDB-BulkWrite-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-BulkWrite-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return err
}

func (u *MongoDBFunctions) Find(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, results interface{}, option baseinterfaces.FunctionOptions) error {
	logger.Log().Debug("MongoDB-Find-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query))
	start := time.Now()
	skip := (option.GetPage() - 1) * option.GetLimit()
	opts := options.Find().SetSkip(int64(skip))
	if option.IsLimitSet() {
		opts.SetLimit(int64(option.GetLimit()))
	}

	if option.IsProjectionSet() {
		opts.SetProjection(option.GetProjection())
	}

	if option.IsSortSet() {
		opts.SetSort(option.GetSort())
	}

	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	cursor, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).Find(ctx, query, opts)
	if err != nil {
		logger.Log().Debug("MongoDB-Find-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
		return err
	}
	if err = cursor.All(ctx, results); err == nil {
		return nil
	}
	logger.Log().Debug("MongoDB-Find-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	logger.Log().Debug("MongoDB-Find-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	return err
}

func (u *MongoDBFunctions) FindCursor(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, option baseinterfaces.FunctionOptions, action func(result customtypes.M, err error)) {
	logger.Log().Debug("MongoDB-FindCursor-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query))
	skip := (option.GetPage() - 1) * option.GetLimit()
	opts := options.Find().SetSkip(int64(skip))

	if option.IsLimitSet() {
		opts.SetLimit(int64(option.GetLimit()))
	}

	if option.IsProjectionSet() {
		opts.SetProjection(option.GetProjection())
	}

	if option.IsSortSet() {
		opts.SetSort(option.GetSort())
	}

	if option.IsTimeRangeSet() {
		query[option.GetTimeRangeKey()] = option.GetTimeRange()
	}
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	cursor, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).Find(ctx, query, opts)

	if err != nil {
		action(nil, err)
		return
	}

	for cursor.Next(context.Background()) {
		var result customtypes.M
		if err := cursor.Decode(&result); err != nil {
			action(nil, err)
		}
		action(result, nil)
	}
	logger.Log().Debug("MongoDB-FindCursor-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())))
}

func (u *MongoDBFunctions) Distinct(ctx context.Context, controller baseinterfaces.Controller, key string, filter interface{}) ([]interface{}, error) {
	logger.Log().Debug("MongoDB-Distinct-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("key", key), logger.AnyField("filter", filter))
	start := time.Now()
	opts := options.DistinctOptions{}
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	objectArray, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).Distinct(ctx, key, filter, &opts)
	if err != nil {
		logger.Log().Debug("MongoDB-Distinct-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	}
	logger.Log().Debug("MongoDB-Distinct-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return objectArray, err
}

func (u *MongoDBFunctions) Paginate(ctx context.Context, controller baseinterfaces.Controller, query customtypes.M, results interface{}, option baseinterfaces.FunctionOptions) (genericmodels.PaginationResults, error) {
	logger.Log().Debug("MongoDB-Paginate-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("query", query))
	start := time.Now()
	skip := (option.GetPage() - 1) * option.GetLimit()
	opts := options.Find().SetLimit(int64(option.GetLimit())).SetSkip(int64(skip))

	if option.IsProjectionSet() {
		opts.SetProjection(option.GetProjection())
	}

	if option.IsSortSet() {
		opts.SetSort(option.GetSort())
	}

	if option.IsTimeRangeSet() {
		query[option.GetTimeRangeKey()] = option.GetTimeRange()
	}

	// Track active query to prevent health checker from closing connection mid-query
	baseconnections.GetConnectionPool().IncrementActiveQueries(basetypes.MONGO)
	defer baseconnections.GetConnectionPool().DecrementActiveQueries(basetypes.MONGO)

	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	totalDocuments, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).CountDocuments(ctx, query)
	if err != nil {
		logger.Log().Debug("MongoDB-CountDocuments-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
		return genericmodels.PaginationResults{}, err
	}
	cursor, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).Find(ctx, query, opts)
	if err != nil {
		logger.Log().Debug("MongoDB-Find-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
		return genericmodels.PaginationResults{}, err
	}
	if err = cursor.All(ctx, results); err == nil {
		paginationResults := genericmodels.PaginationResults{
			Pagination: genericmodels.Pagination{
				Limit:          int64(option.GetLimit()),
				TotalDocuments: int64(totalDocuments),
				TotalPages:     int64(math.Ceil(float64(totalDocuments) / float64(option.GetLimit()))),
				CurrentPage:    int64(option.GetPage()),
			},
			Data: results,
		}
		return paginationResults, nil
	}
	logger.Log().Debug("MongoDB-Paginate-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	logger.Log().Debug("MongoDB-Paginate-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
	return genericmodels.PaginationResults{Pagination: genericmodels.Pagination{
		Limit: int64(option.GetLimit()),
	}}, err
}

// Use bson.D in the aggregate function in which order matters on mongodb
func (u *MongoDBFunctions) Aggregate(ctx context.Context, controller baseinterfaces.Controller, results interface{}, option baseinterfaces.FunctionOptions, pipelines ...customtypes.M) error {
	logger.Log().Debug("MongoDB-Aggregate-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("pipelines", pipelines))
	start := time.Now()
	skipStage := bson.D{{Key: "$skip", Value: (option.GetPage() - 1) * option.GetLimit()}}
	limitStage := bson.D{{Key: "$limit", Value: option.GetLimit()}}

	pipeline := mongo.Pipeline{}

	for _, element := range pipelines {
		dict, _ := utils.GetInstance().ToBsonDict(element)
		pipeline = append(pipeline, *dict)
	}

	if option.IsSortSet() {
		sortStage := bson.D{{Key: "$sort", Value: option.GetSort()}}
		pipeline = append(pipeline, sortStage)
	}

	pipeline = append(pipeline, skipStage, limitStage)

	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)
	cursor, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).Aggregate(ctx, pipeline)

	if err != nil {
		logger.Log().Debug("MongoDB-Aggregate-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
		return err
	}

	if err = cursor.All(ctx, results); err == nil {
		logger.Log().Debug("MongoDB-Aggregate-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
		return nil
	}
	logger.Log().Debug("MongoDB-Aggregate-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	return err
}

func (u *MongoDBFunctions) AggregatePaginate(ctx context.Context, controller baseinterfaces.Controller, option baseinterfaces.FunctionOptions, pipelines ...customtypes.M) (genericmodels.PaginationResults, error) {
	logger.Log().Debug("MongoDB-AggregatePaginate-Start", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.AnyField("pipelines", pipelines))
	start := time.Now()
	conn := baseconnections.DBConnection().GetConnection(basetypes.MONGO).GetDB(basetypes.MONGO).(*mongo.Client)

	// Convert pipelines into a mongo.Pipeline dynamically
	pipeline := mongo.Pipeline{}
	for _, element := range pipelines {
		dict, _ := utils.GetInstance().ToBsonDict(element)
		pipeline = append(pipeline, *dict)
	}

	// Stage for counting total documents
	secondGroupStage := customtypes.M{"$group": customtypes.M{"_id": nil, "count": customtypes.M{"$sum": 1}}}
	dict, _ := utils.GetInstance().ToBsonDict(secondGroupStage)
	pipeline = append(pipeline, *dict)

	// Execute the first aggregation for count
	cursor, err := conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).Aggregate(ctx, pipeline)
	if err != nil {
		logger.Log().Debug("MongoDB-Aggregate-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
		return genericmodels.PaginationResults{}, err
	}

	// Parse results and get the total count
	paginationResults := []customtypes.M{}
	if err = cursor.All(ctx, &paginationResults); err == nil {
		totalDocuments := int32(0)
		if len(paginationResults) > 0 {
			totalDocuments = paginationResults[0]["count"].(int32)
		}

		// Build the final pipeline with custom stages, pagination, and sorting
		finalPipeline := mongo.Pipeline{}

		// Rebuild initial pipeline stages (without the count group stage)
		for _, element := range pipelines {
			dict, _ := utils.GetInstance().ToBsonDict(element)
			finalPipeline = append(finalPipeline, *dict)
		}

		// Add pagination stages
		skipStage := bson.D{{Key: "$skip", Value: (option.GetPage() - 1) * option.GetLimit()}}
		limitStage := bson.D{{Key: "$limit", Value: option.GetLimit()}}
		finalPipeline = append(finalPipeline, skipStage, limitStage)

		// Add optional sorting if specified
		if option.IsSortSet() {
			sortStage := bson.D{{Key: "$sort", Value: option.GetSort()}}
			finalPipeline = append(finalPipeline, sortStage)
		}

		// Execute the final aggregation
		cursor, err = conn.Database(string(controller.GetDBName())).Collection(string(controller.GetCollectionName())).Aggregate(ctx, finalPipeline)
		if err != nil {
			logger.Log().Debug("MongoDB-Aggregate-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
			return genericmodels.PaginationResults{}, err
		}
		var results []customtypes.M
		if err = cursor.All(ctx, &results); err == nil {
			paginationResults := genericmodels.PaginationResults{
				Pagination: genericmodels.Pagination{
					Limit:          int64(option.GetLimit()),
					TotalDocuments: int64(totalDocuments),
					TotalPages:     int64(math.Ceil(float64(totalDocuments) / float64(option.GetLimit()))),
					CurrentPage:    int64(option.GetPage()),
				},
				Data: results,
			}
			logger.Log().Debug("MongoDB-AggregatePaginate-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
			return paginationResults, nil
		}
		logger.Log().Debug("MongoDB-Aggregate-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	} else {
		paginationResults := genericmodels.PaginationResults{
			Pagination: genericmodels.Pagination{
				Limit:          int64(option.GetLimit()),
				TotalDocuments: int64(0),
				TotalPages:     int64(0),
				CurrentPage:    int64(option.GetPage()),
			},
			Data: []customtypes.M{},
		}
		logger.Log().Debug("MongoDB-AggregatePaginate-End", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.StringField("duration", time.Since(start).String()))
		return paginationResults, nil
	}
	logger.Log().Debug("MongoDB-Aggregate-Error", logger.StringField("collection", string(controller.GetCollectionName())), logger.StringField("db", string(controller.GetDBName())), logger.ErrorField("error", err))
	return genericmodels.PaginationResults{}, err
}

func (u *MongoDBFunctions) SqlFind(ctx context.Context, controller baseinterfaces.Controller, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool) (*sql.Rows, error) {
	return nil, errors.New("unimplemented exception")
}

func (u *MongoDBFunctions) SqlPaginate(ctx context.Context, controller baseinterfaces.Controller, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool, pageSize int, page int) (*sql.Rows, int64, error) {
	return nil, -1, errors.New("unimplemented exception")
}

func (u *MongoDBFunctions) Delete(ctx context.Context, controller baseinterfaces.Controller, condition map[string]interface{}, useOr bool, addParenthesis bool) error {
	return errors.New("unimplemented exception, use deleteone or deletemany")
}

func (u *MongoDBFunctions) Update(ctx context.Context, controller baseinterfaces.Controller, query string, data []interface{}, upsert bool) error {
	return errors.New("unimplemented exception, use updateone or updatemany")
}

func (u *MongoDBFunctions) RawQuery(ctx context.Context, controller baseinterfaces.Controller, query string, data []interface{}) (*sql.Rows, error) {
	return nil, errors.New("unimplemented exception, for mongodb")
}
