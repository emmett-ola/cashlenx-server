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
			"id":             user.Id.Hex(),
			"username":       user.Username,
			"password_hash":  user.PasswordHash,
			"create_time":    user.CreateTime,
			"update_time":    user.UpdateTime,
			"is_active":      user.IsActive,
			"role":           user.Role,
			"create_user_id": user.CreateUserId.Hex(),
			"update_user_id": user.UpdateUserId.Hex(),
			"is_delete":      user.IsDelete,
		}
		if user.DeleteUserId != nil {
			userMap["delete_user_id"] = user.DeleteUserId.Hex()
		}
		if user.DeleteTime != nil {
			userMap["delete_time"] = user.DeleteTime
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
			"id":              cat.Id.Hex(),
			"belongs_user_id": cat.BelongsUserId.Hex(),
			"name":            cat.Name,
			"type":            cat.Type,
			"parent_id":      cat.ParentId.Hex(),
			"remark":         cat.Remark,
			"create_time":    cat.CreateTime,
			"update_time":    cat.UpdateTime,
			"create_user_id": cat.CreateUserId.Hex(),
			"update_user_id": cat.UpdateUserId.Hex(),
			"is_delete":      cat.IsDelete,
		}
		if cat.DeleteUserId != nil {
			catMap["delete_user_id"] = cat.DeleteUserId.Hex()
		}
		if cat.DeleteTime != nil {
			catMap["delete_time"] = cat.DeleteTime
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
			"id":              cf.Id.Hex(),
			"belongs_user_id": cf.BelongsUserId.Hex(),
			"category_id":     cf.CategoryId.Hex(),
			"belongs_date":    cf.BelongsDate,
			// "flow_type":      cf.FlowType,
			"amount":         cf.Amount,
			"description":    cf.Description,
			"remark":         cf.Remark,
			"create_time":    cf.CreateTime,
			"update_time":    cf.UpdateTime,
			"create_user_id": cf.CreateUserId.Hex(),
			"update_user_id": cf.UpdateUserId.Hex(),
			"is_delete":      cf.IsDelete,
		}
		if cf.DeleteUserId != nil {
			cfMap["delete_user_id"] = cf.DeleteUserId.Hex()
		}
		if cf.DeleteTime != nil {
			cfMap["delete_time"] = cf.DeleteTime
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
