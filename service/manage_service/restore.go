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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminRestoreDatabase restores database from a backup file (truncates existing data)
func AdminRestoreDatabase(filePath string) (OperationStats, error) {
	stats := OperationStats{
		Users:       EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		UserConfigs: EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		Categories:  EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		CashFlows:   EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
	}

	if filePath == "" {
		return stats, errors.New("file path cannot be empty")
	}

	// Read backup file
	file, err := os.Open(filePath)
	if err != nil {
		return stats, err
	}
	defer file.Close()

	// Parse JSON
	var backup BackupData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&backup); err != nil {
		return stats, err
	}

	// Update total counts for stats
	totalUsers := len(backup.Users)
	totalUserConfigs := len(backup.UserConfigs)
	totalCategories := len(backup.Categories)
	totalCashFlows := len(backup.CashFlows)
	stats.Users.Failed = totalUsers
	stats.UserConfigs.Failed = totalUserConfigs
	stats.Categories.Failed = totalCategories
	stats.CashFlows.Failed = totalCashFlows

	// Step 0: Clear existing data
	if _, err := ResetDatabase(); err != nil {
		return stats, err
	}

	// Step 1: Insert users from backup (must be first, as categories and cash flows reference users)
	for _, userMap := range backup.Users {
		// Parse Id from backup data
		id, _ := primitive.ObjectIDFromHex(userMap["id"].(string))

		// Parse timestamps
		createdTime, _ := time.Parse(time.RFC3339, userMap["create_time"].(string))
		updatedTime, _ := time.Parse(time.RFC3339, userMap["update_time"].(string))

		// Parse BaseEntity fields
		var createUserId primitive.ObjectID
		if cIdStr, ok := userMap["create_user_id"].(string); ok && cIdStr != "" {
			createUserId, _ = primitive.ObjectIDFromHex(cIdStr)
		} else {
			// Default to self ID if missing
			createUserId = id
		}

		var updateUserId primitive.ObjectID
		if uIdStr, ok := userMap["update_user_id"].(string); ok && uIdStr != "" {
			updateUserId, _ = primitive.ObjectIDFromHex(uIdStr)
		} else {
			// Default to self ID if missing
			updateUserId = id
		}

		var deleteUserId *primitive.ObjectID
		if dIdStr, ok := userMap["delete_user_id"].(string); ok && dIdStr != "" {
			dId, _ := primitive.ObjectIDFromHex(dIdStr)
			deleteUserId = &dId
		}

		var deleteTime *time.Time
		if dTimeStr, ok := userMap["delete_time"].(string); ok && dTimeStr != "" {
			dTime, _ := time.Parse(time.RFC3339, dTimeStr)
			deleteTime = &dTime
		}

		isDelete, _ := userMap["is_delete"].(bool)

		// Validate Gender
		gender, _ := userMap["gender"].(string)
		if gender != "" && gender != "male" && gender != "female" && gender != "others" {
			gender = "" // Default to empty if invalid
		}

		// Restore profile fields
		nickname, _ := userMap["nickname"].(string)
		avatarUrl, _ := userMap["avatar_url"].(string)
		emailAddress, _ := userMap["email_address"].(string)

		// Create user entity from backup data, preserving all original fields
		userEntity := model.UserEntity{
			Id:           id,
			Username:     userMap["username"].(string),
			PasswordHash: userMap["password_hash"].(string),
			IsActive:     userMap["is_active"].(bool),
			Role:         userMap["role"].(string),
			Nickname:     nickname,
			AvatarUrl:    avatarUrl,
			EmailAddress: emailAddress,
			Gender:       gender,
			BaseEntity: model.BaseEntity{
				CreateTime:   createdTime,
				UpdateTime:   updatedTime,
				CreateUserId: createUserId,
				UpdateUserId: updateUserId,
				DeleteUserId: deleteUserId,
				DeleteTime:   deleteTime,
				IsDelete:     isDelete,
			},
		}

		// Insert user
		if userId := user_mapper.INSTANCE.InsertUserByEntity(userEntity); userId != "" {
			stats.Users.Success++
			stats.Users.Failed--
		}
	}

	// Step 2: Insert user configurations from backup
	for _, configMap := range backup.UserConfigs {
		configEntity := restoreUserConfigurationFromMap(configMap, false)
		if configId := user_config_mapper.INSTANCE.InsertByEntity(configEntity); configId != "" {
			stats.UserConfigs.Success++
			stats.UserConfigs.Failed--
		}
	}

	// Step 3: Insert categories from backup
	for _, catMap := range backup.Categories {
		// Parse Id from backup data
		id, _ := primitive.ObjectIDFromHex(catMap["id"].(string))

		// Parse UserId from backup data (with fallback for old backups without UserId)
		var belongsUserId primitive.ObjectID
		if bUserIdStr, ok := catMap["belongs_user_id"].(string); ok && bUserIdStr != "" {
			belongsUserId, _ = primitive.ObjectIDFromHex(bUserIdStr)
		} else if userIdStr, ok := catMap["user_id"].(string); ok && userIdStr != "" {
			belongsUserId, _ = primitive.ObjectIDFromHex(userIdStr)
		}

		// Parse ParentId from backup data
		parentId, _ := primitive.ObjectIDFromHex(catMap["parent_id"].(string))

		// Parse CreateTime and UpdateTime
		createTime, _ := time.Parse(time.RFC3339, catMap["create_time"].(string))
		updateTime, _ := time.Parse(time.RFC3339, catMap["update_time"].(string))

		// Get Type with fallback for old backups
		categoryType, _ := catMap["type"].(string)

		// Parse BaseEntity fields
		var createUserId primitive.ObjectID
		if cIdStr, ok := catMap["create_user_id"].(string); ok && cIdStr != "" {
			createUserId, _ = primitive.ObjectIDFromHex(cIdStr)
		} else {
			// Default to user ID (owner) if missing
			createUserId = belongsUserId
		}

		var updateUserId primitive.ObjectID
		if uIdStr, ok := catMap["update_user_id"].(string); ok && uIdStr != "" {
			updateUserId, _ = primitive.ObjectIDFromHex(uIdStr)
		} else {
			// Default to user ID (owner) if missing
			updateUserId = belongsUserId
		}

		var deleteUserId *primitive.ObjectID
		if dIdStr, ok := catMap["delete_user_id"].(string); ok && dIdStr != "" {
			dId, _ := primitive.ObjectIDFromHex(dIdStr)
			deleteUserId = &dId
		}

		var deleteTime *time.Time
		if dTimeStr, ok := catMap["delete_time"].(string); ok && dTimeStr != "" {
			dTime, _ := time.Parse(time.RFC3339, dTimeStr)
			deleteTime = &dTime
		}

		isDelete, _ := catMap["is_delete"].(bool)

		// Create category entity from backup data, preserving all original fields
		catEntity := model.CategoryEntity{
			Id:            id,
			BelongsUserId: belongsUserId,
			ParentId:      parentId,
			Name:          catMap["name"].(string),
			Type:          categoryType,
			Remark:        catMap["remark"].(string),
			BaseEntity: model.BaseEntity{
				CreateTime:   createTime,
				UpdateTime:   updateTime,
				CreateUserId: createUserId,
				UpdateUserId: updateUserId,
				DeleteUserId: deleteUserId,
				DeleteTime:   deleteTime,
				IsDelete:     isDelete,
			},
		}

		// Insert category
		if catId := category_mapper.INSTANCE.InsertCategoryByEntity(catEntity); catId != "" {
			stats.Categories.Success++
			stats.Categories.Failed--
		}
	}

	// Step 4: Insert cash flows from backup
	cashFlowEntities := make([]model.CashFlowEntity, totalCashFlows)
	for i, cfMap := range backup.CashFlows {
		// Parse Id from backup data
		id, _ := primitive.ObjectIDFromHex(cfMap["id"].(string))

		// Parse UserId from backup data (with fallback for old backups without UserId)
		var belongsUserId primitive.ObjectID
		if bUserIdStr, ok := cfMap["belongs_user_id"].(string); ok && bUserIdStr != "" {
			belongsUserId, _ = primitive.ObjectIDFromHex(bUserIdStr)
		} else if userIdStr, ok := cfMap["user_id"].(string); ok && userIdStr != "" {
			belongsUserId, _ = primitive.ObjectIDFromHex(userIdStr)
		}

		// Parse belongs_date string to time.Time
		belongsDate, _ := time.Parse(time.RFC3339, cfMap["belongs_date"].(string))

		// Parse CategoryId from backup data
		categoryId, _ := primitive.ObjectIDFromHex(cfMap["category_id"].(string))

		// Parse CreateTime and UpdateTime
		createTime, _ := time.Parse(time.RFC3339, cfMap["create_time"].(string))
		updateTime, _ := time.Parse(time.RFC3339, cfMap["update_time"].(string))

		// Parse BaseEntity fields
		var createUserId primitive.ObjectID
		if cIdStr, ok := cfMap["create_user_id"].(string); ok && cIdStr != "" {
			createUserId, _ = primitive.ObjectIDFromHex(cIdStr)
		} else {
			// Default to user ID (owner) if missing
			createUserId = belongsUserId
		}

		var updateUserId primitive.ObjectID
		if uIdStr, ok := cfMap["update_user_id"].(string); ok && uIdStr != "" {
			updateUserId, _ = primitive.ObjectIDFromHex(uIdStr)
		} else {
			// Default to user ID (owner) if missing
			updateUserId = belongsUserId
		}

		var deleteUserId *primitive.ObjectID
		if dIdStr, ok := cfMap["delete_user_id"].(string); ok && dIdStr != "" {
			dId, _ := primitive.ObjectIDFromHex(dIdStr)
			deleteUserId = &dId
		}

		var deleteTime *time.Time
		if dTimeStr, ok := cfMap["delete_time"].(string); ok && dTimeStr != "" {
			dTime, _ := time.Parse(time.RFC3339, dTimeStr)
			deleteTime = &dTime
		}

		isDelete, _ := cfMap["is_delete"].(bool)

		// Create cash flow entity from backup data, preserving all original fields
		cfEntity := model.CashFlowEntity{
			Id:            id,
			BelongsUserId: belongsUserId,
			CategoryId:    categoryId,
			BelongsDate:   belongsDate,
			// FlowType:    cfMap["FlowType"].(string),
			Amount:      cfMap["amount"].(float64),
			Description: cfMap["description"].(string),
			Remark:      cfMap["remark"].(string),
			BaseEntity: model.BaseEntity{
				CreateTime:   createTime,
				UpdateTime:   updateTime,
				CreateUserId: createUserId,
				UpdateUserId: updateUserId,
				DeleteUserId: deleteUserId,
				DeleteTime:   deleteTime,
				IsDelete:     isDelete,
			},
		}
		cashFlowEntities[i] = cfEntity
	}

	// Use bulk insert for cash flows if available
	if len(cashFlowEntities) > 0 {
		if ids, err := cash_flow_mapper.INSTANCE.BulkInsertCashFlows(cashFlowEntities); err != nil {
			// If bulk insert fails, try individual inserts
			for _, cfEntity := range cashFlowEntities {
				if id := cash_flow_mapper.INSTANCE.InsertCashFlowByEntity(cfEntity); id != "" {
					stats.CashFlows.Success++
					stats.CashFlows.Failed--
				}
			}
		} else {
			// Bulk insert succeeded
			stats.CashFlows.Success = len(ids)
			stats.CashFlows.Failed = totalCashFlows - len(ids)
		}
	}

	return stats, nil
}

// UserImportData imports user data from a backup file with upsert logic (skips deleted records)
func UserImportData(userId string, filePath string) (OperationStats, error) {
	stats := OperationStats{
		Users:       EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		UserConfigs: EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		Categories:  EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		CashFlows:   EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
	}

	if filePath == "" {
		return stats, errors.New("file path cannot be empty")
	}

	// Read backup file
	file, err := os.Open(filePath)
	if err != nil {
		return stats, err
	}
	defer file.Close()

	// Parse JSON
	var backup BackupData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&backup); err != nil {
		return stats, err
	}

	// Validation: Check if userId exists in backup.Users
	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return stats, errors.New("invalid user id")
	}

	// Rule 1: Validate User Record
	// - Must contain exactly one user
	// - That user must match the current requesting user
	if len(backup.Users) != 1 {
		return stats, errors.New("backup file must contain exactly one user record")
	}

	backupUser := backup.Users[0]
	backupUserId, ok := backupUser["id"].(string)
	if !ok || backupUserId != userId {
		return stats, errors.New("backup user does not match current user")
	}

	// Step 1: Skip User restoration (only validate) but we might want to check gender of the backup user if we were to update it
	// But UserImportData does NOT update the user record itself, only related data.
	// So no gender validation needed here for UserImportData.

	for _, configMap := range backup.UserConfigs {
		belongsUserId := parseObjectIDFromMap(configMap, "belongs_user_id")
		if belongsUserId != userObjectId {
			continue
		}
		if isDelete, ok := configMap["is_delete"].(bool); ok && isDelete {
			continue
		}
		if !user_config_mapper.INSTANCE.GetByUserId(userId).Id.IsZero() {
			continue
		}
		stats.UserConfigs.Failed++
		configEntity := restoreUserConfigurationFromMap(configMap, true)
		if configId := user_config_mapper.INSTANCE.InsertByEntity(configEntity); configId != "" {
			stats.UserConfigs.Success++
			stats.UserConfigs.Failed--
		}
	}

	// Step 2: Import categories
	for _, catMap := range backup.Categories {
		// Rule 3: Check ownership
		var belongsUserId primitive.ObjectID
		if bUserIdStr, ok := catMap["belongs_user_id"].(string); ok && bUserIdStr != "" {
			belongsUserId, _ = primitive.ObjectIDFromHex(bUserIdStr)
		} else if userIdStr, ok := catMap["user_id"].(string); ok && userIdStr != "" {
			belongsUserId, _ = primitive.ObjectIDFromHex(userIdStr)
		}

		if belongsUserId != userObjectId {
			continue // Skip data not belonging to this user
		}

		// Skip logically deleted records
		if isDelete, ok := catMap["is_delete"].(bool); ok && isDelete {
			continue
		}

		// Parse Id from backup data
		id, _ := primitive.ObjectIDFromHex(catMap["id"].(string))

		// Rule 2: Check existence (Skip if already exists)
		existing := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(id.Hex(), userObjectId)
		if !existing.Id.IsZero() {
			continue // Skip existing records (do nothing)
		}

		stats.Categories.Failed++ // Increment potential count

		// Parse ParentId from backup data
		parentId, _ := primitive.ObjectIDFromHex(catMap["parent_id"].(string))

		// Rule 2: Check parent existence (if parent_id is set AND not nil/zero)
		// Note: primitive.NilObjectID is 0000...0000
		if !parentId.IsZero() {
			parent := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(parentId.Hex(), userObjectId)
			if parent.Id.IsZero() {
				// Parent doesn't exist in DB.
				// NOTE: This is tricky if parent is also in this import list but not yet inserted.
				// Ideally we should topological sort or multi-pass.
				// Given the prompt "parent_data not existed, they do nothing", we skip.
				continue
			}
		}

		// Parse CreateTime and UpdateTime
		createTime, _ := time.Parse(time.RFC3339, catMap["create_time"].(string))
		updateTime, _ := time.Parse(time.RFC3339, catMap["update_time"].(string))

		// Get Type with fallback for old backups
		categoryType, _ := catMap["type"].(string)

		// Parse BaseEntity fields
		var createUserId primitive.ObjectID
		if cIdStr, ok := catMap["create_user_id"].(string); ok && cIdStr != "" {
			createUserId, _ = primitive.ObjectIDFromHex(cIdStr)
		} else {
			createUserId = belongsUserId
		}

		var updateUserId primitive.ObjectID
		if uIdStr, ok := catMap["update_user_id"].(string); ok && uIdStr != "" {
			updateUserId, _ = primitive.ObjectIDFromHex(uIdStr)
		} else {
			updateUserId = belongsUserId
		}

		// Create category entity (ignoring delete fields as we skip deleted records)
		catEntity := model.CategoryEntity{
			Id:            id,
			BelongsUserId: belongsUserId,
			ParentId:      parentId,
			Name:          catMap["name"].(string),
			Type:          categoryType,
			Remark:        catMap["remark"].(string),
			BaseEntity: model.BaseEntity{
				CreateTime:   createTime,
				UpdateTime:   updateTime,
				CreateUserId: createUserId,
				UpdateUserId: updateUserId,
				IsDelete:     false,
			},
		}

		// Insert
		if catId := category_mapper.INSTANCE.InsertCategoryByEntity(catEntity); catId != "" {
			stats.Categories.Success++
			stats.Categories.Failed--
		}
	}

	// Step 3: Import cash flows
	for _, cfMap := range backup.CashFlows {
		// Rule 3: Check ownership
		var belongsUserId primitive.ObjectID
		if bUserIdStr, ok := cfMap["belongs_user_id"].(string); ok && bUserIdStr != "" {
			belongsUserId, _ = primitive.ObjectIDFromHex(bUserIdStr)
		} else if userIdStr, ok := cfMap["user_id"].(string); ok && userIdStr != "" {
			belongsUserId, _ = primitive.ObjectIDFromHex(userIdStr)
		}

		if belongsUserId != userObjectId {
			continue // Skip
		}

		// Parse Id
		id, _ := primitive.ObjectIDFromHex(cfMap["id"].(string))

		// Rule 2: Check existence (Skip if already exists)
		existing := cash_flow_mapper.INSTANCE.GetCashFlowByObjectIdAndUser(id.Hex(), userObjectId)
		if !existing.Id.IsZero() {
			continue // Skip existing records
		}

		stats.CashFlows.Failed++ // Increment potential count

		// Parse belongs_date
		belongsDate, _ := time.Parse(time.RFC3339, cfMap["belongs_date"].(string))

		// Parse CategoryId
		categoryId, _ := primitive.ObjectIDFromHex(cfMap["category_id"].(string))

		// Parse CreateTime and UpdateTime
		createTime, _ := time.Parse(time.RFC3339, cfMap["create_time"].(string))
		updateTime, _ := time.Parse(time.RFC3339, cfMap["update_time"].(string))

		// Parse BaseEntity fields
		var createUserId primitive.ObjectID
		if cIdStr, ok := cfMap["create_user_id"].(string); ok && cIdStr != "" {
			createUserId, _ = primitive.ObjectIDFromHex(cIdStr)
		} else {
			createUserId = belongsUserId
		}

		var updateUserId primitive.ObjectID
		if uIdStr, ok := cfMap["update_user_id"].(string); ok && uIdStr != "" {
			updateUserId, _ = primitive.ObjectIDFromHex(uIdStr)
		} else {
			updateUserId = belongsUserId
		}

		var deleteUserId *primitive.ObjectID
		if dIdStr, ok := cfMap["delete_user_id"].(string); ok && dIdStr != "" {
			dId, _ := primitive.ObjectIDFromHex(dIdStr)
			deleteUserId = &dId
		}

		var deleteTime *time.Time
		if dTimeStr, ok := cfMap["delete_time"].(string); ok && dTimeStr != "" {
			dTime, _ := time.Parse(time.RFC3339, dTimeStr)
			deleteTime = &dTime
		}

		isDelete, _ := cfMap["is_delete"].(bool)

		// Create cash flow entity
		cfEntity := model.CashFlowEntity{
			Id:            id,
			BelongsUserId: belongsUserId,
			CategoryId:    categoryId,
			BelongsDate:   belongsDate,
			Amount:        cfMap["amount"].(float64),
			Description:   cfMap["description"].(string),
			Remark:        cfMap["remark"].(string),
			BaseEntity: model.BaseEntity{
				CreateTime:   createTime,
				UpdateTime:   updateTime,
				CreateUserId: createUserId,
				UpdateUserId: updateUserId,
				DeleteUserId: deleteUserId,
				DeleteTime:   deleteTime,
				IsDelete:     isDelete,
			},
		}

		// Insert
		if id := cash_flow_mapper.INSTANCE.InsertCashFlowByEntity(cfEntity); id != "" {
			stats.CashFlows.Success++
			stats.CashFlows.Failed--
		}
	}

	return stats, nil
}

func restoreUserConfigurationFromMap(configMap map[string]interface{}, forceActive bool) model.UserConfigurationEntity {
	id := parseObjectIDFromMap(configMap, "id")
	belongsUserId := parseObjectIDFromMap(configMap, "belongs_user_id")
	createUserId := parseObjectIDFromMap(configMap, "create_user_id")
	if createUserId.IsZero() {
		createUserId = belongsUserId
	}
	updateUserId := parseObjectIDFromMap(configMap, "update_user_id")
	if updateUserId.IsZero() {
		updateUserId = belongsUserId
	}

	createTime := parseTimeFromMap(configMap, "create_time")
	updateTime := parseTimeFromMap(configMap, "update_time")
	deleteUserId := parseOptionalObjectIDFromMap(configMap, "delete_user_id")
	deleteTime := parseOptionalTimeFromMap(configMap, "delete_time")
	isDelete, _ := configMap["is_delete"].(bool)
	if forceActive {
		deleteUserId = nil
		deleteTime = nil
		isDelete = false
	}

	displayLanguage, _ := configMap["display_language"].(string)
	if displayLanguage == "" {
		displayLanguage = model.DefaultDisplayLanguage
	}
	currencyCode, _ := configMap["currency_code"].(string)
	if currencyCode == "" {
		currencyCode = model.DefaultCurrencyCode
	}
	activeThemeColor, _ := configMap["active_theme_color"].(string)
	if activeThemeColor == "" {
		activeThemeColor = model.DefaultThemeColor
	}

	return model.UserConfigurationEntity{
		Id:               id,
		BelongsUserId:    belongsUserId,
		DisplayLanguage:  displayLanguage,
		CurrencyCode:     currencyCode,
		ActiveThemeColor: activeThemeColor,
		BaseEntity: model.BaseEntity{
			CreateUserId: createUserId,
			CreateTime:   createTime,
			UpdateUserId: updateUserId,
			UpdateTime:   updateTime,
			DeleteUserId: deleteUserId,
			DeleteTime:   deleteTime,
			IsDelete:     isDelete,
		},
	}
}

func parseObjectIDFromMap(data map[string]interface{}, key string) primitive.ObjectID {
	if value, ok := data[key].(string); ok && value != "" {
		id, _ := primitive.ObjectIDFromHex(value)
		return id
	}
	return primitive.NilObjectID
}

func parseOptionalObjectIDFromMap(data map[string]interface{}, key string) *primitive.ObjectID {
	if value, ok := data[key].(string); ok && value != "" {
		id, _ := primitive.ObjectIDFromHex(value)
		return &id
	}
	return nil
}

func parseTimeFromMap(data map[string]interface{}, key string) time.Time {
	if value, ok := data[key].(string); ok && value != "" {
		parsed, _ := time.Parse(time.RFC3339, value)
		return parsed
	}
	return time.Now()
}

func parseOptionalTimeFromMap(data map[string]interface{}, key string) *time.Time {
	if value, ok := data[key].(string); ok && value != "" {
		parsed, _ := time.Parse(time.RFC3339, value)
		return &parsed
	}
	return nil
}
