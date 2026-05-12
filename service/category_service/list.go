package category_service

import (
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QueryAllForUser queries all categories for a user with optional filtering and pagination
func QueryAllForUser(userId string, categoryType string, limit int, offset int) ([]model.CategoryEntity, int64, error) {
	return defaultCategoryService().QueryAllForUser(userId, categoryType, limit, offset)
}

func (s *CategoryService) QueryAllForUser(userId string, categoryType string, limit int, offset int) ([]model.CategoryEntity, int64, error) {
	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return nil, 0, errors.NewInvalidInputError("invalid user ID")
	}

	// Get total count for this user
	var totalCount int64

	// Get paginated results for this user
	var categories []model.CategoryEntity

	if categoryType != "" {
		// Get filtered count
		count, err := s.categoryMapper.CountCategoriesByUserAndType(userObjectId, categoryType)
		if err != nil {
			return nil, 0, err
		}
		totalCount = count

		// Filter by type if provided
		categoriesByType, err := s.categoryMapper.GetCategoriesByUserAndType(userObjectId, categoryType, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		categories = categoriesByType
	} else {
		// Get total count
		totalCount = s.categoryMapper.CountAllCategoriesByUser(userObjectId)

		// Get all categories for user
		categories = s.categoryMapper.GetAllCategoriesByUser(userObjectId, limit, offset)
	}

	return categories, totalCount, nil
}
