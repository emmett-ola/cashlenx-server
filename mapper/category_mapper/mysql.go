package category_mapper

import (
	"bytes"
	"database/sql"
	"time"

	"github.com/macar-x/cashlenx-server/cache"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CategoryMySqlMapper struct{}

func (CategoryMySqlMapper) GetCategoryByObjectId(plainId string) model.CategoryEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, PARENT_ID, NAME, TYPE FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), plainId)
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
	}

	var categoryEntity model.CategoryEntity
	for rows.Next() {
		categoryEntity = convertRow2CategoryEntity(rows)
		break
	}
	return categoryEntity
}

func (CategoryMySqlMapper) GetCategoryByName(categoryName string) model.CategoryEntity {
	// Check cache first
	categoryCache := cache.GetCategoryCache()
	if cached, ok := categoryCache.GetByName(categoryName); ok {
		return *cached
	}

	// Cache miss - query database
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, PARENT_ID, NAME, TYPE FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE NAME = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), categoryName)
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
	}

	var categoryEntity model.CategoryEntity
	for rows.Next() {
		categoryEntity = convertRow2CategoryEntity(rows)
		break
	}

	// Store in cache if found
	if !categoryEntity.IsEmpty() {
		categoryCache.Set(&categoryEntity)
	}

	return categoryEntity
}

func (CategoryMySqlMapper) GetCategoryByParentId(parentPlainId string) []model.CategoryEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, PARENT_ID, NAME, TYPE FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE PARENT_ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), parentPlainId)
	if err != nil {
		util.Logger.Errorw("query failed", "error", err)
	}

	var targetEntityList []model.CategoryEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CategoryEntity(rows))
	}
	return targetEntityList
}

func (CategoryMySqlMapper) InsertCategoryByEntity(newEntity model.CategoryEntity) string {
	// Use existing ID if provided, otherwise generate a new one
	newPlainId := newEntity.Id.Hex()
	if newPlainId == "" {
		newPlainId = primitive.NewObjectID().Hex()
	}

	// Only set CreateTime and UpdateTime if they're not already set (e.g., during restoration)
	if newEntity.CreateTime.IsZero() || newEntity.UpdateTime.IsZero() {
		operatingTime := time.Now().UTC() // Store in UTC
		if newEntity.CreateTime.IsZero() {
			newEntity.CreateTime = operatingTime
		}
		if newEntity.UpdateTime.IsZero() {
			newEntity.UpdateTime = operatingTime
		}
	}

	var sqlString bytes.Buffer
	sqlString.WriteString("INSERT ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" SET ID = ?, ")
	sqlString.WriteString(" BELONGS_USER_ID = ?, ") // Added USER_ID
	sqlString.WriteString(" PARENT_ID = ?, ")
	sqlString.WriteString(" NAME = ?, ")
	sqlString.WriteString(" TYPE = ?, ")
	sqlString.WriteString(" REMARK = ?, ")
	sqlString.WriteString(" CREATE_USER_ID = ?, ")
	sqlString.WriteString(" CREATE_TIME = ?, ")
	sqlString.WriteString(" UPDATE_USER_ID = ?, ")
	sqlString.WriteString(" UPDATE_TIME = ?, ")
	sqlString.WriteString(" IS_DELETE = FALSE ")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("insert failed", "error", err)
	}

	result, err := statement.Exec(newPlainId, newEntity.BelongsUserId.Hex(), newEntity.ParentId.Hex(), newEntity.Name,
		newEntity.Type, newEntity.Remark, newEntity.CreateUserId.Hex(), newEntity.CreateTime, newEntity.UpdateUserId.Hex(), newEntity.UpdateTime)
	if err != nil {
		util.Logger.Errorw("insert failed", "error", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		// fixme: maybe we should have a rollback here.
		util.Logger.Errorw("insert failed", "error", err, "rows_affected", rowsAffected)
	}

	// Invalidate cache on insert
	cache.GetCategoryCache().Clear()

	return newPlainId
}

func (CategoryMySqlMapper) UpdateCategoryByEntity(plainId string, updatedEntity model.CategoryEntity) model.CategoryEntity {
	targetEntity := INSTANCE.GetCategoryByObjectId(plainId)
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("category is not exist")
		return model.CategoryEntity{}
	}

	// Update fields from updatedEntity while preserving ID and CreateTime
	updatedEntity.Id = targetEntity.Id
	updatedEntity.CreateTime = targetEntity.CreateTime
	updatedEntity.UpdateTime = time.Now().UTC() // Store in UTC

	var sqlString bytes.Buffer
	sqlString.WriteString("UPDATE ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" SET PARENT_ID = ?, ")
	sqlString.WriteString(" NAME = ?, ")
	sqlString.WriteString(" TYPE = ?, ")
	sqlString.WriteString(" REMARK = ?, ")
	sqlString.WriteString(" UPDATE_USER_ID = ?, ")
	sqlString.WriteString(" UPDATE_TIME = ? ")
	sqlString.WriteString(" WHERE ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("update failed", "error", err)
	}

	result, err := statement.Exec(updatedEntity.ParentId.Hex(), updatedEntity.Name, updatedEntity.Type, updatedEntity.Remark,
		updatedEntity.UpdateUserId.Hex(), updatedEntity.UpdateTime, updatedEntity.Id)
	if err != nil {
		util.Logger.Errorw("update failed", "error", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		// fixme: maybe we should have a rollback here.
		util.Logger.Errorw("update failed", "error", err, "rows_affected", rowsAffected)
	}

	// Invalidate cache on update
	cache.GetCategoryCache().Clear()

	return updatedEntity
}

func (CategoryMySqlMapper) DeleteCategoryByObjectId(plainId string) model.CategoryEntity {
	targetEntity := INSTANCE.GetCategoryByObjectId(plainId)
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("category is not exist")
		return model.CategoryEntity{}
	}

	// can not delete a category that has referred child-categories.
	if len(INSTANCE.GetCategoryByParentId(plainId)) != 0 {
		util.Logger.Infoln("can not delete a category which has child-categories refer to")
		return model.CategoryEntity{}
	}

	var sqlString bytes.Buffer
	sqlString.WriteString("UPDATE ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" SET IS_DELETE = TRUE, DELETE_TIME = NOW() ")
	sqlString.WriteString(" WHERE ID = ? ")
	sqlString.WriteString(database.SqlExcludeDeleted)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
	}

	result, err := statement.Exec(plainId)
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		// fixme: maybe we should have a rollback here.
		util.Logger.Errorw("delete failed", "error", err, "rows_affected", rowsAffected)
	}

	// Invalidate cache on delete
	cache.GetCategoryCache().Clear()

	targetEntity.IsDelete = true
	now := time.Now().UTC()
	targetEntity.DeleteTime = &now
	return targetEntity
}

func (CategoryMySqlMapper) GetAllCategories(limit, offset int) []model.CategoryEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, PARENT_ID, NAME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE TRUE ")
	sqlString.WriteString(database.SqlExcludeDeleted)
	sqlString.WriteString(" ORDER BY NAME ASC ")

	if limit > 0 {
		sqlString.WriteString(" LIMIT ? OFFSET ? ")
	}

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = connection.Query(sqlString.String(), limit, offset)
	} else {
		rows, err = connection.Query(sqlString.String())
	}

	if err != nil {
		util.Logger.Errorw("query all categories failed", "error", err)
		return []model.CategoryEntity{}
	}

	var targetEntityList []model.CategoryEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CategoryEntity(rows))
	}
	return targetEntityList
}

func (CategoryMySqlMapper) GetAllCategoriesIncludeDeleted(limit, offset int) []model.CategoryEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, PARENT_ID, NAME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE TRUE ")
	// No SqlExcludeDeleted
	sqlString.WriteString(" ORDER BY NAME ASC ")

	if limit > 0 {
		sqlString.WriteString(" LIMIT ? OFFSET ? ")
	}

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = connection.Query(sqlString.String(), limit, offset)
	} else {
		rows, err = connection.Query(sqlString.String())
	}

	if err != nil {
		util.Logger.Errorw("query all categories failed", "error", err)
		return []model.CategoryEntity{}
	}

	var targetEntityList []model.CategoryEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CategoryEntity(rows))
	}
	return targetEntityList
}

func (CategoryMySqlMapper) CountAllCategories() int64 {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT COUNT(1) FROM ")
	sqlString.WriteString(database.CategoryTableName)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String())
	if err != nil {
		util.Logger.Errorw("count all categories failed", "error", err)
		return 0
	}

	var count int64
	rows.Next()
	if err = rows.Scan(&count); err != nil {
		util.Logger.Errorw("parse count failed", "error", err)
		return 0
	}
	return count
}

// User-specific methods for data isolation

func (CategoryMySqlMapper) GetCategoriesByUserAndType(userId primitive.ObjectID, categoryType string, limit, offset int) ([]model.CategoryEntity, error) {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, BELONGS_USER_ID, PARENT_ID, NAME, TYPE, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE BELONGS_USER_ID = ? AND TYPE = ? AND IS_DELETE = FALSE ORDER BY NAME ASC")

	if limit > 0 {
		sqlString.WriteString(" LIMIT ? OFFSET ?")
	}

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = connection.Query(sqlString.String(), userId.Hex(), categoryType, limit, offset)
	} else {
		rows, err = connection.Query(sqlString.String(), userId.Hex(), categoryType)
	}

	if err != nil {
		util.Logger.Errorw("query categories by user and type failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	var categories []model.CategoryEntity
	for rows.Next() {
		categories = append(categories, convertRow2CategoryEntityWithUser(rows))
	}

	return categories, nil
}

func (CategoryMySqlMapper) CountCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) (int64, error) {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT COUNT(1) FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE BELONGS_USER_ID = ? AND TYPE = ? AND IS_DELETE = FALSE")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), userId.Hex(), categoryType)
	if err != nil {
		util.Logger.Errorw("count categories by user and type failed", "error", err)
		return 0, err
	}
	defer rows.Close()

	var count int64
	if rows.Next() {
		if err = rows.Scan(&count); err != nil {
			util.Logger.Errorw("parse count failed", "error", err)
			return 0, err
		}
	}
	return count, nil
}

func (CategoryMySqlMapper) GetRootCategoriesByUser(userId primitive.ObjectID) ([]model.CategoryEntity, error) {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, BELONGS_USER_ID, PARENT_ID, NAME, TYPE, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE BELONGS_USER_ID = ? AND (PARENT_ID = '' OR PARENT_ID IS NULL OR PARENT_ID = '000000000000000000000000') AND IS_DELETE = FALSE")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), userId.Hex())
	if err != nil {
		util.Logger.Errorw("query root categories by user failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	var categories []model.CategoryEntity
	for rows.Next() {
		categories = append(categories, convertRow2CategoryEntityWithUser(rows))
	}

	return categories, nil
}

func (CategoryMySqlMapper) GetRootCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, BELONGS_USER_ID, PARENT_ID, NAME, TYPE, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE BELONGS_USER_ID = ? AND TYPE = ? AND (PARENT_ID = '' OR PARENT_ID IS NULL OR PARENT_ID = '000000000000000000000000') AND IS_DELETE = FALSE")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), userId.Hex(), categoryType)
	if err != nil {
		util.Logger.Errorw("query root categories by user and type failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	var categories []model.CategoryEntity
	for rows.Next() {
		categories = append(categories, convertRow2CategoryEntityWithUser(rows))
	}

	return categories, nil
}

func (CategoryMySqlMapper) GetCategoriesByParentIdAndUser(parentId primitive.ObjectID, userId primitive.ObjectID) ([]model.CategoryEntity, error) {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, BELONGS_USER_ID, PARENT_ID, NAME, TYPE, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE PARENT_ID = ? AND BELONGS_USER_ID = ? AND IS_DELETE = FALSE")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), parentId.Hex(), userId.Hex())
	if err != nil {
		util.Logger.Errorw("query categories by parent and user failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	var categories []model.CategoryEntity
	for rows.Next() {
		categories = append(categories, convertRow2CategoryEntityWithUser(rows))
	}

	return categories, nil
}

func (CategoryMySqlMapper) GetCategoriesByParentIdUserAndType(parentId primitive.ObjectID, userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, BELONGS_USER_ID, PARENT_ID, NAME, TYPE, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE PARENT_ID = ? AND BELONGS_USER_ID = ? AND TYPE = ? AND IS_DELETE = FALSE")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), parentId.Hex(), userId.Hex(), categoryType)
	if err != nil {
		util.Logger.Errorw("query categories by parent, user and type failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	var categories []model.CategoryEntity
	for rows.Next() {
		categories = append(categories, convertRow2CategoryEntityWithUser(rows))
	}

	return categories, nil
}

func (CategoryMySqlMapper) GetCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, BELONGS_USER_ID, PARENT_ID, NAME, TYPE, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE ID = ? AND BELONGS_USER_ID = ? AND IS_DELETE = FALSE")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), plainId, userId.Hex())
	if err != nil {
		util.Logger.Errorw("query category by id and user failed", "error", err)
		return model.CategoryEntity{}
	}
	defer rows.Close()

	var categoryEntity model.CategoryEntity
	if rows.Next() {
		categoryEntity = convertRow2CategoryEntityWithUser(rows)
	}
	return categoryEntity
}

func (CategoryMySqlMapper) GetCategoryByNameAndUser(categoryName string, userId primitive.ObjectID) model.CategoryEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, BELONGS_USER_ID, PARENT_ID, NAME, TYPE, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE NAME = ? AND BELONGS_USER_ID = ? AND IS_DELETE = FALSE")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), categoryName, userId.Hex())
	if err != nil {
		util.Logger.Errorw("query category by name and user failed", "error", err)
		return model.CategoryEntity{}
	}
	defer rows.Close()

	var categoryEntity model.CategoryEntity
	if rows.Next() {
		categoryEntity = convertRow2CategoryEntityWithUser(rows)
	}
	return categoryEntity
}

func (CategoryMySqlMapper) GetCategoryByNameUserAndType(categoryName string, userId primitive.ObjectID, categoryType string) model.CategoryEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, BELONGS_USER_ID, PARENT_ID, NAME, TYPE, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE NAME = ? AND BELONGS_USER_ID = ? AND TYPE = ? AND IS_DELETE = FALSE")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), categoryName, userId.Hex(), categoryType)
	if err != nil {
		util.Logger.Errorw("query category by name, user and type failed", "error", err)
		return model.CategoryEntity{}
	}
	defer rows.Close()

	var categoryEntity model.CategoryEntity
	if rows.Next() {
		categoryEntity = convertRow2CategoryEntityWithUser(rows)
	}
	return categoryEntity
}

func (CategoryMySqlMapper) GetCategoryByNameUserTypeAndParent(categoryName string, userId primitive.ObjectID, categoryType string, parentId primitive.ObjectID) model.CategoryEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, BELONGS_USER_ID, PARENT_ID, NAME, TYPE, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE NAME = ? AND BELONGS_USER_ID = ? AND TYPE = ? AND IS_DELETE = FALSE")

	var rows *sql.Rows
	var err error

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	if parentId != primitive.NilObjectID {
		sqlString.WriteString(" AND PARENT_ID = ?")
		rows, err = connection.Query(sqlString.String(), categoryName, userId.Hex(), categoryType, parentId.Hex())
	} else {
		sqlString.WriteString(" AND (PARENT_ID = '' OR PARENT_ID IS NULL OR PARENT_ID = '000000000000000000000000')")
		rows, err = connection.Query(sqlString.String(), categoryName, userId.Hex(), categoryType)
	}

	if err != nil {
		util.Logger.Errorw("query category by name, user, type and parent failed", "error", err)
		return model.CategoryEntity{}
	}
	defer rows.Close()

	var categoryEntity model.CategoryEntity
	if rows.Next() {
		categoryEntity = convertRow2CategoryEntityWithUser(rows)
	}
	return categoryEntity
}

func (CategoryMySqlMapper) DeleteCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity {
	targetEntity := INSTANCE.GetCategoryByObjectIdAndUser(plainId, userId)
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("category is not exist or access denied")
		return model.CategoryEntity{}
	}

	// can not delete a category that has referred child-categories (user-specific check)
	childCategories, _ := INSTANCE.GetCategoriesByParentIdAndUser(targetEntity.Id, userId)
	if len(childCategories) != 0 {
		util.Logger.Infoln("can not delete a category which has child-categories refer to")
		return model.CategoryEntity{}
	}

	var sqlString bytes.Buffer
	sqlString.WriteString("UPDATE ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" SET IS_DELETE = TRUE, DELETE_TIME = NOW(), DELETE_USER_ID = ? ")
	sqlString.WriteString(" WHERE ID = ? AND BELONGS_USER_ID = ?")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
		return model.CategoryEntity{}
	}

	result, err := statement.Exec(userId.Hex(), plainId, userId.Hex())
	if err != nil {
		util.Logger.Errorw("delete failed", "error", err)
		return model.CategoryEntity{}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		util.Logger.Errorw("delete failed", "error", err, "rows_affected", rowsAffected)
		return model.CategoryEntity{}
	}

	// Invalidate cache on delete
	cache.GetCategoryCache().Clear()

	targetEntity.IsDelete = true
	now := time.Now().UTC()
	targetEntity.DeleteTime = &now
	targetEntity.DeleteUserId = &userId
	return targetEntity
}

func (CategoryMySqlMapper) UpdateCategoryByEntityAndUser(plainId string, updatedEntity model.CategoryEntity, userId primitive.ObjectID) model.CategoryEntity {
	targetEntity := INSTANCE.GetCategoryByObjectIdAndUser(plainId, userId)
	if targetEntity.IsEmpty() {
		util.Logger.Infoln("category is not exist or access denied")
		return model.CategoryEntity{}
	}

	// Update fields from updatedEntity while preserving ID, UserId, and CreateTime
	updatedEntity.Id = targetEntity.Id
	updatedEntity.BelongsUserId = userId
	updatedEntity.CreateTime = targetEntity.CreateTime
	updatedEntity.UpdateTime = time.Now().UTC()

	var sqlString bytes.Buffer
	sqlString.WriteString("UPDATE ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" SET PARENT_ID = ?, ")
	sqlString.WriteString(" NAME = ?, ")
	sqlString.WriteString(" TYPE = ?, ")
	sqlString.WriteString(" REMARK = ?, ")
	sqlString.WriteString(" UPDATE_USER_ID = ?, ")
	sqlString.WriteString(" UPDATE_TIME = ? ")
	sqlString.WriteString(" WHERE ID = ? AND BELONGS_USER_ID = ? AND IS_DELETE = FALSE")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	statement, err := connection.Prepare(sqlString.String())
	if err != nil {
		util.Logger.Errorw("update failed", "error", err)
		return model.CategoryEntity{}
	}

	result, err := statement.Exec(updatedEntity.ParentId.Hex(), updatedEntity.Name, updatedEntity.Type,
		updatedEntity.Remark, updatedEntity.UpdateUserId.Hex(), updatedEntity.UpdateTime, updatedEntity.Id.Hex(), userId.Hex())
	if err != nil {
		util.Logger.Errorw("update failed", "error", err)
		return model.CategoryEntity{}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected != 1 {
		util.Logger.Errorw("update failed", "error", err, "rows_affected", rowsAffected)
		return model.CategoryEntity{}
	}

	// Invalidate cache on update
	cache.GetCategoryCache().Clear()

	return updatedEntity
}

func (CategoryMySqlMapper) GetAllCategoriesByUser(userId primitive.ObjectID, limit, offset int) []model.CategoryEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, BELONGS_USER_ID, PARENT_ID, NAME, TYPE, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE BELONGS_USER_ID = ? AND IS_DELETE = FALSE ORDER BY NAME ASC")

	if limit > 0 {
		sqlString.WriteString(" LIMIT ? OFFSET ?")
	}

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var rows *sql.Rows
	var err error

	if limit > 0 {
		rows, err = connection.Query(sqlString.String(), userId.Hex(), limit, offset)
	} else {
		rows, err = connection.Query(sqlString.String(), userId.Hex())
	}

	if err != nil {
		util.Logger.Errorw("query all categories by user failed", "error", err)
		return []model.CategoryEntity{}
	}
	defer rows.Close()

	var targetEntityList []model.CategoryEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CategoryEntityWithUser(rows))
	}
	return targetEntityList
}

func (CategoryMySqlMapper) CountAllCategoriesByUser(userId primitive.ObjectID) int64 {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT COUNT(1) FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE BELONGS_USER_ID = ? AND IS_DELETE = FALSE")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), userId.Hex())
	if err != nil {
		util.Logger.Errorw("count all categories by user failed", "error", err)
		return 0
	}
	defer rows.Close()

	var count int64
	if rows.Next() {
		if err = rows.Scan(&count); err != nil {
			util.Logger.Errorw("parse count failed", "error", err)
			return 0
		}
	}
	return count
}

func (CategoryMySqlMapper) TruncateCategories() error {
	var sqlString bytes.Buffer
	sqlString.WriteString("TRUNCATE TABLE ")
	sqlString.WriteString(database.CategoryTableName)

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	_, err := connection.Exec(sqlString.String())
	if err != nil {
		util.Logger.Errorw("truncate categories failed", "error", err)
		return err
	}

	// Clear cache after truncate
	cache.GetCategoryCache().Clear()

	util.Logger.Infow("Categories truncated successfully")
	return nil
}

func (CategoryMySqlMapper) GetAllCategoriesByUserIncludeDeleted(userId primitive.ObjectID) []model.CategoryEntity {
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT ID, BELONGS_USER_ID, PARENT_ID, NAME, TYPE, REMARK, CREATE_USER_ID, CREATE_TIME, UPDATE_USER_ID, UPDATE_TIME FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE BELONGS_USER_ID = ? ORDER BY NAME ASC")
	// No SqlExcludeDeleted

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	rows, err := connection.Query(sqlString.String(), userId.Hex())
	if err != nil {
		util.Logger.Errorw("query all categories by user include deleted failed", "error", err)
		return []model.CategoryEntity{}
	}
	defer rows.Close()

	var targetEntityList []model.CategoryEntity
	for rows.Next() {
		targetEntityList = append(targetEntityList, convertRow2CategoryEntityWithUser(rows))
	}
	return targetEntityList
}

func (CategoryMySqlMapper) DeleteAllCategoriesByUser(userId primitive.ObjectID) (int64, error) {
	var sqlString bytes.Buffer
	sqlString.WriteString("DELETE FROM ")
	sqlString.WriteString(database.CategoryTableName)
	sqlString.WriteString(" WHERE BELONGS_USER_ID = ? ")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	result, err := connection.Exec(sqlString.String(), userId.Hex())
	if err != nil {
		util.Logger.Errorw("delete all categories by user failed", "error", err)
		return 0, err
	}

	// Clear cache after delete
	cache.GetCategoryCache().Clear()

	return result.RowsAffected()
}

func convertRow2CategoryEntity(rows *sql.Rows) model.CategoryEntity {
	var id string
	var parentId string
	var name string
	var categoryType string

	err := rows.Scan(&id, &parentId, &name, &categoryType)
	if err != nil {
		util.Logger.Errorw("covert into entity failed", "error", err)
	}

	return model.CategoryEntity{
		Id:       util.Convert2ObjectId(id),
		ParentId: util.Convert2ObjectId(parentId),
		Name:     name,
		Type:     categoryType,
	}
}

// convertRow2CategoryEntityWithUser converts SQL rows to CategoryEntity including all fields
func convertRow2CategoryEntityWithUser(rows *sql.Rows) model.CategoryEntity {
	var id, userId, parentId, name, categoryType, remark string
	var createTime, updateTime time.Time
	var createUserId, updateUserId string

	err := rows.Scan(&id, &userId, &parentId, &name, &categoryType, &remark, &createUserId, &createTime, &updateUserId, &updateTime)
	if err != nil {
		util.Logger.Errorw("convert into entity failed", "error", err)
	}

	return model.CategoryEntity{
		Id:            util.Convert2ObjectId(id),
		BelongsUserId: util.Convert2ObjectId(userId),
		ParentId:      util.Convert2ObjectId(parentId),
		Name:          name,
		Type:          categoryType,
		Remark:        remark,
		BaseEntity: model.BaseEntity{
			CreateUserId: util.Convert2ObjectId(createUserId),
			CreateTime:   createTime,
			UpdateUserId: util.Convert2ObjectId(updateUserId),
			UpdateTime:   updateTime,
		},
	}
}
