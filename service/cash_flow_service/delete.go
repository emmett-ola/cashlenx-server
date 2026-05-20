package cash_flow_service

import (
	"errors"
	"reflect"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func IsDeleteFieldsConflicted(plainId, belongsDate string) bool {
	// check if already one semi-optional field is filled
	semiOptionalFieldFilledFlag := false

	// plain_id is not empty
	if plainId != "" {
		semiOptionalFieldFilledFlag = true
	}

	// belongs_date is not empty
	if belongsDate != "" {
		if semiOptionalFieldFilledFlag {
			return true
		}
		semiOptionalFieldFilledFlag = true
	}

	// should have one and only one field filled
	return !semiOptionalFieldFilledFlag
}

func DeleteById(plainId string) (model.CashFlowEntity, error) {
	return defaultCashFlowService().DeleteById(plainId)
}

func (s *CashFlowService) DeleteById(plainId string) (model.CashFlowEntity, error) {
	// Validate ID
	if err := validation.ValidateID(plainId); err != nil {
		return model.CashFlowEntity{}, err
	}

	existCashFlowEntity := s.cashFlowMapper.GetCashFlowByObjectId(plainId)
	if existCashFlowEntity.IsEmpty() {
		return model.CashFlowEntity{}, errors.New("cash_flow not found")
	}

	existCashFlowEntity = s.cashFlowMapper.DeleteCashFlowByObjectId(plainId)
	if existCashFlowEntity.IsEmpty() {
		return model.CashFlowEntity{}, errors.New("cash_flow delete failed")
	}
	return existCashFlowEntity, nil
}

func DeleteByDate(belongsDate string) ([]model.CashFlowEntity, error) {
	return defaultCashFlowService().DeleteByDate(belongsDate)
}

func (s *CashFlowService) DeleteByDate(belongsDate string) ([]model.CashFlowEntity, error) {
	// Validate date
	if err := validation.ValidateDate(belongsDate); err != nil {
		return []model.CashFlowEntity{}, err
	}

	deleteDate := util.FormatDateFromStringWithoutDash(belongsDate)
	if reflect.DeepEqual(deleteDate, time.Time{}) {
		return []model.CashFlowEntity{}, errors.New("belongs_date error, try format like 19700101")
	}

	cashFlowList := s.cashFlowMapper.DeleteCashFlowByBelongsDate(deleteDate)
	return cashFlowList, nil
}

// DeleteByIdForUser deletes a cash flow by ID, ensuring it belongs to the user
func DeleteByIdForUser(plainId string, userId string) (model.CashFlowEntity, error) {
	return defaultCashFlowService().DeleteByIdForUser(plainId, userId)
}

func (s *CashFlowService) DeleteByIdForUser(plainId string, userId string) (model.CashFlowEntity, error) {
	// Validate ID
	if err := validation.ValidateID(plainId); err != nil {
		return model.CashFlowEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return model.CashFlowEntity{}, errors.New("invalid user ID")
	}

	// Check if it exists and belongs to user
	existCashFlowEntity := s.cashFlowMapper.GetCashFlowByObjectIdAndUser(plainId, userObjectId)
	if existCashFlowEntity.IsEmpty() {
		return model.CashFlowEntity{}, errors.New("cash_flow not found or access denied")
	}

	// Delete it
	deletedEntity := s.cashFlowMapper.DeleteCashFlowByObjectIdAndUser(plainId, userObjectId)
	if deletedEntity.IsEmpty() {
		return model.CashFlowEntity{}, errors.New("cash_flow delete failed")
	}

	// Populate category info for return value
	category := s.categoryMapper.GetCategoryByObjectId(deletedEntity.CategoryId.Hex())
	if !category.IsEmpty() {
		deletedEntity.CategoryName = category.Name
		deletedEntity.CategoryType = category.Type
	} else {
		deletedEntity.CategoryName = "Unknown"
	}

	return deletedEntity, nil
}

// DeleteByDateForUser deletes cash flows for a specific date for the user
func DeleteByDateForUser(belongsDate string, userId string) ([]model.CashFlowEntity, error) {
	return defaultCashFlowService().DeleteByDateForUser(belongsDate, userId)
}

func (s *CashFlowService) DeleteByDateForUser(belongsDate string, userId string) ([]model.CashFlowEntity, error) {
	// Validate date
	if err := validation.ValidateDate(belongsDate); err != nil {
		return []model.CashFlowEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return []model.CashFlowEntity{}, errors.New("invalid user ID")
	}

	// Parse date
	deleteDate := util.FormatDateFromStringWithoutDash(belongsDate)
	if deleteDate.IsZero() {
		return []model.CashFlowEntity{}, errors.New("belongs_date error, try format like 19700101")
	}

	cashFlowList := s.cashFlowMapper.DeleteCashFlowsByBelongsDateAndUser(deleteDate, userObjectId)

	// Populate category info for each deleted item
	for i := range cashFlowList {
		entity := &cashFlowList[i]
		category := s.categoryMapper.GetCategoryByObjectId(entity.CategoryId.Hex())
		if !category.IsEmpty() {
			entity.CategoryName = category.Name
			entity.CategoryType = category.Type
		} else {
			entity.CategoryName = "Unknown"
		}
	}

	return cashFlowList, nil
}
