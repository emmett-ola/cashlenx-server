package category_service

import (
	"errors"
	"strings"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateByIdForUser updates a category record by ID, ensuring it belongs to the user
func UpdateByIdForUser(plainId, name, categoryType, remark string, parentId string, userId string) (model.CategoryEntity, error) {
	// Validate ID
	if err := validation.ValidateID(plainId); err != nil {
		return model.CategoryEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return model.CategoryEntity{}, errors.New("invalid user ID")
	}

	// Validate optional fields if provided
	if name != "" {
		if err := validation.ValidateCategoryName(name); err != nil {
			return model.CategoryEntity{}, err
		}
	}

	if categoryType != "" {
		categoryType = strings.ToLower(categoryType)
		if err := validation.ValidateFlowType(categoryType); err != nil {
			return model.CategoryEntity{}, err
		}
	}

	// Query existing record - ensure it belongs to the user
	existingEntity := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(plainId, userObjectId)
	if existingEntity.IsEmpty() {
		return model.CategoryEntity{}, errors.New("category not found or access denied")
	}

	// Prevent changing category type
	if categoryType != "" && categoryType != existingEntity.Type {
		return model.CategoryEntity{}, errors.New("category type cannot be changed once created")
	}

	// Update fields that are provided
	if name != "" {
		// Check uniqueness if name is changing
		if name != existingEntity.Name {
			// Determine parent to check against (use existing parent if not provided)
			checkParentId := existingEntity.ParentId
			if parentId != "" {
				parentObjectId := util.Convert2ObjectId(parentId)
				if parentObjectId != primitive.NilObjectID {
					checkParentId = parentObjectId
				}
			}

			conflictCategory := category_mapper.INSTANCE.GetCategoryByNameUserTypeAndParent(name, userObjectId, existingEntity.Type, checkParentId)
			if !conflictCategory.IsEmpty() {
				return model.CategoryEntity{}, errors.New("category with this name already exists for this user and type under this parent")
			}
		}
		existingEntity.Name = name
	}

	if remark != "" {
		existingEntity.Remark = remark
	}

	if parentId != "" {
		parentObjectId := util.Convert2ObjectId(parentId)
		if parentObjectId != primitive.NilObjectID {
			// Verify parent category belongs to same user
			parentEntity := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(parentId, userObjectId)
			if parentEntity.IsEmpty() {
				return model.CategoryEntity{}, errors.New("parent category not found or access denied")
			}
			// Validate parent type matches existing category type
			if parentEntity.Type != existingEntity.Type {
				return model.CategoryEntity{}, errors.New("child category type must match parent category type")
			}
			existingEntity.ParentId = parentObjectId
		}
	}

	// Update modify time
	existingEntity.ModifyTime = time.Now().UTC()

	// Call mapper to update the record
	updatedEntity := category_mapper.INSTANCE.UpdateCategoryByEntityAndUser(plainId, existingEntity, userObjectId)
	if updatedEntity.IsEmpty() {
		return model.CategoryEntity{}, errors.New("failed to update category")
	}

	return updatedEntity, nil
}
