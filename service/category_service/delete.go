package category_service

import (
	"errors"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteByIdForUser deletes a category by ID, ensuring it belongs to the user
func DeleteByIdForUser(plainId string, userId string, force bool) (model.CategoryEntity, error) {
	return defaultCategoryService().DeleteByIdForUser(plainId, userId, force)
}

func (s *CategoryService) DeleteByIdForUser(plainId string, userId string, force bool) (model.CategoryEntity, error) {
	// Validate ID
	if err := validation.ValidateID(plainId); err != nil {
		return model.CategoryEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return model.CategoryEntity{}, errors.New("invalid user ID")
	}

	// Check if it exists and belongs to user
	existCategoryEntity := s.categoryMapper.GetCategoryByObjectIdAndUser(plainId, userObjectId)
	if existCategoryEntity.IsEmpty() {
		return model.CategoryEntity{}, errors.New("category not found or access denied")
	}

	relatedFlows := s.cashFlowMapper.GetCashFlowsByCategoryIdAndUser(plainId, userObjectId)
	if len(relatedFlows) > 0 {
		if !force {
			return model.CategoryEntity{}, errors.New("cannot delete category: associated cash flows exist. use force=true to delete them first")
		}

		s.cashFlowMapper.DeleteCashFlowsByCategoryIdAndUser(plainId, userObjectId)
	}

	// Delete it
	deletedEntity := s.categoryMapper.DeleteCategoryByObjectIdAndUser(plainId, userObjectId)
	if deletedEntity.IsEmpty() {
		return model.CategoryEntity{}, errors.New("category delete failed")
	}
	return deletedEntity, nil
}
