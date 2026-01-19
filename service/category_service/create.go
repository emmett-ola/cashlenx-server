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

// CreateForUser creates a new category for a specific user
func CreateForUser(name, categoryType, remark string, parentId string, userId string) (model.CategoryEntity, error) {
	// Validate required fields
	if err := validation.ValidateCategoryName(name); err != nil {
		return model.CategoryEntity{}, err
	}

	// Normalize and validate category type
	categoryType = strings.ToLower(categoryType)
	if err := validation.ValidateFlowType(categoryType); err != nil {
		return model.CategoryEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return model.CategoryEntity{}, errors.New("invalid user ID")
	}

	// Check if category with same name already exists for this user and type under the same parent
	var parentObjId primitive.ObjectID
	if parentId != "" {
		parentObjId = util.Convert2ObjectId(parentId)
	} else {
		parentObjId = primitive.NilObjectID
	}

	existingCategory := category_mapper.INSTANCE.GetCategoryByNameUserTypeAndParent(name, userObjectId, categoryType, parentObjId)
	if !existingCategory.IsEmpty() {
		return model.CategoryEntity{}, errors.New("category with this name already exists for this user and type under this parent")
	}

	// Create new category entity
	newEntity := model.CategoryEntity{
		BelongsUserId: userObjectId,
		Name:          name,
		Type:          categoryType,
		Remark:        remark,
		BaseEntity: model.BaseEntity{
			CreateUserId: userObjectId,
			CreateTime:   time.Now().UTC(),
			UpdateUserId: userObjectId,
			UpdateTime:   time.Now().UTC(),
		},
	}

	// Handle parent category if provided
	if parentId != "" {
		parentObjectId := util.Convert2ObjectId(parentId)
		if parentObjectId != primitive.NilObjectID {
			// Verify parent category belongs to same user
			parentEntity := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(parentId, userObjectId)
			if parentEntity.IsEmpty() {
				return model.CategoryEntity{}, errors.New("parent category not found or access denied")
			}
			// Validate parent type matches new category type
			if parentEntity.Type != categoryType {
				return model.CategoryEntity{}, errors.New("child category type must match parent category type")
			}
			newEntity.ParentId = parentObjectId
		}
	}

	// Insert the category
	newId := category_mapper.INSTANCE.InsertCategoryByEntity(newEntity)
	if newId == "" {
		return model.CategoryEntity{}, errors.New("failed to create category")
	}

	// Retrieve and return the created category
	createdEntity := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(newId, userObjectId)
	return createdEntity, nil
}
