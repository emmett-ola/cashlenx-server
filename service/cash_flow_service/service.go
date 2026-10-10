package cash_flow_service

import (
	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/model"
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

func (s *CashFlowService) enrichCategoryInfo(entity *model.CashFlowEntity) {
	entity.CategoryName = "Unknown"
	entity.CategoryType = ""
	entity.CategoryEmoji = model.DefaultCategoryEmoji
	entity.CategoryBgColor = model.DefaultCategoryBackgroundColor

	category := s.categoryMapper.GetCategoryByObjectId(entity.CategoryId.Hex())
	if category.IsEmpty() {
		return
	}

	entity.CategoryName = category.Name
	entity.CategoryType = category.Type
	if category.Emoji != "" {
		entity.CategoryEmoji = category.Emoji
	}
	if category.BgColor != "" {
		entity.CategoryBgColor = category.BgColor
	}
}
