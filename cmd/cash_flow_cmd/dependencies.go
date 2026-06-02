package cash_flow_cmd

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
)

var (
	saveExpenseForUser      = cash_flow_service.SaveExpense
	saveIncomeForUser       = cash_flow_service.SaveIncome
	updateCashForUser       = cash_flow_service.UpdateByIdForUser
	queryCashForUser        = cash_flow_service.QueryAllForUser
	queryCashByIDForUser    = cash_flow_service.QueryByIdForUser
	queryCashByDateForUser  = cash_flow_service.QueryByDateForUser
	queryCashRangeForUser   = cash_flow_service.QueryByDateRangeForUser
	deleteCashByIDForUser   = cash_flow_service.DeleteByIdForUser
	deleteCashByDateForUser = cash_flow_service.DeleteByDateForUser
	getCashSummaryForUser   = cash_flow_service.GetSummaryForUser
)

type queryCashForUserFunc func(
	userId string,
	cashType string,
	categoryId string,
	description string,
	exactDescription string,
	fromDate string,
	toDate string,
	limit int,
	offset int,
) ([]*model.CashFlowEntity, int64, error)
