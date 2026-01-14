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

	// Update fields that are provided
	if name != "" {
		// Check uniqueness if name is changing
		if name != existingEntity.Name {
			// Determine type to check against (use existing type if not provided)
			checkType := existingEntity.Type
			if categoryType != "" {
				checkType = categoryType
			}
			
			// Determine parent to check against (use existing parent if not provided)
			checkParentId := existingEntity.ParentId
			if parentId != "" {
				parentObjectId := util.Convert2ObjectId(parentId)
				if parentObjectId != primitive.NilObjectID {
					checkParentId = parentObjectId
				}
			}

			conflictCategory := category_mapper.INSTANCE.GetCategoryByNameUserTypeAndParent(name, userObjectId, checkType, checkParentId)
			if !conflictCategory.IsEmpty() {
				return model.CategoryEntity{}, errors.New("category with this name already exists for this user and type under this parent")
			}
		}
		existingEntity.Name = name
	}

	if categoryType != "" {
		// Check uniqueness if type is changing (and name wasn't changed above, or was changed and checked)
		if categoryType != existingEntity.Type && name == "" {
			// Determine parent to check against (use existing parent if not provided)
			checkParentId := existingEntity.ParentId
			if parentId != "" {
				parentObjectId := util.Convert2ObjectId(parentId)
				if parentObjectId != primitive.NilObjectID {
					checkParentId = parentObjectId
				}
			}

			conflictCategory := category_mapper.INSTANCE.GetCategoryByNameUserTypeAndParent(existingEntity.Name, userObjectId, categoryType, checkParentId)
			if !conflictCategory.IsEmpty() {
				return model.CategoryEntity{}, errors.New("category with this name already exists for this user and type under this parent")
			}
		}
		existingEntity.Type = categoryType
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
			// Validate parent type matches existing or new category type
			checkType := existingEntity.Type
			if categoryType != "" {
				checkType = categoryType
			}
			if parentEntity.Type != checkType {
				return model.CategoryEntity{}, errors.New("child category type must match parent category type")
			}
			existingEntity.ParentId = parentObjectId
		}
	} else if categoryType != "" {
		// If only changing type (without changing parent), verify against existing parent if one exists
		if existingEntity.ParentId != primitive.NilObjectID {
			parentEntity := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(existingEntity.ParentId.Hex(), userObjectId)
			if !parentEntity.IsEmpty() {
				if parentEntity.Type != categoryType {
					return model.CategoryEntity{}, errors.New("child category type must match parent category type")
				}
			}
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
