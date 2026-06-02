package statistic_controller

import "github.com/macar-x/cashlenx-server/service/statistic_service"

var (
	exportStatisticXLSXForUser = statistic_service.ExportForUser
	exportStatisticCSVForUser  = statistic_service.ExportToCSVForUser
	exportStatisticPDFForUser  = statistic_service.ExportToPDFForUser
	importStatisticForUser     = statistic_service.ImportForUser
	getTopExpensesForUser      = statistic_service.GetTopExpensesForUser
)
