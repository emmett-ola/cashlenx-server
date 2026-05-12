package cash_flow_service

import (
	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
)

type CashFlowService struct {
	cashFlowMapper cash_flow_mapper.CashFlowMapper
	categoryMapper category_mapper.CategoryMapper
}

func NewCashFlowService(cashFlowMapper cash_flow_mapper.CashFlowMapper, categoryMapper category_mapper.CategoryMapper) *CashFlowService {
	return &CashFlowService{
		cashFlowMapper: cashFlowMapper,
		categoryMapper: categoryMapper,
	}
}

func defaultCashFlowService() *CashFlowService {
	return NewCashFlowService(cash_flow_mapper.INSTANCE, category_mapper.INSTANCE)
}
