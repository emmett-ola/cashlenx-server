package cash_flow_service

import (
	"errors"
	"time"

	myErrors "github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func IsQueryFieldsConflicted(plainId, belongsDate, exactDescription, fuzzyDescription string) bool {
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

	// exact_description is not empty
	if exactDescription != "" {
		if semiOptionalFieldFilledFlag {
			return true
		}
		semiOptionalFieldFilledFlag = true
	}

	// fuzzy_description is not empty
	if fuzzyDescription != "" {
		if semiOptionalFieldFilledFlag {
			return true
		}
		semiOptionalFieldFilledFlag = true
	}

	// should have one and only one field filled
	return !semiOptionalFieldFilledFlag
}

func QueryById(plainId string) (model.CashFlowEntity, error) {
	cashFlowEntity := cash_flow_mapper.INSTANCE.GetCashFlowByObjectId(plainId)
	if cashFlowEntity.IsEmpty() {
		return model.CashFlowEntity{}, myErrors.NewNotFoundError("cash_flow not found")
	}
	return cashFlowEntity, nil
}

func QueryByDate(belongsDate string) ([]model.CashFlowEntity, error) {
	// Validate date
	if err := validation.ValidateDate(belongsDate); err != nil {
		return []model.CashFlowEntity{}, err
	}

	// Parse the date string using our multi-format parser
	parsedDate, err := util.ParseDate(belongsDate)
	if err != nil {
		return []model.CashFlowEntity{}, myErrors.NewInvalidInputError("belongs_date error, try format like 19700101, 1970-01-01, or 1970/01/01")
	}

	// Use UTC time for consistent querying (MongoDB stores dates in UTC)
	// Set start to beginning of the day in UTC
	startOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
	// Set end to end of the day in UTC
	endOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 999999999, time.UTC)

	matchedCashFlowList := cash_flow_mapper.INSTANCE.GetCashFlowsByDateRange(startOfDay, endOfDay)
	return matchedCashFlowList, nil
}

func QueryByExactDescription(exactDescription string) ([]model.CashFlowEntity, error) {
	matchedCashFlowList := cash_flow_mapper.INSTANCE.GetCashFlowsByExactDesc(exactDescription)
	return matchedCashFlowList, nil
}

func QueryByFuzzyDescription(fuzzyDescription string) ([]model.CashFlowEntity, error) {
	matchedCashFlowList := cash_flow_mapper.INSTANCE.GetCashFlowsByFuzzyDesc(fuzzyDescription)
	return matchedCashFlowList, nil
}

// User-specific operations

// QueryByIdForUser retrieves a cash flow by ID, ensuring it belongs to the user
func QueryByIdForUser(plainId string, userId string) (model.CashFlowEntity, error) {
	// Validate ID
	if err := validation.ValidateID(plainId); err != nil {
		return model.CashFlowEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return model.CashFlowEntity{}, errors.New("invalid user ID")
	}

	cashFlowEntity := cash_flow_mapper.INSTANCE.GetCashFlowByObjectIdAndUser(plainId, userObjectId)
	if cashFlowEntity.IsEmpty() {
		return model.CashFlowEntity{}, errors.New("cash_flow not found or access denied")
	}
	return cashFlowEntity, nil
}

// QueryByDateForUser retrieves cash flows for a specific date for the user
func QueryByDateForUser(belongsDate string, userId string) ([]model.CashFlowEntity, error) {
	// Validate date
	if err := validation.ValidateDate(belongsDate); err != nil {
		return []model.CashFlowEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		// Try to use the user ID string directly if it's not a valid ObjectID
		// This is for cases where userId might not be a MongoDB ObjectID (e.g. from OIDC or other sources)
		// But since we are using MongoDB, we should ensure it's a valid ObjectID
		// For now, let's log a warning and return error
		util.Logger.Warnf("Invalid user ID format: %s", userId)
		return []model.CashFlowEntity{}, errors.New("invalid user ID")
	}

	// Parse the date string
	parsedDate, err := util.ParseDate(belongsDate)
	if err != nil {
		return []model.CashFlowEntity{}, errors.New("belongs_date error, try format like 19700101, 1970-01-01, or 1970/01/01")
	}

	// Use UTC time for consistent querying
	startOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 999999999, time.UTC)

	matchedCashFlowList := cash_flow_mapper.INSTANCE.GetCashFlowsByDateRangeAndUser(startOfDay, endOfDay, userObjectId)
	return matchedCashFlowList, nil
}

// QueryByDateRangeForUser retrieves cash flows within a date range for the user
func QueryByDateRangeForUser(fromDateStr, toDateStr string, userId string) ([]model.CashFlowEntity, error) {
	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		util.Logger.Warnf("Invalid user ID format: '%s'", userId)
		return []model.CashFlowEntity{}, errors.New("invalid user ID")
	}

	// Parse date strings
	fromDate, err := util.ParseDate(fromDateStr)
	if err != nil {
		return []model.CashFlowEntity{}, errors.New("from_date error, try format like 19700101, 1970-01-01, or 1970/01/01")
	}

	toDate, err := util.ParseDate(toDateStr)
	if err != nil {
		return []model.CashFlowEntity{}, errors.New("to_date error, try format like 19700101, 1970-01-01, or 1970/01/01")
	}

	// Use UTC time
	startDate := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 23, 59, 59, 999999999, time.UTC)

	matchedCashFlowList := cash_flow_mapper.INSTANCE.GetCashFlowsByDateRangeAndUser(startDate, endDate, userObjectId)
	return matchedCashFlowList, nil
}
