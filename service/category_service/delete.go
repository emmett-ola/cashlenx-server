package category_service

import (
	"errors"

	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteByIdForUser deletes a category by ID, ensuring it belongs to the user
func DeleteByIdForUser(plainId string, userId string) (model.CategoryEntity, error) {
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
	existCategoryEntity := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(plainId, userObjectId)
	if existCategoryEntity.IsEmpty() {
		return model.CategoryEntity{}, errors.New("category not found or access denied")
	}

	// Delete it
	deletedEntity := category_mapper.INSTANCE.DeleteCategoryByObjectIdAndUser(plainId, userObjectId)
	if deletedEntity.IsEmpty() {
		return model.CategoryEntity{}, errors.New("category delete failed")
	}
	return deletedEntity, nil
}
