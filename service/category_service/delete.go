package category_service

import (
	"errors"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeleteByIdForUser deletes a category by ID, ensuring it belongs to the user
func DeleteByIdForUser(plainId string, userId string, force bool) (model.CategoryEntity, error) {
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

	// Check for associated cash flows
	// Since we don't have a direct CountCashFlowsByCategoryIdAndUser yet, we can check by getting them
	// Or ideally we should add that method to the mapper.
	// But wait, cash_flows are user-isolated, so we should check if this category is used by ANY cash flow of this user.
	// Let's use GetCashFlowsByCategoryIdAndUser which we already have, and check length.
	// Optimization: If list is large, count is better.
	// But we just added DeleteCashFlowsByCategoryIdAndUser to interface, let's assume we can also add Count.
	// For now, let's use GetCashFlowsByCategoryIdAndUser to check existence.
	relatedFlows := cash_flow_mapper.INSTANCE.GetCashFlowsByCategoryIdAndUser(plainId, userObjectId)
	if len(relatedFlows) > 0 {
		if !force {
			return model.CategoryEntity{}, errors.New("cannot delete category: associated cash flows exist. use force=true to delete them first")
		}

		// Force delete: delete associated cash flows first
		cash_flow_mapper.INSTANCE.DeleteCashFlowsByCategoryIdAndUser(plainId, userObjectId)
	}

	// Delete it
	deletedEntity := category_mapper.INSTANCE.DeleteCategoryByObjectIdAndUser(plainId, userObjectId)
	if deletedEntity.IsEmpty() {
		return model.CategoryEntity{}, errors.New("category delete failed")
	}
	return deletedEntity, nil
}
