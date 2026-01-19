package manage_service

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
)

// BackupData represents the structure of backup data
type BackupData struct {
	Version    string                   `json:"version"`
	Timestamp  string                   `json:"timestamp"`
	Users      []map[string]interface{} `json:"users"`
	Categories []map[string]interface{} `json:"categories"`
	CashFlows  []map[string]interface{} `json:"cash_flows"`
}

// EntityStats represents statistics for a single entity type (cash_flows, categories, or users)
type EntityStats struct {
	Success    int      `json:"success"`
	Failed     int      `json:"failed"`
	FailedList []string `json:"failed_list,omitempty"`
}

// OperationStats represents comprehensive statistics for an operation (backup/restore/truncate)
type OperationStats struct {
	Users      EntityStats `json:"users"`
	Categories EntityStats `json:"categories"`
	CashFlows  EntityStats `json:"cash_flows"`
}

// CreateBackup creates a backup of all database data
func CreateBackup(filePath string) (OperationStats, error) {
	stats := OperationStats{
		Users:      EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		Categories: EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		CashFlows:  EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
	}

	if filePath == "" {
		return stats, errors.New("file path cannot be empty")
	}

	// Get all users (no pagination limit - get everything)
	users := user_mapper.INSTANCE.GetAllUsersIncludeDeleted(0, 0)
	stats.Users.Success = len(users)

	// Convert users to map format for JSON serialization
	userMaps := make([]map[string]interface{}, len(users))
	for i, user := range users {
		userMap := map[string]interface{}{
			"Id":           user.Id.Hex(),
			"Username":     user.Username,
			"PasswordHash": user.PasswordHash,
			"CreateTime":   user.CreateTime,
			"UpdateTime":   user.UpdateTime,
			"IsActive":     user.IsActive,
			"Role":         user.Role,
			"CreateUserId": user.CreateUserId.Hex(),
			"UpdateUserId": user.UpdateUserId.Hex(),
			"IsDelete":     user.IsDelete,
		}
		if user.DeleteUserId != nil {
			userMap["DeleteUserId"] = user.DeleteUserId.Hex()
		}
		if user.DeleteTime != nil {
			userMap["DeleteTime"] = user.DeleteTime
		}
		userMaps[i] = userMap
	}

	// Get all categories (no pagination limit - get everything)
	categories := category_mapper.INSTANCE.GetAllCategoriesIncludeDeleted(0, 0)
	stats.Categories.Success = len(categories)

	// Convert categories to map format for JSON serialization
	categoryMaps := make([]map[string]interface{}, len(categories))
	for i, cat := range categories {
		catMap := map[string]interface{}{
			"Id":           cat.Id.Hex(),
			"UserId":       cat.UserId.Hex(),
			"Name":         cat.Name,
			"Type":         cat.Type,
			"ParentId":     cat.ParentId.Hex(),
			"Remark":       cat.Remark,
			"CreateTime":   cat.CreateTime,
			"ModifyTime":   cat.UpdateTime,
			"CreateUserId": cat.CreateUserId.Hex(),
			"UpdateUserId": cat.UpdateUserId.Hex(),
			"IsDelete":     cat.IsDelete,
		}
		if cat.DeleteUserId != nil {
			catMap["DeleteUserId"] = cat.DeleteUserId.Hex()
		}
		if cat.DeleteTime != nil {
			catMap["DeleteTime"] = cat.DeleteTime
		}
		categoryMaps[i] = catMap
	}

	// Get all cash flows (no pagination limit - get everything)
	cashFlows := cash_flow_mapper.INSTANCE.GetAllCashFlowsIncludeDeleted(0, 0)
	stats.CashFlows.Success = len(cashFlows)

	// Convert cash flows to map format for JSON serialization
	cashFlowMaps := make([]map[string]interface{}, len(cashFlows))
	for i, cf := range cashFlows {
		cfMap := map[string]interface{}{
			"Id":           cf.Id.Hex(),
			"UserId":       cf.UserId.Hex(),
			"CategoryId":   cf.CategoryId.Hex(),
			"BelongsDate":  cf.BelongsDate,
			// "FlowType":    cf.FlowType,
			"Amount":       cf.Amount,
			"Description":  cf.Description,
			"Remark":       cf.Remark,
			"CreateTime":   cf.CreateTime,
			"UpdateTime":   cf.UpdateTime,
			"CreateUserId": cf.CreateUserId.Hex(),
			"UpdateUserId": cf.UpdateUserId.Hex(),
			"IsDelete":     cf.IsDelete,
		}
		if cf.DeleteUserId != nil {
			cfMap["DeleteUserId"] = cf.DeleteUserId.Hex()
		}
		if cf.DeleteTime != nil {
			cfMap["DeleteTime"] = cf.DeleteTime
		}
		cashFlowMaps[i] = cfMap
	}

	// Create backup structure
	backup := BackupData{
		Version:    "2.0.0", // Version updated for user data isolation
		Timestamp:  time.Now().Format(time.RFC3339),
		Users:      userMaps,
		Categories: categoryMaps,
		CashFlows:  cashFlowMaps,
	}

	// Write to file
	file, err := os.Create(filePath)
	if err != nil {
		return stats, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(backup); err != nil {
		return stats, err
	}

	return stats, nil
}
