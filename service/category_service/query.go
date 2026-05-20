package category_service

import (
	"errors"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QueryByIdForUser retrieves a category by ID, ensuring it belongs to the user
func QueryByIdForUser(plainId string, userId string) (model.CategoryEntity, error) {
	return defaultCategoryService().QueryByIdForUser(plainId, userId)
}

func (s *CategoryService) QueryByIdForUser(plainId string, userId string) (model.CategoryEntity, error) {
	// Validate ID
	if err := validation.ValidateID(plainId); err != nil {
		return model.CategoryEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return model.CategoryEntity{}, errors.New("invalid user ID")
	}

	categoryEntity := s.categoryMapper.GetCategoryByObjectIdAndUser(plainId, userObjectId)
	if categoryEntity.IsEmpty() {
		return model.CategoryEntity{}, errors.New("category not found or access denied")
	}
	return categoryEntity, nil
}

// QueryByNameForUser retrieves a category by name, ensuring it belongs to the user
func QueryByNameForUser(categoryName string, userId string) (model.CategoryEntity, error) {
	return defaultCategoryService().QueryByNameForUser(categoryName, userId)
}

func (s *CategoryService) QueryByNameForUser(categoryName string, userId string) (model.CategoryEntity, error) {
	// Validate category name
	if err := validation.ValidateCategoryName(categoryName); err != nil {
		return model.CategoryEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return model.CategoryEntity{}, errors.New("invalid user ID")
	}

	categoryEntity := s.categoryMapper.GetCategoryByNameAndUser(categoryName, userObjectId)
	if categoryEntity.IsEmpty() {
		return model.CategoryEntity{}, errors.New("category not found or access denied")
	}
	return categoryEntity, nil
}

// GetRootCategoriesForUser retrieves root categories (no parent) for a specific user
func GetRootCategoriesForUser(userId string, categoryType string) ([]model.CategoryEntity, error) {
	return defaultCategoryService().GetRootCategoriesForUser(userId, categoryType)
}

func (s *CategoryService) GetRootCategoriesForUser(userId string, categoryType string) ([]model.CategoryEntity, error) {
	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return nil, errors.New("invalid user ID")
	}

	var categories []model.CategoryEntity
	var err error

	if categoryType != "" {
		categories, err = s.categoryMapper.GetRootCategoriesByUserAndType(userObjectId, categoryType)
	} else {
		categories, err = s.categoryMapper.GetRootCategoriesByUser(userObjectId)
	}

	if err != nil {
		return nil, err
	}

	return categories, nil
}

// GetChildCategoriesForUser retrieves child categories of a parent for a specific user
func GetChildCategoriesForUser(parentId string, userId string, categoryType string) ([]model.CategoryEntity, error) {
	return defaultCategoryService().GetChildCategoriesForUser(parentId, userId, categoryType)
}

func (s *CategoryService) GetChildCategoriesForUser(parentId string, userId string, categoryType string) ([]model.CategoryEntity, error) {
	// Validate parent ID
	if err := validation.ValidateID(parentId); err != nil {
		return nil, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return nil, errors.New("invalid user ID")
	}

	// Convert parent ID
	parentObjectId := util.Convert2ObjectId(parentId)
	if parentObjectId == primitive.NilObjectID {
		// If parent ID is nil (000000000000000000000000), return root categories
		return s.GetRootCategoriesForUser(userId, categoryType)
	}

	// Verify parent category belongs to user
	parentEntity := s.categoryMapper.GetCategoryByObjectIdAndUser(parentId, userObjectId)
	if parentEntity.IsEmpty() {
		return nil, errors.New("parent category not found or access denied")
	}

	var categories []model.CategoryEntity
	var err error

	if categoryType != "" {
		categories, err = s.categoryMapper.GetCategoriesByParentIdUserAndType(parentObjectId, userObjectId, categoryType)
	} else {
		categories, err = s.categoryMapper.GetCategoriesByParentIdAndUser(parentObjectId, userObjectId)
	}

	if err != nil {
		return nil, err
	}

	return categories, nil
}
