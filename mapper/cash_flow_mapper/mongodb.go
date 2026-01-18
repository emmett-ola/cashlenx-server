package cash_flow_mapper

import (
	"context"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CashFlowMongoDbMapper struct{}

func (CashFlowMongoDbMapper) GetCashFlowByObjectId(plainId string) model.CashFlowEntity {
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("cash_flow's id is not acceptable")
		return model.CashFlowEntity{}
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()
	return convertBsonM2CashFlowEntity(database.GetOneInMongoDB(filter))
}

func (CashFlowMongoDbMapper) GetCashFlowsByObjectIdArray(plainIdList []string) []model.CashFlowEntity {
	objectIdArray := make([]primitive.ObjectID, len(plainIdList))
	for _, plainId := range plainIdList {
		objectId := util.Convert2ObjectId(plainId)
		objectIdArray = append(objectIdArray, objectId)
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: bson.M{"$in": objectIdArray}},
	}

	// Open connection to cashFlow table
	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	// Get query results and convert to entity objects
	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}
	return targetEntityList
}

func (CashFlowMongoDbMapper) GetCashFlowsByBelongsDate(belongsDate time.Time) []model.CashFlowEntity {
	filter := bson.D{
		primitive.E{Key: "belongs_date", Value: belongsDate},
	}

	// Open connection to cashFlow table
	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	// 获取查询结果并转入结构对象
	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}
	return targetEntityList
}

func (CashFlowMongoDbMapper) GetCashFlowsByDateRange(from, to time.Time) []model.CashFlowEntity {
	filter := bson.D{
		primitive.E{Key: "belongs_date", Value: bson.M{
			"$gte": from,
			"$lte": to,
		}},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}
	return targetEntityList
}

func (CashFlowMongoDbMapper) GetCashFlowsByCategoryId(categoryPlainId string) []model.CashFlowEntity {
	categoryObjectId := util.Convert2ObjectId(categoryPlainId)
	if categoryPlainId == "" || categoryObjectId == primitive.NilObjectID {
		util.Logger.Warnln("category's id is not acceptable")
		return nil
	}

	filter := bson.D{
		primitive.E{Key: "category_id", Value: categoryObjectId},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}
	return targetEntityList
}

func (CashFlowMongoDbMapper) CountCashFLowsByCategoryId(categoryPlainId string) int64 {
	categoryObjectId := util.Convert2ObjectId(categoryPlainId)
	if categoryPlainId == "" || categoryObjectId == primitive.NilObjectID {
		util.Logger.Warnln("category's id is not acceptable")
		return 0
	}

	filter := bson.D{
		primitive.E{Key: "category_id", Value: categoryObjectId},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	return database.CountInMongoDB(filter)
}

func (CashFlowMongoDbMapper) GetCashFlowsByExactDesc(description string) []model.CashFlowEntity {
	filter := bson.D{
		primitive.E{Key: "description", Value: description},
	}

	// Open connection to cashFlow table
	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	// 获取查询结果并转入结构对象
	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}

	return targetEntityList
}

func (CashFlowMongoDbMapper) GetCashFlowsByFuzzyDesc(description string) []model.CashFlowEntity {
	// Options i for disable case sensitive.
	filter := bson.D{
		primitive.E{Key: "description", Value: primitive.Regex{
			Pattern: description,
			Options: "i",
		}},
	}

	// Open connection to cash_flow table
	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	// 获取查询结果并转入结构对象
	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}
	return targetEntityList
}

func (CashFlowMongoDbMapper) InsertCashFlowByEntity(newEntity model.CashFlowEntity) string {
	// Only set CreateTime and UpdateTime if they're not already set (e.g., during restoration)
	operatingTime := time.Now().UTC() // Store in UTC
	if newEntity.CreateTime.IsZero() {
		newEntity.CreateTime = operatingTime
	}
	if newEntity.UpdateTime.IsZero() {
		newEntity.UpdateTime = operatingTime
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	newCashFlowId := database.InsertOneInMongoDB(convertCashFlowEntity2BsonD(newEntity))
	return newCashFlowId.Hex()
}

func (CashFlowMongoDbMapper) BulkInsertCashFlows(entities []model.CashFlowEntity) ([]string, error) {
	if len(entities) == 0 {
		return []string{}, nil
	}

	operatingTime := time.Now().UTC() // Store in UTC
	documents := make([]interface{}, len(entities))

	for i, entity := range entities {
		// Only set CreateTime and UpdateTime if they're not already set (e.g., during restoration)
		if entity.CreateTime.IsZero() {
			entity.CreateTime = operatingTime
		}
		if entity.UpdateTime.IsZero() {
			entity.UpdateTime = operatingTime
		}
		documents[i] = convertCashFlowEntity2BsonD(entity)
	}

	collection := database.GetMongoCollection(database.CashFlowTableName)
	result, err := collection.InsertMany(context.TODO(), documents)
	if err != nil {
		util.Logger.Errorw("bulk insert failed", "error", err)
		return nil, err
	}

	ids := make([]string, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		ids[i] = id.(primitive.ObjectID).Hex()
	}

	util.Logger.Infow("bulk insert successful", "count", len(ids))
	return ids, nil
}

func (CashFlowMongoDbMapper) UpdateCashFlowByEntity(plainId string, updatedEntity model.CashFlowEntity) model.CashFlowEntity {
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("cash_flow's id is not acceptable")
		return model.CashFlowEntity{}
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	targetEntity := convertBsonM2CashFlowEntity(database.GetOneInMongoDB(filter))
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("cash_flow is not exist")
		return model.CashFlowEntity{}
	}

	// Update fields from updatedEntity while preserving ID and CreateTime
	updatedEntity.Id = targetEntity.Id
	updatedEntity.CreateTime = targetEntity.CreateTime
	updatedEntity.CreateUserId = targetEntity.CreateUserId
	updatedEntity.UpdateTime = time.Now().UTC() // Store in UTC

	rowsAffected := database.UpdateManyInMongoDB(filter, convertCashFlowEntity2BsonD(updatedEntity))
	if rowsAffected != 1 {
		// fixme: maybe we should have a rollback here.
		util.Logger.Errorw("update failed", "rows_affected", rowsAffected)
		return model.CashFlowEntity{}
	}

	return updatedEntity
}

func (CashFlowMongoDbMapper) DeleteCashFlowByObjectId(plainId string) model.CashFlowEntity {
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("cash_flow's id is not acceptable")
		return model.CashFlowEntity{}
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()
	targetEntity := convertBsonM2CashFlowEntity(database.GetOneInMongoDB(filter))
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("cash_flow is not exist")
		return model.CashFlowEntity{}
	}
	
	// Soft delete: Update is_delete to true
	update := bson.D{
		primitive.E{Key: "is_delete", Value: true},
		primitive.E{Key: "delete_time", Value: time.Now()},
	}

	rowsAffected := database.UpdateManyInMongoDB(filter, update)
	if rowsAffected != 1 {
		// fixme: maybe we should have a rollback here.
		util.Logger.Errorw("delete failed", "rows_affected", rowsAffected)
		return model.CashFlowEntity{}
	}
	
	targetEntity.IsDelete = true
	now := time.Now().UTC()
	targetEntity.DeleteTime = &now
	return targetEntity
}

func (CashFlowMongoDbMapper) DeleteCashFlowByBelongsDate(belongsDate time.Time) []model.CashFlowEntity {
	filter := bson.D{
		primitive.E{Key: "belongs_date", Value: belongsDate},
	}

	cashFlowList := INSTANCE.GetCashFlowsByBelongsDate(belongsDate)
	if cashFlowList == nil {
		util.Logger.Infoln("no cash_flow(s) found")
		return []model.CashFlowEntity{}
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	// Soft delete: Update is_delete to true
	update := bson.D{
		primitive.E{Key: "is_delete", Value: true},
		primitive.E{Key: "delete_time", Value: time.Now()},
	}

	rowsAffected := database.UpdateManyInMongoDB(filter, update)
	if rowsAffected != int64(len(cashFlowList)) {
		// fixme: maybe we should have a rollback here.
		util.Logger.Errorw("delete failed", "rows_affected", rowsAffected)
	}
	
	now := time.Now().UTC()
	for i := range cashFlowList {
		cashFlowList[i].IsDelete = true
		cashFlowList[i].DeleteTime = &now
	}
	return cashFlowList
}

func (CashFlowMongoDbMapper) GetAllCashFlows(limit, offset int) []model.CashFlowEntity {
	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	// Filter out deleted records (now handled by utility)
	filter := bson.D{}

	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDBWithPagination(filter, int64(limit), int64(offset))
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}

	return targetEntityList
}

func (CashFlowMongoDbMapper) GetAllCashFlowsIncludeDeleted(limit, offset int) []model.CashFlowEntity {
	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	filter := bson.D{}

	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDBWithPaginationIncludeDeleted(filter, int64(limit), int64(offset))
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}

	return targetEntityList
}

func (CashFlowMongoDbMapper) CountAllCashFlows() int64 {
	filter := bson.D{}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	return database.CountInMongoDB(filter)
}

func (CashFlowMongoDbMapper) TruncateCashFlows() error {
	// Open database connection
	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	// Empty filter to delete all documents
	filter := bson.D{}

	// Delete all documents (including soft deleted ones)
	deletedCount := database.DeleteManyInMongoDBIncludeDeleted(filter)

	util.Logger.Infow("Cash flows truncated successfully", "deleted_count", deletedCount)
	return nil
}

// User-specific methods for data isolation

func (CashFlowMongoDbMapper) GetCashFlowByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CashFlowEntity {
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("cash_flow's id is not acceptable")
		return model.CashFlowEntity{}
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
		primitive.E{Key: "user_id", Value: userId},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()
	return convertBsonM2CashFlowEntity(database.GetOneInMongoDB(filter))
}

func (CashFlowMongoDbMapper) GetCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	filter := bson.D{
		primitive.E{Key: "belongs_date", Value: belongsDate},
		primitive.E{Key: "user_id", Value: userId},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}
	return targetEntityList
}

func (CashFlowMongoDbMapper) GetCashFlowsByDateRangeAndUser(from, to time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	filter := bson.D{
		primitive.E{Key: "belongs_date", Value: bson.M{
			"$gte": from,
			"$lte": to,
		}},
		primitive.E{Key: "user_id", Value: userId},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}
	return targetEntityList
}

func (CashFlowMongoDbMapper) GetCashFlowsByCategoryIdAndUser(categoryPlainId string, userId primitive.ObjectID) []model.CashFlowEntity {
	categoryObjectId := util.Convert2ObjectId(categoryPlainId)
	if categoryPlainId == "" || categoryObjectId == primitive.NilObjectID {
		util.Logger.Warnln("category's id is not acceptable")
		return nil
	}

	filter := bson.D{
		primitive.E{Key: "category_id", Value: categoryObjectId},
		primitive.E{Key: "user_id", Value: userId},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}
	return targetEntityList
}

func (CashFlowMongoDbMapper) GetAllCashFlowsByUser(userId primitive.ObjectID, limit, offset int) []model.CashFlowEntity {
	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	filter := bson.D{
		primitive.E{Key: "user_id", Value: userId},
	}

	var targetEntityList []model.CashFlowEntity
	queryResultList := database.GetManyInMongoDBWithPagination(filter, int64(limit), int64(offset))
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CashFlowEntity(queryResult))
	}

	return targetEntityList
}

func (CashFlowMongoDbMapper) CountAllCashFlowsByUser(userId primitive.ObjectID) int64 {
	filter := bson.D{
		primitive.E{Key: "user_id", Value: userId},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	return database.CountInMongoDB(filter)
}

func (CashFlowMongoDbMapper) DeleteCashFlowByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CashFlowEntity {
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("cash_flow's id is not acceptable")
		return model.CashFlowEntity{}
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
		primitive.E{Key: "user_id", Value: userId},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()
	targetEntity := convertBsonM2CashFlowEntity(database.GetOneInMongoDB(filter))
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("cash_flow is not exist or does not belong to user")
		return model.CashFlowEntity{}
	}
	// Soft delete: Update is_delete to true
	now := time.Now()
	update := bson.D{
		primitive.E{Key: "is_delete", Value: true},
		primitive.E{Key: "delete_time", Value: now},
		primitive.E{Key: "delete_user_id", Value: userId},
	}

	rowsAffected := database.UpdateManyInMongoDB(filter, update)
	if rowsAffected != 1 {
		util.Logger.Errorw("delete failed", "rows_affected", rowsAffected)
		return model.CashFlowEntity{}
	}
	
	targetEntity.IsDelete = true
	targetEntity.DeleteTime = &now
	targetEntity.DeleteUserId = &userId
	return targetEntity
}

func (CashFlowMongoDbMapper) DeleteCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	filter := bson.D{
		primitive.E{Key: "belongs_date", Value: belongsDate},
		primitive.E{Key: "user_id", Value: userId},
	}

	cashFlowList := INSTANCE.GetCashFlowsByBelongsDateAndUser(belongsDate, userId)
	if cashFlowList == nil {
		util.Logger.Infoln("no cash_flow(s) found")
		return []model.CashFlowEntity{}
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	// Soft delete: Update is_delete to true
	update := bson.D{
		primitive.E{Key: "is_delete", Value: true},
		primitive.E{Key: "delete_time", Value: time.Now()},
		primitive.E{Key: "delete_user_id", Value: userId},
	}

	rowsAffected := database.UpdateManyInMongoDB(filter, update)
	if rowsAffected != int64(len(cashFlowList)) {
		util.Logger.Errorw("delete failed", "rows_affected", rowsAffected)
	}
	return cashFlowList
}

func (CashFlowMongoDbMapper) DeleteCashFlowsByCategoryIdAndUser(categoryPlainId string, userId primitive.ObjectID) int64 {
	categoryObjectId := util.Convert2ObjectId(categoryPlainId)
	if categoryPlainId == "" || categoryObjectId == primitive.NilObjectID {
		util.Logger.Warnln("category's id is not acceptable")
		return 0
	}

	filter := bson.D{
		primitive.E{Key: "category_id", Value: categoryObjectId},
		primitive.E{Key: "user_id", Value: userId},
	}

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	// Soft delete: Update is_delete to true
	update := bson.D{
		primitive.E{Key: "is_delete", Value: true},
		primitive.E{Key: "delete_time", Value: time.Now()},
		primitive.E{Key: "delete_user_id", Value: userId},
	}

	return database.UpdateManyInMongoDB(filter, update)
}

// Helper functions

func convertCashFlowEntity2BsonD(entity model.CashFlowEntity) bson.D {
	// Generate a new Id automatically if it's empty
	if entity.Id == primitive.NilObjectID {
		entity.Id = primitive.NewObjectID()
	}

	return bson.D{
		primitive.E{Key: "_id", Value: entity.Id},
		primitive.E{Key: "user_id", Value: entity.UserId},
		primitive.E{Key: "category_id", Value: entity.CategoryId},
		primitive.E{Key: "belongs_date", Value: entity.BelongsDate},
		// primitive.E{Key: "flow_type", Value: entity.FlowType},
		primitive.E{Key: "amount", Value: entity.Amount},
		primitive.E{Key: "description", Value: entity.Description},
		primitive.E{Key: "remark", Value: entity.Remark},
		primitive.E{Key: "create_user_id", Value: entity.CreateUserId},
		primitive.E{Key: "create_time", Value: entity.CreateTime},
		primitive.E{Key: "update_user_id", Value: entity.UpdateUserId},
		primitive.E{Key: "update_time", Value: entity.UpdateTime},
		primitive.E{Key: "delete_user_id", Value: entity.DeleteUserId},
		primitive.E{Key: "delete_time", Value: entity.DeleteTime},
		primitive.E{Key: "is_delete", Value: entity.IsDelete},
	}
}

func convertBsonM2CashFlowEntity(bsonM bson.M) model.CashFlowEntity {
	var newEntity model.CashFlowEntity
	bsonBytes, err := bson.Marshal(bsonM)
	if err != nil {
		util.Logger.Errorln(err)
		panic(err)
	}
	if err = bson.Unmarshal(bsonBytes, &newEntity); err != nil {
		util.Logger.Errorln(err)
		panic(err)
	}
	
	// Ensure BaseEntity fields are correctly mapped if not handled by bson tags
	
	return newEntity
}