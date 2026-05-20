package category_service

import (
	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
)

type CategoryService struct {
	categoryMapper category_mapper.CategoryMapper
	cashFlowMapper cash_flow_mapper.CashFlowMapper
}

func NewCategoryService(categoryMapper category_mapper.CategoryMapper, cashFlowMapper cash_flow_mapper.CashFlowMapper) *CategoryService {
	return &CategoryService{
		categoryMapper: categoryMapper,
		cashFlowMapper: cashFlowMapper,
	}
}

func defaultCategoryService() *CategoryService {
	return NewCategoryService(category_mapper.INSTANCE, cash_flow_mapper.INSTANCE)
}
