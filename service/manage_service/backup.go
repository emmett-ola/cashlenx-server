package manage_service

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/mapper/user_config_mapper"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
)

// BackupData represents the structure of backup data
type BackupData struct {
	Version     string                   `json:"version"`
	Timestamp   string                   `json:"timestamp"`
	Users       []map[string]interface{} `json:"users"`
	UserConfigs []map[string]interface{} `json:"user_configurations,omitempty"`
	Categories  []map[string]interface{} `json:"categories"`
	CashFlows   []map[string]interface{} `json:"cash_flows"`
}

// EntityStats represents statistics for a single entity type (cash_flows, categories, or users)
type EntityStats struct {
	Success    int      `json:"success"`
	Failed     int      `json:"failed"`
	FailedList []string `json:"failed_list,omitempty"`
}

// OperationStats represents comprehensive statistics for an operation (backup/restore/truncate)
type OperationStats struct {
	Users       EntityStats `json:"users"`
	UserConfigs EntityStats `json:"user_configurations"`
	Categories  EntityStats `json:"categories"`
	CashFlows   EntityStats `json:"cash_flows"`
}

// AdminDumpDatabase creates a full database dump (including deleted records)
func AdminDumpDatabase(filePath string) (OperationStats, error) {
	return AdminDumpDatabaseWithProgress(filePath, nil)
}

func AdminDumpDatabaseWithProgress(filePath string, progress ProgressFunc) (OperationStats, error) {
	stats := OperationStats{
		Users:       EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		UserConfigs: EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		Categories:  EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		CashFlows:   EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
	}

	if filePath == "" {
		return stats, errors.New("file path cannot be empty")
	}
	if err := PreflightBackupDestination(filePath); err != nil {
		return stats, err
	}
	emit(progress, "preflight", "backup", 1, 1, "Backup destination validated")

	// Get all users (no pagination limit - get everything)
	users := user_mapper.INSTANCE.GetAllUsersIncludeDeleted(0, 0)
	stats.Users.Success = len(users)
	emit(progress, "read", "users", len(users), len(users), "Users collected")

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
		if user.Nickname != "" {
			userMap["nickname"] = user.Nickname
		} else {
			userMap["nickname"] = nil
		}
		if user.AvatarUrl != "" {
			userMap["avatar_url"] = user.AvatarUrl
		} else {
			userMap["avatar_url"] = nil
		}
		if user.EmailAddress != "" {
			userMap["email_address"] = user.EmailAddress
		} else {
			userMap["email_address"] = nil
		}
		if user.Gender != "" {
			userMap["gender"] = user.Gender
		} else {
			userMap["gender"] = nil
		}
		if user.DeleteUserId != nil {
			userMap["delete_user_id"] = user.DeleteUserId.Hex()
		} else {
			userMap["delete_user_id"] = nil
		}
		if user.DeleteTime != nil {
			userMap["delete_time"] = user.DeleteTime
		} else {
			userMap["delete_time"] = nil
		}
		userMaps[i] = userMap
	}

	configs := user_config_mapper.INSTANCE.GetAllIncludeDeleted(0, 0)
	stats.UserConfigs.Success = len(configs)
	emit(progress, "read", "user_configurations", len(configs), len(configs), "User configurations collected")
	configMaps := make([]map[string]interface{}, len(configs))
	for i, config := range configs {
		configMaps[i] = userConfigurationToMap(config)
	}

	// Get all categories (no pagination limit - get everything)
	categories := category_mapper.INSTANCE.GetAllCategoriesIncludeDeleted(0, 0)
	stats.Categories.Success = len(categories)
	emit(progress, "read", "categories", len(categories), len(categories), "Categories collected")

	// Convert categories to map format for JSON serialization
	categoryMaps := make([]map[string]interface{}, len(categories))
	for i, cat := range categories {
		catMap := map[string]interface{}{
			"id":              cat.Id.Hex(),
			"belongs_user_id": cat.BelongsUserId.Hex(),
			"name":            cat.Name,
			"type":            cat.Type,
			"parent_id":       cat.ParentId.Hex(),
			"remark":          cat.Remark,
			"create_time":     cat.CreateTime,
			"update_time":     cat.UpdateTime,
			"create_user_id":  cat.CreateUserId.Hex(),
			"update_user_id":  cat.UpdateUserId.Hex(),
			"is_delete":       cat.IsDelete,
		}
		if cat.DeleteUserId != nil {
			catMap["delete_user_id"] = cat.DeleteUserId.Hex()
		} else {
			catMap["delete_user_id"] = nil
		}
		if cat.DeleteTime != nil {
			catMap["delete_time"] = cat.DeleteTime
		} else {
			catMap["delete_time"] = nil
		}
		categoryMaps[i] = catMap
	}

	// Get all cash flows (no pagination limit - get everything)
	cashFlows := cash_flow_mapper.INSTANCE.GetAllCashFlowsIncludeDeleted(0, 0)
	stats.CashFlows.Success = len(cashFlows)
	emit(progress, "read", "cash_flows", len(cashFlows), len(cashFlows), "Cash flows collected")

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
		} else {
			cfMap["delete_user_id"] = nil
		}
		if cf.DeleteTime != nil {
			cfMap["delete_time"] = cf.DeleteTime
		} else {
			cfMap["delete_time"] = nil
		}
		cashFlowMaps[i] = cfMap
	}

	// Create backup structure
	backup := BackupData{
		Version:     "1.0.0", // Version updated for user data isolation
		Timestamp:   time.Now().Format(time.RFC3339),
		Users:       userMaps,
		UserConfigs: configMaps,
		Categories:  categoryMaps,
		CashFlows:   cashFlowMaps,
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
	emit(progress, "write", "backup", 1, 1, "Backup file written")

	return stats, nil
}

// UserExportData creates a backup of data for a specific user (excludes deleted records)
func UserExportData(userId string, filePath string) (OperationStats, error) {
	return UserExportDataWithProgress(userId, filePath, nil)
}

func UserExportDataWithProgress(userId string, filePath string, progress ProgressFunc) (OperationStats, error) {
	stats := OperationStats{
		Users:       EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		UserConfigs: EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		Categories:  EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		CashFlows:   EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
	}

	if filePath == "" {
		return stats, errors.New("file path cannot be empty")
	}
	if err := PreflightBackupDestination(filePath); err != nil {
		return stats, err
	}
	emit(progress, "preflight", "backup", 1, 1, "Backup destination validated")

	// Get user
	user := user_mapper.INSTANCE.GetUserByObjectId(userId)
	if user.Id.IsZero() {
		return stats, errors.New("user not found")
	}
	stats.Users.Success = 1
	emit(progress, "read", "users", 1, 1, "User collected")

	// Convert user to map format
	userMaps := make([]map[string]interface{}, 1)
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

	if user.Nickname != "" {
		userMap["nickname"] = user.Nickname
	} else {
		userMap["nickname"] = nil
	}
	if user.AvatarUrl != "" {
		userMap["avatar_url"] = user.AvatarUrl
	} else {
		userMap["avatar_url"] = nil
	}
	if user.EmailAddress != "" {
		userMap["email_address"] = user.EmailAddress
	} else {
		userMap["email_address"] = nil
	}
	if user.Gender != "" {
		userMap["gender"] = user.Gender
	} else {
		userMap["gender"] = nil
	}

	// Include deletion info even for active user export (if they happen to be set for some reason,
	// or if user requested fields to be present)
	if user.DeleteUserId != nil {
		userMap["delete_user_id"] = user.DeleteUserId.Hex()
	}
	if user.DeleteTime != nil {
		// Format as RFC3339 string
		userMap["delete_time"] = user.DeleteTime.Format(time.RFC3339)
	}

	userMaps[0] = userMap

	var configMaps []map[string]interface{}
	config := user_config_mapper.INSTANCE.GetByUserId(user.Id.Hex())
	if !config.Id.IsZero() {
		stats.UserConfigs.Success = 1
		configMaps = []map[string]interface{}{userConfigurationToMap(config)}
	}
	emit(progress, "read", "user_configurations", stats.UserConfigs.Success, stats.UserConfigs.Success, "User configuration collected")

	// Get categories for user (Exclude deleted)
	// Use GetAllCategoriesByUser instead of IncludeDeleted version
	categories := category_mapper.INSTANCE.GetAllCategoriesByUser(user.Id, 0, 0)
	stats.Categories.Success = len(categories)
	emit(progress, "read", "categories", len(categories), len(categories), "Categories collected")

	// Convert categories to map format
	categoryMaps := make([]map[string]interface{}, len(categories))
	for i, cat := range categories {
		catMap := map[string]interface{}{
			"id":              cat.Id.Hex(),
			"belongs_user_id": cat.BelongsUserId.Hex(),
			"name":            cat.Name,
			"type":            cat.Type,
			"parent_id":       cat.ParentId.Hex(),
			"remark":          cat.Remark,
			"create_time":     cat.CreateTime,
			"update_time":     cat.UpdateTime,
			"create_user_id":  cat.CreateUserId.Hex(),
			"update_user_id":  cat.UpdateUserId.Hex(),
			"is_delete":       false, // Force false for export
		}
		categoryMaps[i] = catMap
	}

	// Get cash flows for user (Exclude deleted)
	// Use GetAllCashFlowsByUser instead of IncludeDeleted version
	cashFlows := cash_flow_mapper.INSTANCE.GetAllCashFlowsByUser(user.Id, 0, 0)
	stats.CashFlows.Success = len(cashFlows)
	emit(progress, "read", "cash_flows", len(cashFlows), len(cashFlows), "Cash flows collected")

	// Convert cash flows to map format
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
			"is_delete":      false, // Force false for export
		}
		cashFlowMaps[i] = cfMap
	}

	// Create backup structure
	backup := BackupData{
		Version:     "2.0.0",
		Timestamp:   time.Now().Format(time.RFC3339),
		Users:       userMaps,
		UserConfigs: configMaps,
		Categories:  categoryMaps,
		CashFlows:   cashFlowMaps,
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
	emit(progress, "write", "backup", 1, 1, "Backup file written")

	return stats, nil
}

func userConfigurationToMap(config model.UserConfigurationEntity) map[string]interface{} {
	configMap := map[string]interface{}{
		"id":                 config.Id.Hex(),
		"belongs_user_id":    config.BelongsUserId.Hex(),
		"display_language":   config.DisplayLanguage,
		"currency_code":      config.CurrencyCode,
		"active_theme_color": config.ActiveThemeColor,
		"create_time":        config.CreateTime,
		"update_time":        config.UpdateTime,
		"create_user_id":     config.CreateUserId.Hex(),
		"update_user_id":     config.UpdateUserId.Hex(),
		"is_delete":          config.IsDelete,
	}
	if config.DeleteUserId != nil {
		configMap["delete_user_id"] = config.DeleteUserId.Hex()
	} else {
		configMap["delete_user_id"] = nil
	}
	if config.DeleteTime != nil {
		configMap["delete_time"] = config.DeleteTime
	} else {
		configMap["delete_time"] = nil
	}
	return configMap
}
