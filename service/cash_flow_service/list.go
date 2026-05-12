package cash_flow_service

import (
	"strings"
	"time"

	myErrors "github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QueryAll queries all cash flows with optional filtering and pagination
func QueryAll(
	cashType string,
	categoryId string,
	description string,
	exactDescription string,
	fromDateStr string,
	toDateStr string,
	limit int,
	offset int,
) ([]*model.CashFlowEntity, int64, error) {
	return defaultCashFlowService().QueryAll(cashType, categoryId, description, exactDescription, fromDateStr, toDateStr, limit, offset)
}

func (s *CashFlowService) QueryAll(
	cashType string,
	categoryId string,
	description string,
	exactDescription string,
	fromDateStr string,
	toDateStr string,
	limit int,
	offset int,
) ([]*model.CashFlowEntity, int64, error) {
	totalCount := s.cashFlowMapper.CountAllCashFlows()
	cashFlows := s.cashFlowMapper.GetAllCashFlows(limit, offset)

	var fromDate, toDate time.Time
	var err error

	if cashType != "" {
		cashType = strings.ToLower(cashType)
	}

	if fromDateStr != "" {
		fromDate, err = util.ParseDate(fromDateStr)
		if err != nil {
			return nil, 0, err
		}
	}

	if toDateStr != "" {
		toDate, err = util.ParseDate(toDateStr)
		if err != nil {
			return nil, 0, err
		}
	}

	var filteredResults []*model.CashFlowEntity
	for i := range cashFlows {
		entity := cashFlows[i]

		category := s.categoryMapper.GetCategoryByObjectId(entity.CategoryId.Hex())
		if !category.IsEmpty() {
			entity.CategoryName = category.Name
			entity.CategoryType = category.Type
		} else {
			entity.CategoryName = "Unknown"
		}

		match := true

		if cashType != "" && !strings.EqualFold(entity.CategoryType, cashType) {
			match = false
		}

		if categoryId != "" && entity.CategoryId.Hex() != categoryId {
			match = false
		}

		if exactDescription != "" && entity.Description != exactDescription {
			match = false
		}

		if description != "" && exactDescription == "" {
			if !strings.Contains(entity.Description, description) {
				match = false
			}
		}

		if !fromDate.IsZero() && entity.BelongsDate.Before(fromDate) {
			match = false
		}
		if !toDate.IsZero() && entity.BelongsDate.After(toDate) {
			match = false
		}

		if match {
			filteredResults = append(filteredResults, &entity)
		}
	}

	return filteredResults, totalCount, nil
}

// QueryAllForUser queries all cash flows for a user with optional filtering and pagination
func QueryAllForUser(
	userId string,
	cashType string,
	categoryId string,
	description string,
	exactDescription string,
	fromDateStr string,
	toDateStr string,
	limit int,
	offset int,
) ([]*model.CashFlowEntity, int64, error) {
	return defaultCashFlowService().QueryAllForUser(userId, cashType, categoryId, description, exactDescription, fromDateStr, toDateStr, limit, offset)
}

func (s *CashFlowService) QueryAllForUser(
	userId string,
	cashType string,
	categoryId string,
	description string,
	exactDescription string,
	fromDateStr string,
	toDateStr string,
	limit int,
	offset int,
) ([]*model.CashFlowEntity, int64, error) {
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return nil, 0, myErrors.NewInvalidInputError("invalid user ID")
	}

	filter := model.CashFlowFilter{
		UserId:           userObjectId,
		CashType:         cashType,
		CategoryId:       categoryId,
		Description:      description,
		ExactDescription: exactDescription,
		Limit:            limit,
		Offset:           offset,
	}

	if fromDateStr != "" {
		fromDate, err := util.ParseDate(fromDateStr)
		if err != nil {
			return nil, 0, err
		}
		filter.FromDate = fromDate
	}

	if toDateStr != "" {
		toDate, err := util.ParseDate(toDateStr)
		if err != nil {
			return nil, 0, err
		}
		filter.ToDate = toDate
	}

	cashFlows, err := s.cashFlowMapper.GetCashFlowsByFilter(filter)
	if err != nil {
		return nil, 0, err
	}

	if cashType != "" {
		unpagedFilter := filter
		unpagedFilter.Limit = 0
		unpagedFilter.Offset = 0

		cashFlows, err = s.cashFlowMapper.GetCashFlowsByFilter(unpagedFilter)
		if err != nil {
			return nil, 0, err
		}
	}

	results := s.enrichAndFilterByType(cashFlows, cashType)

	if cashType != "" {
		totalCount := int64(len(results))
		return pageCashFlowPointers(results, limit, offset), totalCount, nil
	}

	totalCount, err := s.cashFlowMapper.CountCashFlowsByFilter(filter)
	if err != nil {
		return nil, 0, err
	}
	return results, totalCount, nil
}

func (s *CashFlowService) enrichAndFilterByType(cashFlows []model.CashFlowEntity, cashType string) []*model.CashFlowEntity {
	var results []*model.CashFlowEntity
	for i := range cashFlows {
		entity := cashFlows[i]

		category := s.categoryMapper.GetCategoryByObjectId(entity.CategoryId.Hex())
		if !category.IsEmpty() {
			entity.CategoryName = category.Name
			entity.CategoryType = category.Type
		} else {
			entity.CategoryName = "Unknown"
		}

		if cashType != "" && !strings.EqualFold(entity.CategoryType, cashType) {
			continue
		}

		results = append(results, &entity)
	}
	return results
}

func pageCashFlowPointers(flows []*model.CashFlowEntity, limit, offset int) []*model.CashFlowEntity {
	if offset < 0 {
		offset = 0
	}
	if offset > len(flows) {
		return []*model.CashFlowEntity{}
	}
	if limit <= 0 {
		return flows[offset:]
	}
	end := offset + limit
	if end > len(flows) {
		end = len(flows)
	}
	return flows[offset:end]
}
