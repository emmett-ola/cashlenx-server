package cash_flow_service

import (
	"errors"
	"time"

	myErrors "github.com/macar-x/cashlenx-server/errors"
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
	return defaultCashFlowService().QueryById(plainId)
}

func (s *CashFlowService) QueryById(plainId string) (model.CashFlowEntity, error) {
	if err := validation.ValidateID(plainId); err != nil {
		return model.CashFlowEntity{}, err
	}

	cashFlowEntity := s.cashFlowMapper.GetCashFlowByObjectId(plainId)
	if cashFlowEntity.IsEmpty() {
		return model.CashFlowEntity{}, myErrors.NewNotFoundError("cash_flow not found")
	}

	// Populate category info
	category := s.categoryMapper.GetCategoryByObjectId(cashFlowEntity.CategoryId.Hex())
	if !category.IsEmpty() {
		cashFlowEntity.CategoryName = category.Name
		cashFlowEntity.CategoryType = category.Type
	} else {
		cashFlowEntity.CategoryName = "Unknown"
	}

	return cashFlowEntity, nil
}

func QueryByDate(belongsDate string) ([]model.CashFlowEntity, error) {
	return defaultCashFlowService().QueryByDate(belongsDate)
}

func (s *CashFlowService) QueryByDate(belongsDate string) ([]model.CashFlowEntity, error) {
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

	// Use unified filter
	filter := model.CashFlowFilter{
		FromDate: startOfDay,
		ToDate:   endOfDay,
	}

	matchedCashFlowList, err := s.cashFlowMapper.GetCashFlowsByFilter(filter)
	if err != nil {
		return nil, err
	}

	// Populate category info
	s.populateCategoryInfo(matchedCashFlowList)

	return matchedCashFlowList, nil
}

func QueryByExactDescription(exactDescription string) ([]model.CashFlowEntity, error) {
	return defaultCashFlowService().QueryByExactDescription(exactDescription)
}

func (s *CashFlowService) QueryByExactDescription(exactDescription string) ([]model.CashFlowEntity, error) {
	// Use unified filter
	filter := model.CashFlowFilter{
		ExactDescription: exactDescription,
	}

	matchedCashFlowList, err := s.cashFlowMapper.GetCashFlowsByFilter(filter)
	if err != nil {
		return nil, err
	}

	// Populate category info
	s.populateCategoryInfo(matchedCashFlowList)

	return matchedCashFlowList, nil
}

func QueryByFuzzyDescription(fuzzyDescription string) ([]model.CashFlowEntity, error) {
	return defaultCashFlowService().QueryByFuzzyDescription(fuzzyDescription)
}

func (s *CashFlowService) QueryByFuzzyDescription(fuzzyDescription string) ([]model.CashFlowEntity, error) {
	// Use unified filter
	filter := model.CashFlowFilter{
		Description: fuzzyDescription,
	}

	matchedCashFlowList, err := s.cashFlowMapper.GetCashFlowsByFilter(filter)
	if err != nil {
		return nil, err
	}

	// Populate category info
	s.populateCategoryInfo(matchedCashFlowList)

	return matchedCashFlowList, nil
}

// User-specific operations

// QueryByIdForUser retrieves a cash flow by ID, ensuring it belongs to the user
func QueryByIdForUser(plainId string, userId string) (model.CashFlowEntity, error) {
	return defaultCashFlowService().QueryByIdForUser(plainId, userId)
}

func (s *CashFlowService) QueryByIdForUser(plainId string, userId string) (model.CashFlowEntity, error) {
	// Validate ID
	if err := validation.ValidateID(plainId); err != nil {
		return model.CashFlowEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return model.CashFlowEntity{}, errors.New("invalid user ID")
	}

	cashFlowEntity := s.cashFlowMapper.GetCashFlowByObjectIdAndUser(plainId, userObjectId)
	if cashFlowEntity.IsEmpty() {
		return model.CashFlowEntity{}, errors.New("cash_flow not found or access denied")
	}

	// Populate category info
	category := s.categoryMapper.GetCategoryByObjectId(cashFlowEntity.CategoryId.Hex())
	if !category.IsEmpty() {
		cashFlowEntity.CategoryName = category.Name
		cashFlowEntity.CategoryType = category.Type
	} else {
		cashFlowEntity.CategoryName = "Unknown"
	}

	return cashFlowEntity, nil
}

// QueryByDateForUser retrieves cash flows for a specific date for the user
func QueryByDateForUser(belongsDate string, userId string) ([]model.CashFlowEntity, error) {
	return defaultCashFlowService().QueryByDateForUser(belongsDate, userId)
}

func (s *CashFlowService) QueryByDateForUser(belongsDate string, userId string) ([]model.CashFlowEntity, error) {
	// Validate date
	if err := validation.ValidateDate(belongsDate); err != nil {
		return []model.CashFlowEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
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

	// Use unified filter
	filter := model.CashFlowFilter{
		UserId:   userObjectId,
		FromDate: startOfDay,
		ToDate:   endOfDay,
	}

	matchedCashFlowList, err := s.cashFlowMapper.GetCashFlowsByFilter(filter)
	if err != nil {
		return nil, err
	}

	// Populate category info
	s.populateCategoryInfo(matchedCashFlowList)

	return matchedCashFlowList, nil
}

// QueryByDateRangeForUser retrieves cash flows within a date range for the user
func QueryByDateRangeForUser(fromDateStr, toDateStr string, userId string) ([]model.CashFlowEntity, error) {
	return defaultCashFlowService().QueryByDateRangeForUser(fromDateStr, toDateStr, userId)
}

func (s *CashFlowService) QueryByDateRangeForUser(fromDateStr, toDateStr string, userId string) ([]model.CashFlowEntity, error) {
	fromDate, toDate, err := validateDateRange(fromDateStr, toDateStr)
	if err != nil {
		return []model.CashFlowEntity{}, err
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		util.Logger.Warnf("Invalid user ID format: '%s'", userId)
		return []model.CashFlowEntity{}, errors.New("invalid user ID")
	}

	// Use UTC time
	startDate := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 23, 59, 59, 999999999, time.UTC)

	// Use unified filter
	filter := model.CashFlowFilter{
		UserId:   userObjectId,
		FromDate: startDate,
		ToDate:   endDate,
	}

	matchedCashFlowList, err := s.cashFlowMapper.GetCashFlowsByFilter(filter)
	if err != nil {
		return nil, err
	}

	// Populate category info
	s.populateCategoryInfo(matchedCashFlowList)

	return matchedCashFlowList, nil
}

func validateDateRange(fromDateStr, toDateStr string) (time.Time, time.Time, error) {
	if fromDateStr == "" || toDateStr == "" {
		return time.Time{}, time.Time{}, errors.New("from_date and to_date are required")
	}

	fromDate, err := util.ParseDate(fromDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("from_date error, try format like 19700101, 1970-01-01, or 1970/01/01")
	}

	toDate, err := util.ParseDate(toDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("to_date error, try format like 19700101, 1970-01-01, or 1970/01/01")
	}

	if fromDate.After(toDate) {
		return time.Time{}, time.Time{}, errors.New("from_date should be before or equal to to_date")
	}

	return fromDate, toDate, nil
}

// Helper function to populate category info
func populateCategoryInfo(cashFlowList []model.CashFlowEntity) {
	defaultCashFlowService().populateCategoryInfo(cashFlowList)
}

func (s *CashFlowService) populateCategoryInfo(cashFlowList []model.CashFlowEntity) {
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
}
