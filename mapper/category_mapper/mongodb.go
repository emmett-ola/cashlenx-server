package category_mapper

import (
	"context"
	"time"

	"github.com/macar-x/cashlenx-server/cache"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CategoryMongoDbMapper struct{}

func (CategoryMongoDbMapper) GetCategoryByObjectId(plainId string) model.CategoryEntity {
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("category's id is not acceptable")
		return model.CategoryEntity{}
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()
	return convertBsonM2CategoryEntity(database.GetOneInMongoDB(filter))
}

func (CategoryMongoDbMapper) GetCategoryByName(categoryName string) model.CategoryEntity {
	filter := bson.D{
		primitive.E{Key: "name", Value: categoryName},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()
	return convertBsonM2CategoryEntity(database.GetOneInMongoDB(filter))
}

func (CategoryMongoDbMapper) GetCategoriesByUserAndType(userObjectId primitive.ObjectID, categoryType string, page, pageSize int) ([]model.CategoryEntity, error) {
	filter := bson.D{
		primitive.E{Key: "belongs_user_id", Value: userObjectId},
		primitive.E{Key: "type", Value: categoryType},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	// Calculate skip for pagination
	skip := (page - 1) * pageSize

	// Find documents with pagination
	results := database.GetManyInMongoDBWithPagination(filter, int64(skip), int64(pageSize))

	// No error handling needed as GetManyInMongoDBWithPagination handles it internally

	// Convert to CategoryEntity slice
	categories := make([]model.CategoryEntity, 0, len(results))
	for _, result := range results {
		categories = append(categories, convertBsonM2CategoryEntity(result))
	}

	return categories, nil
}
func (CategoryMongoDbMapper) GetCategoryByParentId(parentPlainId string) []model.CategoryEntity {
	parentObjectId := util.Convert2ObjectId(parentPlainId)
	filter := bson.D{
		primitive.E{Key: "parent_id", Value: parentObjectId},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	var targetEntityList []model.CategoryEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CategoryEntity(queryResult))
	}

	return targetEntityList
}

func (CategoryMongoDbMapper) InsertCategoryByEntity(newEntity model.CategoryEntity) string {
	// Only set CreateTime and UpdateTime if they're not already set (e.g., during restoration)
	operatingTime := time.Now().UTC() // Store in UTC
	if newEntity.CreateTime.IsZero() {
		newEntity.CreateTime = operatingTime
	}
	if newEntity.UpdateTime.IsZero() {
		newEntity.UpdateTime = operatingTime
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	newCategoryId := database.InsertOneInMongoDB(convertCategoryEntity2BsonD(newEntity))
	newEntity.Id = newCategoryId
	cache.GetCategoryCache().Set(&newEntity)

	return newCategoryId.Hex()
}

func (CategoryMongoDbMapper) UpdateCategoryByEntity(plainId string, updatedEntity model.CategoryEntity) model.CategoryEntity {
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("category's id is not acceptable")
		return model.CategoryEntity{}
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	targetEntity := convertBsonM2CategoryEntity(database.GetOneInMongoDB(filter))
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("category is not exist")
		return model.CategoryEntity{}
	}

	// Update fields from updatedEntity while preserving ID and CreateTime
	updatedEntity.Id = targetEntity.Id
	updatedEntity.CreateTime = targetEntity.CreateTime
	updatedEntity.UpdateTime = time.Now().UTC() // Store in UTC

	rowsAffected := database.UpdateManyInMongoDB(filter, convertCategoryEntity2BsonD(updatedEntity))
	if rowsAffected != 1 {
		// fixme: maybe we should have a rollback here.
		util.Logger.Errorw("update failed", "rows_affected", rowsAffected)
		return model.CategoryEntity{}
	}

	cache.GetCategoryCache().InvalidateByID(plainId)
	cache.GetCategoryCache().Set(&updatedEntity)

	return updatedEntity
}

func (CategoryMongoDbMapper) DeleteCategoryByObjectId(plainId string) model.CategoryEntity {
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("category's id is not acceptable")
		return model.CategoryEntity{}
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	targetEntity := convertBsonM2CategoryEntity(database.GetOneInMongoDB(filter))
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("category is not exist")
		return model.CategoryEntity{}
	}

	// can not delete a category that has referred child-categories.
	if len(INSTANCE.GetCategoryByParentId(plainId)) != 0 {
		util.Logger.Infoln("can not delete a category which has child-categories refer to")
		return model.CategoryEntity{}
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
		return model.CategoryEntity{}
	}

	cache.GetCategoryCache().InvalidateByID(plainId)

	return targetEntity
}

func (CategoryMongoDbMapper) GetAllCategories(limit, offset int) []model.CategoryEntity {
	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	// Filter out deleted records (handled by utility)
	filter := bson.D{}

	var targetEntityList []model.CategoryEntity
	queryResultList := database.GetManyInMongoDBWithPagination(filter, int64(limit), int64(offset))
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CategoryEntity(queryResult))
	}

	return targetEntityList
}

func (CategoryMongoDbMapper) GetAllCategoriesIncludeDeleted(limit, offset int) []model.CategoryEntity {
	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	filter := bson.D{}

	var targetEntityList []model.CategoryEntity
	queryResultList := database.GetManyInMongoDBWithPaginationIncludeDeleted(filter, int64(limit), int64(offset))
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CategoryEntity(queryResult))
	}

	return targetEntityList
}

func (CategoryMongoDbMapper) CountAllCategories() int64 {
	filter := bson.D{}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	return database.CountInMongoDB(filter)
}

// CountCategoriesByUserAndType counts categories filtered by user ID and type
func (CategoryMongoDbMapper) CountCategoriesByUserAndType(userObjectId primitive.ObjectID, categoryType string) (int64, error) {
	filter := bson.D{
		primitive.E{Key: "belongs_user_id", Value: userObjectId},
		primitive.E{Key: "type", Value: categoryType},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	count, err := database.CountInMongoDBWithError(filter)
	if err != nil {
		util.Logger.Errorw("Failed to count categories by user and type", "error", err)
		return 0, err
	}
	return count, nil
}

func (CategoryMongoDbMapper) GetRootCategoriesByUser(userId primitive.ObjectID) ([]model.CategoryEntity, error) {
	filter := bson.D{
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "parent_id", Value: bson.M{"$exists": false}}},
			bson.D{{Key: "parent_id", Value: primitive.NilObjectID}},
		}},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	results := database.GetManyInMongoDB(filter)

	categories := make([]model.CategoryEntity, 0, len(results))
	for _, result := range results {
		categories = append(categories, convertBsonM2CategoryEntity(result))
	}

	return categories, nil
}

func (CategoryMongoDbMapper) GetRootCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) {
	filter := bson.D{
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "type", Value: categoryType},
		primitive.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "parent_id", Value: bson.M{"$exists": false}}},
			bson.D{{Key: "parent_id", Value: primitive.NilObjectID}},
		}},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	results := database.GetManyInMongoDB(filter)

	categories := make([]model.CategoryEntity, 0, len(results))
	for _, result := range results {
		categories = append(categories, convertBsonM2CategoryEntity(result))
	}

	return categories, nil
}

func (CategoryMongoDbMapper) GetCategoriesByParentIdAndUser(parentId primitive.ObjectID, userId primitive.ObjectID) ([]model.CategoryEntity, error) {
	filter := bson.D{
		primitive.E{Key: "parent_id", Value: parentId},
		primitive.E{Key: "belongs_user_id", Value: userId},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	results := database.GetManyInMongoDB(filter)

	categories := make([]model.CategoryEntity, 0, len(results))
	for _, result := range results {
		categories = append(categories, convertBsonM2CategoryEntity(result))
	}

	return categories, nil
}

func (CategoryMongoDbMapper) GetCategoriesByParentIdUserAndType(parentId primitive.ObjectID, userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) {
	filter := bson.D{
		primitive.E{Key: "parent_id", Value: parentId},
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "type", Value: categoryType},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	results := database.GetManyInMongoDB(filter)

	categories := make([]model.CategoryEntity, 0, len(results))
	for _, result := range results {
		categories = append(categories, convertBsonM2CategoryEntity(result))
	}

	return categories, nil
}

func (CategoryMongoDbMapper) TruncateCategories() error {
	// Open database connection
	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	// Empty filter to delete all documents
	filter := bson.D{}

	// Delete all documents
	deletedCount := database.DeleteManyInMongoDBIncludeDeleted(filter)

	// Clear cache after truncate
	cache.GetCategoryCache().Clear()

	util.Logger.Infow("Categories truncated successfully", "deleted_count", deletedCount)
	return nil
}

// GetCategoryByObjectIdAndUser retrieves a category by ID, ensuring it belongs to the user
func (CategoryMongoDbMapper) GetCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity {
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("category's id is not acceptable")
		return model.CategoryEntity{}
	}
	if cached, ok := cache.GetCategoryCache().GetByID(plainId); ok && cached.BelongsUserId == userId {
		return *cached
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()
	entity := convertBsonM2CategoryEntity(database.GetOneInMongoDB(filter))
	if !entity.IsEmpty() {
		cache.GetCategoryCache().Set(&entity)
	}
	return entity
}

// GetCategoryByNameAndUser retrieves a category by name, ensuring it belongs to the user
func (CategoryMongoDbMapper) GetCategoryByNameAndUser(categoryName string, userId primitive.ObjectID) model.CategoryEntity {
	filter := bson.D{
		primitive.E{Key: "name", Value: categoryName},
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()
	return convertBsonM2CategoryEntity(database.GetOneInMongoDB(filter))
}

// GetCategoryByNameUserAndType retrieves a category by name, user and type
func (CategoryMongoDbMapper) GetCategoryByNameUserAndType(categoryName string, userId primitive.ObjectID, categoryType string) model.CategoryEntity {
	filter := bson.D{
		primitive.E{Key: "name", Value: categoryName},
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "type", Value: categoryType},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()
	return convertBsonM2CategoryEntity(database.GetOneInMongoDB(filter))
}

// GetCategoryByNameUserTypeAndParent retrieves a category by name, user, type and parent
func (CategoryMongoDbMapper) GetCategoryByNameUserTypeAndParent(categoryName string, userId primitive.ObjectID, categoryType string, parentId primitive.ObjectID) model.CategoryEntity {
	if cached, ok := cache.GetCategoryCache().GetByScope(userId.Hex(), categoryType, parentId.Hex(), categoryName); ok {
		return *cached
	}
	filter := bson.D{
		primitive.E{Key: "name", Value: categoryName},
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "type", Value: categoryType},
		primitive.E{Key: "is_delete", Value: false},
	}

	if parentId != primitive.NilObjectID {
		filter = append(filter, primitive.E{Key: "parent_id", Value: parentId})
	} else {
		// Explicitly check for null or non-existent parent_id for root categories
		filter = append(filter, primitive.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "parent_id", Value: bson.M{"$exists": false}}},
			bson.D{{Key: "parent_id", Value: primitive.NilObjectID}},
		}})
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()
	entity := convertBsonM2CategoryEntity(database.GetOneInMongoDB(filter))
	if !entity.IsEmpty() {
		cache.GetCategoryCache().Set(&entity)
	}
	return entity
}

// DeleteCategoryByObjectIdAndUser deletes a category by ID, ensuring it belongs to the user
func (CategoryMongoDbMapper) DeleteCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity {
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("category's id is not acceptable")
		return model.CategoryEntity{}
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	targetEntity := convertBsonM2CategoryEntity(database.GetOneInMongoDB(filter))
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("category is not exist or access denied")
		return model.CategoryEntity{}
	}

	// can not delete a category that has referred child-categories (user-specific check)
	childCategories, _ := INSTANCE.GetCategoriesByParentIdAndUser(objectId, userId)
	if len(childCategories) != 0 {
		util.Logger.Infoln("can not delete a category which has child-categories refer to")
		return model.CategoryEntity{}
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
		return model.CategoryEntity{}
	}

	cache.GetCategoryCache().InvalidateByID(plainId)

	targetEntity.IsDelete = true
	targetEntity.DeleteTime = &now
	targetEntity.DeleteUserId = &userId
	return targetEntity
}

// UpdateCategoryByEntityAndUser updates a category by ID, ensuring it belongs to the user
func (CategoryMongoDbMapper) UpdateCategoryByEntityAndUser(plainId string, updatedEntity model.CategoryEntity, userId primitive.ObjectID) model.CategoryEntity {
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("category's id is not acceptable")
		return model.CategoryEntity{}
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	targetEntity := convertBsonM2CategoryEntity(database.GetOneInMongoDB(filter))
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("category is not exist or access denied")
		return model.CategoryEntity{}
	}

	// Update fields from updatedEntity while preserving ID, UserId, and CreateTime
	updatedEntity.Id = targetEntity.Id
	updatedEntity.BelongsUserId = userId
	updatedEntity.CreateTime = targetEntity.CreateTime
	updatedEntity.UpdateTime = time.Now().UTC()

	rowsAffected := database.UpdateManyInMongoDB(filter, convertCategoryEntity2BsonD(updatedEntity))
	if rowsAffected != 1 {
		util.Logger.Errorw("update failed", "rows_affected", rowsAffected)
		return model.CategoryEntity{}
	}

	cache.GetCategoryCache().InvalidateByID(plainId)
	cache.GetCategoryCache().Set(&updatedEntity)

	return updatedEntity
}

// GetAllCategoriesByUser retrieves all categories for a specific user with pagination
func (CategoryMongoDbMapper) GetAllCategoriesByUser(userId primitive.ObjectID, limit, offset int) []model.CategoryEntity {
	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	collection := database.GetMongoCollection(database.CategoryTableName)

	filter := bson.D{
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "is_delete", Value: false},
	}

	ctx := context.TODO()
	findOptions := options.Find()
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	if offset > 0 {
		findOptions.SetSkip(int64(offset))
	}
	// Sort by name ascending
	findOptions.SetSort(bson.D{primitive.E{Key: "name", Value: 1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		util.Logger.Errorw("query all categories by user failed", "error", err)
		return []model.CategoryEntity{}
	}
	defer cursor.Close(ctx)

	var targetEntityList []model.CategoryEntity
	for cursor.Next(ctx) {
		var bsonM bson.M
		if err := cursor.Decode(&bsonM); err != nil {
			util.Logger.Errorw("decode failed", "error", err)
			continue
		}
		targetEntityList = append(targetEntityList, convertBsonM2CategoryEntity(bsonM))
	}

	return targetEntityList
}

// CountAllCategoriesByUser counts all categories for a specific user
func (CategoryMongoDbMapper) CountAllCategoriesByUser(userId primitive.ObjectID) int64 {
	filter := bson.D{
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	return database.CountInMongoDB(filter)
}

func (CategoryMongoDbMapper) GetAllCategoriesByUserIncludeDeleted(userId primitive.ObjectID) []model.CategoryEntity {
	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	filter := bson.D{
		primitive.E{Key: "belongs_user_id", Value: userId},
	}

	var targetEntityList []model.CategoryEntity
	queryResultList := database.GetManyInMongoDBWithPaginationIncludeDeleted(filter, 0, 0)
	for _, queryResult := range queryResultList {
		targetEntityList = append(targetEntityList, convertBsonM2CategoryEntity(queryResult))
	}

	return targetEntityList
}

func (CategoryMongoDbMapper) DeleteAllCategoriesByUser(userId primitive.ObjectID) (int64, error) {
	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	filter := bson.D{
		primitive.E{Key: "belongs_user_id", Value: userId},
	}

	return database.DeleteManyInMongoDBIncludeDeleted(filter), nil
}

func convertCategoryEntity2BsonD(entity model.CategoryEntity) bson.D {
	// Generate a new Id automatically if it's empty
	if entity.Id == primitive.NilObjectID {
		entity.Id = primitive.NewObjectID()
	}

	return bson.D{
		primitive.E{Key: "_id", Value: entity.Id},
		primitive.E{Key: "parent_id", Value: entity.ParentId},
		primitive.E{Key: "name", Value: entity.Name},
		primitive.E{Key: "type", Value: entity.Type},
		primitive.E{Key: "belongs_user_id", Value: entity.BelongsUserId},
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

func convertBsonM2CategoryEntity(bsonM bson.M) model.CategoryEntity {
	var newEntity model.CategoryEntity
	bsonBytes, err := bson.Marshal(bsonM)
	if err != nil {
		util.Logger.Errorln(err)
		panic(err)
	}
	if err = bson.Unmarshal(bsonBytes, &newEntity); err != nil {
		util.Logger.Errorln(err)
		panic(err)
	}

	// Manually map fields that might not be automatically mapped due to struct embedding differences or bson tags
	// Note: BaseEntity fields should be mapped by bson tags in the struct if bson.Unmarshal works correctly with inline
	// But let's be safe and check if we need manual mapping for older data or specific cases

	return newEntity
}
