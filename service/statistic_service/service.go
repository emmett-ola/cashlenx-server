package statistic_service

import (
	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
)

type StatisticService struct {
	cashFlowMapper cash_flow_mapper.CashFlowMapper
	categoryMapper category_mapper.CategoryMapper
}

func NewStatisticService(cashFlowMapper cash_flow_mapper.CashFlowMapper, categoryMapper category_mapper.CategoryMapper) *StatisticService {
	return &StatisticService{
		cashFlowMapper: cashFlowMapper,
		categoryMapper: categoryMapper,
	}
}

func defaultStatisticService() *StatisticService {
	return NewStatisticService(cash_flow_mapper.INSTANCE, category_mapper.INSTANCE)
}
