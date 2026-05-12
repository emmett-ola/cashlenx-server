package cash_flow_service

import (
	"errors"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateById updates a cash flow record by ID
func UpdateById(plainId, belongsDate, categoryName string, amount float64, description string) (model.CashFlowEntity, error) {
	return defaultCashFlowService().UpdateById(plainId, belongsDate, categoryName, amount, description)
}

func (s *CashFlowService) UpdateById(plainId, belongsDate, categoryName string, amount float64, description string) (model.CashFlowEntity, error) {
	// Validate ID
	if err := validation.ValidateID(plainId); err != nil {
		return model.CashFlowEntity{}, err
	}

	// Validate optional fields if provided
	if belongsDate != "" {
		if err := validation.ValidateDate(belongsDate); err != nil {
			return model.CashFlowEntity{}, err
		}
	}

	if categoryName != "" {
		if err := validation.ValidateCategoryName(categoryName); err != nil {
			return model.CashFlowEntity{}, err
		}
	}

	if amount != 0 {
		if err := validation.ValidateAmount(amount); err != nil {
			return model.CashFlowEntity{}, err
		}
	}

	if description != "" {
		if err := validation.ValidateDescription(description); err != nil {
			return model.CashFlowEntity{}, err
		}
	}

	// Query existing record
	existingEntity := s.cashFlowMapper.GetCashFlowByObjectId(plainId)
	if existingEntity.IsEmpty() {
		return model.CashFlowEntity{}, errors.New("cash_flow not found")
	}

	// Update fields that are provided
	if belongsDate != "" {
		date := util.FormatDateFromStringWithoutDash(belongsDate)
		if date.IsZero() {
			return model.CashFlowEntity{}, errors.New("invalid date format")
		}
		existingEntity.BelongsDate = date
	}

	if categoryName != "" {
		categoryEntity := s.categoryMapper.GetCategoryByNameAndUser(categoryName, existingEntity.BelongsUserId)
		if categoryEntity.IsEmpty() {
			return model.CashFlowEntity{}, errors.New("category does not exist or access denied")
		}

		existingEntity.CategoryId = categoryEntity.Id
	}

	if amount != 0 {
		// Round to 2 decimal places
		amount, _ = decimal.NewFromFloat(amount).Round(2).Float64()
		existingEntity.Amount = amount
	}

	if description != "" {
		existingEntity.Description = description
	}

	// Update modify time
	existingEntity.UpdateTime = time.Now().UTC() // Store in UTC

	// Call mapper to update the record
	// Update audit fields
	existingEntity.UpdateTime = time.Now().UTC()
	// existingEntity.UpdateUserId = userObjectId // Undefined here

	updatedEntity := s.cashFlowMapper.UpdateCashFlowByEntity(plainId, existingEntity)
	if updatedEntity.IsEmpty() {
		return model.CashFlowEntity{}, errors.New("failed to update cash_flow")
	}

	// Populate category info
	category := s.categoryMapper.GetCategoryByObjectId(updatedEntity.CategoryId.Hex())
	if !category.IsEmpty() {
		updatedEntity.CategoryName = category.Name
		updatedEntity.CategoryType = category.Type
	} else {
		updatedEntity.CategoryName = "Unknown"
	}

	return updatedEntity, nil
}

// UpdateByIdForUser updates a cash flow record by ID, ensuring it belongs to the user
func UpdateByIdForUser(plainId, belongsDate, categoryName string, amount float64, description string, userId string) (model.CashFlowEntity, error) {
	return defaultCashFlowService().UpdateByIdForUser(plainId, belongsDate, categoryName, amount, description, userId)
}

func (s *CashFlowService) UpdateByIdForUser(plainId, belongsDate, categoryName string, amount float64, description string, userId string) (model.CashFlowEntity, error) {
	// Validate ID
	if err := validation.ValidateID(plainId); err != nil {
		return model.CashFlowEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return model.CashFlowEntity{}, errors.New("invalid user ID")
	}

	// Validate optional fields if provided
	if belongsDate != "" {
		if err := validation.ValidateDate(belongsDate); err != nil {
			return model.CashFlowEntity{}, err
		}
	}

	if categoryName != "" {
		if err := validation.ValidateCategoryName(categoryName); err != nil {
			return model.CashFlowEntity{}, err
		}
	}

	if amount != 0 {
		if err := validation.ValidateAmount(amount); err != nil {
			return model.CashFlowEntity{}, err
		}
	}

	if description != "" {
		if err := validation.ValidateDescription(description); err != nil {
			return model.CashFlowEntity{}, err
		}
	}

	// Query existing record - ensure it belongs to the user
	existingEntity := s.cashFlowMapper.GetCashFlowByObjectIdAndUser(plainId, userObjectId)
	if existingEntity.IsEmpty() {
		return model.CashFlowEntity{}, errors.New("cash_flow not found or access denied")
	}

	// Update fields that are provided
	if belongsDate != "" {
		date := util.FormatDateFromStringWithoutDash(belongsDate)
		if date.IsZero() {
			return model.CashFlowEntity{}, errors.New("invalid date format")
		}
		existingEntity.BelongsDate = date
	}

	if categoryName != "" {
		// Note: Category lookup should also be user-specific once categories have user isolation
		categoryEntity := s.categoryMapper.GetCategoryByName(categoryName)
		if categoryEntity.IsEmpty() {
			return model.CashFlowEntity{}, errors.New("category does not exist")
		}
		existingEntity.CategoryId = categoryEntity.Id
	}

	if amount != 0 {
		// Round to 2 decimal places
		roundedAmount, _ := decimal.NewFromFloat(amount).Round(2).Float64()
		existingEntity.Amount = roundedAmount
	}

	if description != "" {
		existingEntity.Description = description
	}

	// Update modify time
	existingEntity.UpdateTime = time.Now().UTC()
	existingEntity.UpdateUserId = userObjectId

	// Call mapper to update the record
	// Note: Using the regular UpdateCashFlowByEntity because we already verified ownership
	updatedEntity := s.cashFlowMapper.UpdateCashFlowByEntity(plainId, existingEntity)
	if updatedEntity.IsEmpty() {
		return model.CashFlowEntity{}, errors.New("failed to update cash_flow")
	}

	// Populate category info
	category := s.categoryMapper.GetCategoryByObjectId(updatedEntity.CategoryId.Hex())
	if !category.IsEmpty() {
		updatedEntity.CategoryName = category.Name
		updatedEntity.CategoryType = category.Type
	} else {
		updatedEntity.CategoryName = "Unknown"
	}

	return updatedEntity, nil
}
