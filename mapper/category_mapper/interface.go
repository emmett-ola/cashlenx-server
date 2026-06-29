package category_mapper

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var INSTANCE CategoryMapper

type CategoryMapper interface {
	GetCategoryByObjectId(plainId string) model.CategoryEntity
	GetCategoryByName(categoryName string) model.CategoryEntity
	GetCategoryByParentId(parentPlainId string) []model.CategoryEntity
	InsertCategoryByEntity(newEntity model.CategoryEntity) string
	UpdateCategoryByEntity(plainId string, updatedEntity model.CategoryEntity) model.CategoryEntity
	GetAllCategories(limit, offset int) []model.CategoryEntity
	GetAllCategoriesIncludeDeleted(limit, offset int) []model.CategoryEntity
	CountAllCategories() int64
	CountCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) (int64, error)
	DeleteCategoryByObjectId(plainId string) model.CategoryEntity
	TruncateCategories() error
	// New methods added for user-specific category operations
	GetCategoriesByUserAndType(userId primitive.ObjectID, categoryType string, limit, offset int) ([]model.CategoryEntity, error)
	GetRootCategoriesByUser(userId primitive.ObjectID) ([]model.CategoryEntity, error)
	GetRootCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error)
	GetCategoriesByParentIdAndUser(parentId primitive.ObjectID, userId primitive.ObjectID) ([]model.CategoryEntity, error)
	GetCategoriesByParentIdUserAndType(parentId primitive.ObjectID, userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error)
	// Additional user-specific methods for data isolation
	GetCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity
	GetCategoryByNameAndUser(categoryName string, userId primitive.ObjectID) model.CategoryEntity
	GetCategoryByNameUserAndType(categoryName string, userId primitive.ObjectID, categoryType string) model.CategoryEntity
	GetCategoryByNameUserTypeAndParent(categoryName string, userId primitive.ObjectID, categoryType string, parentId primitive.ObjectID) model.CategoryEntity
	DeleteCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity
	UpdateCategoryByEntityAndUser(plainId string, updatedEntity model.CategoryEntity, userId primitive.ObjectID) model.CategoryEntity
	GetAllCategoriesByUser(userId primitive.ObjectID, limit, offset int) []model.CategoryEntity
	CountAllCategoriesByUser(userId primitive.ObjectID) int64
	GetAllCategoriesByUserIncludeDeleted(userId primitive.ObjectID) []model.CategoryEntity
	DeleteAllCategoriesByUser(userId primitive.ObjectID) (int64, error)
}

func init() {
	mapper, ok := categoryMapperForDatabase(util.GetConfigByKey("db.type"))
	if !ok {
		panic("database type not supported")
	}
	INSTANCE = mapper
}

func categoryMapperForDatabase(databaseType string) (CategoryMapper, bool) {
	switch databaseType {
	case "mongodb":
		return CategoryMongoDbMapper{}, true
	case "mysql":
		return CategoryMySqlMapper{}, true
	default:
		return nil, false
	}
}
