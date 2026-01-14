package manage_service

import (
	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
)

// ResetDatabase clears all data from the database
func ResetDatabase() (OperationStats, error) {
	stats := OperationStats{
		CashFlows:  EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		Categories: EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
	}

	// Count items before truncation to provide accurate statistics
	stats.CashFlows.Success = int(cash_flow_mapper.INSTANCE.CountAllCashFlows())
	stats.Categories.Success = int(category_mapper.INSTANCE.CountAllCategories())

	// This is a dangerous operation - truncate all data
	// First truncate cash flows (dependent data)
	if err := cash_flow_mapper.INSTANCE.TruncateCashFlows(); err != nil {
		// If truncation fails, set success to 0 and failed to the counted items
		stats.CashFlows.Failed = stats.CashFlows.Success
		stats.CashFlows.Success = 0
		return stats, err
	}

	// Then truncate categories (parent data)
	if err := category_mapper.INSTANCE.TruncateCategories(); err != nil {
		// If truncation fails, set success to 0 and failed to the counted items
		stats.Categories.Failed = stats.Categories.Success
		stats.Categories.Success = 0
		return stats, err
	}

	// Then truncate users (root data)
	if err := user_mapper.INSTANCE.TruncateUsers(); err != nil {
		return stats, err
	}

	return stats, nil
}

// TruncateDatabase is an alias for ResetDatabase - clears all data from the database
func TruncateDatabase() (OperationStats, error) {
	return ResetDatabase()
}
