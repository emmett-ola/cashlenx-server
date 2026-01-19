package manage_service

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RestoreBackup restores database from a backup file
func RestoreBackup(filePath string) (OperationStats, error) {
	stats := OperationStats{
		Users:      EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		Categories: EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
		CashFlows:  EntityStats{Success: 0, Failed: 0, FailedList: []string{}},
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
	totalCategories := len(backup.Categories)
	totalCashFlows := len(backup.CashFlows)
	stats.Users.Failed = totalUsers
	stats.Categories.Failed = totalCategories
	stats.CashFlows.Failed = totalCashFlows

	// Step 0: Clear existing data
	if _, err := ResetDatabase(); err != nil {
		return stats, err
	}

	// Step 1: Insert users from backup (must be first, as categories and cash flows reference users)
	for _, userMap := range backup.Users {
		// Parse Id from backup data
		id, _ := primitive.ObjectIDFromHex(userMap["Id"].(string))

		// Parse timestamps
		createdAt, _ := time.Parse(time.RFC3339, userMap["CreatedAt"].(string))
		updatedAt, _ := time.Parse(time.RFC3339, userMap["UpdatedAt"].(string))

		// Parse BaseEntity fields
		var createUserId primitive.ObjectID
		if cIdStr, ok := userMap["CreateUserId"].(string); ok && cIdStr != "" {
			createUserId, _ = primitive.ObjectIDFromHex(cIdStr)
		} else {
			// Default to self ID if missing
			createUserId = id
		}
		
		var updateUserId primitive.ObjectID
		if uIdStr, ok := userMap["UpdateUserId"].(string); ok && uIdStr != "" {
			updateUserId, _ = primitive.ObjectIDFromHex(uIdStr)
		} else {
			// Default to self ID if missing
			updateUserId = id
		}
		
		var deleteUserId *primitive.ObjectID
		if dIdStr, ok := userMap["DeleteUserId"].(string); ok && dIdStr != "" {
			dId, _ := primitive.ObjectIDFromHex(dIdStr)
			deleteUserId = &dId
		}
		
		var deleteTime *time.Time
		if dTimeStr, ok := userMap["DeleteTime"].(string); ok && dTimeStr != "" {
			dTime, _ := time.Parse(time.RFC3339, dTimeStr)
			deleteTime = &dTime
		}
		
		isDelete, _ := userMap["IsDelete"].(bool)

		// Create user entity from backup data, preserving all original fields
		userEntity := model.UserEntity{
			Id:           id,
			Username:     userMap["Username"].(string),
			PasswordHash: userMap["PasswordHash"].(string),
			IsActive:     userMap["IsActive"].(bool),
			Role:         userMap["Role"].(string),
			BaseEntity: model.BaseEntity{
				CreateTime:   createdAt,
				UpdateTime:   updatedAt,
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

	// Step 2: Insert categories from backup
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
			Type:     categoryType,
			Remark:   catMap["remark"].(string),
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

	// Step 3: Insert cash flows from backup
	cashFlowEntities := make([]model.CashFlowEntity, totalCashFlows)
	for i, cfMap := range backup.CashFlows {
		// Parse Id from backup data
		id, _ := primitive.ObjectIDFromHex(cfMap["id"].(string))

		// Parse UserId from backup data (with fallback for old backups without UserId)
		var userId primitive.ObjectID
		if userIdStr, ok := cfMap["user_id"].(string); ok && userIdStr != "" {
			userId, _ = primitive.ObjectIDFromHex(userIdStr)
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
			createUserId = userId
		}
		
		var updateUserId primitive.ObjectID
		if uIdStr, ok := cfMap["update_user_id"].(string); ok && uIdStr != "" {
			updateUserId, _ = primitive.ObjectIDFromHex(uIdStr)
		} else {
			// Default to user ID (owner) if missing
			updateUserId = userId
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
