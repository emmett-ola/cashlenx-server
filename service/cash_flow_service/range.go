package cash_flow_service

import (
	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
)

// QueryByDateRange queries cash flows within a date range
func QueryByDateRange(fromDate, toDate string) ([]*model.CashFlowEntity, error) {
	// Validate date range
	if err := validation.ValidateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}

	// Parse dates
	from := util.FormatDateFromStringWithoutDash(fromDate)
	to := util.FormatDateFromStringWithoutDash(toDate)

	// Single query for entire date range
	results := cash_flow_mapper.INSTANCE.GetCashFlowsByDateRange(from, to)

	// Populate category info
	for i := range results {
		entity := &results[i]
		category := category_mapper.INSTANCE.GetCategoryByObjectId(entity.CategoryId.Hex())
		if !category.IsEmpty() {
			entity.CategoryName = category.Name
			entity.CategoryType = category.Type
		} else {
			entity.CategoryName = "Unknown"
		}
	}

	// Convert to pointer slice
	var resultPtrs []*model.CashFlowEntity
	for i := range results {
		resultPtrs = append(resultPtrs, &results[i])
	}

	return resultPtrs, nil
}
