package manage_service

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserMapper
type MockUserMapper struct {
	users []model.UserEntity
}
func (m MockUserMapper) GetUserByObjectId(plainId string) model.UserEntity { return model.UserEntity{} }
func (m MockUserMapper) GetUserByUsername(username string) model.UserEntity { return model.UserEntity{} }
func (m MockUserMapper) GetUserByUsernameIncludeDeleted(username string) model.UserEntity { return model.UserEntity{} }
func (m MockUserMapper) GetUserByEmail(email string) model.UserEntity { return model.UserEntity{} }
func (m MockUserMapper) InsertUserByEntity(newEntity model.UserEntity) string { return "" }
func (m MockUserMapper) UpdateUserByEntity(plainId string, updatedEntity model.UserEntity) model.UserEntity { return model.UserEntity{} }
func (m MockUserMapper) GetAllUsers(limit, offset int) []model.UserEntity { return []model.UserEntity{} }
func (m MockUserMapper) GetAllUsersIncludeDeleted(limit, offset int) []model.UserEntity { return m.users }
func (m MockUserMapper) GetUsersByRole(role string) []model.UserEntity { return []model.UserEntity{} }
func (m MockUserMapper) CountAllUsers() int64 { return 0 }
func (m MockUserMapper) DeleteUserByObjectId(plainId string) model.UserEntity { return model.UserEntity{} }
func (m MockUserMapper) TruncateUsers() error { return nil }

// MockCategoryMapper
type MockBackupCategoryMapper struct {
	categories []model.CategoryEntity
}
func (m MockBackupCategoryMapper) GetCategoryByObjectId(plainId string) model.CategoryEntity { return model.CategoryEntity{} }
func (m MockBackupCategoryMapper) GetCategoryByName(categoryName string) model.CategoryEntity { return model.CategoryEntity{} }
func (m MockBackupCategoryMapper) GetCategoryByParentId(parentPlainId string) []model.CategoryEntity { return []model.CategoryEntity{} }
func (m MockBackupCategoryMapper) InsertCategoryByEntity(newEntity model.CategoryEntity) string { return "" }
func (m MockBackupCategoryMapper) UpdateCategoryByEntity(plainId string, updatedEntity model.CategoryEntity) model.CategoryEntity { return model.CategoryEntity{} }
func (m MockBackupCategoryMapper) GetAllCategories(limit, offset int) []model.CategoryEntity { return []model.CategoryEntity{} }
func (m MockBackupCategoryMapper) GetAllCategoriesIncludeDeleted(limit, offset int) []model.CategoryEntity { return m.categories }
func (m MockBackupCategoryMapper) CountAllCategories() int64 { return 0 }
func (m MockBackupCategoryMapper) CountCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) (int64, error) { return 0, nil }
func (m MockBackupCategoryMapper) DeleteCategoryByObjectId(plainId string) model.CategoryEntity { return model.CategoryEntity{} }
func (m MockBackupCategoryMapper) TruncateCategories() error { return nil }
func (m MockBackupCategoryMapper) GetAllCategoriesByUser(userId primitive.ObjectID, limit, offset int) []model.CategoryEntity { return []model.CategoryEntity{} }
func (m MockBackupCategoryMapper) GetCategoryByObjectIdAndUser(id string, userId primitive.ObjectID) model.CategoryEntity { return model.CategoryEntity{} }
func (m MockBackupCategoryMapper) CountAllCategoriesByUser(userId primitive.ObjectID) int64 { return 0 }
func (m MockBackupCategoryMapper) GetCategoriesByUserAndType(userId primitive.ObjectID, categoryType string, limit, offset int) ([]model.CategoryEntity, error) { return []model.CategoryEntity{}, nil }
func (m MockBackupCategoryMapper) GetRootCategoriesByUser(userId primitive.ObjectID) ([]model.CategoryEntity, error) { return []model.CategoryEntity{}, nil }
func (m MockBackupCategoryMapper) GetRootCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) { return []model.CategoryEntity{}, nil }
func (m MockBackupCategoryMapper) GetCategoriesByParentIdAndUser(parentId primitive.ObjectID, userId primitive.ObjectID) ([]model.CategoryEntity, error) { return []model.CategoryEntity{}, nil }
func (m MockBackupCategoryMapper) GetCategoriesByParentIdUserAndType(parentId primitive.ObjectID, userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) { return []model.CategoryEntity{}, nil }
func (m MockBackupCategoryMapper) GetCategoryByNameAndUser(categoryName string, userId primitive.ObjectID) model.CategoryEntity { return model.CategoryEntity{} }
func (m MockBackupCategoryMapper) GetCategoryByNameUserAndType(categoryName string, userId primitive.ObjectID, categoryType string) model.CategoryEntity { return model.CategoryEntity{} }
func (m MockBackupCategoryMapper) GetCategoryByNameUserTypeAndParent(categoryName string, userId primitive.ObjectID, categoryType string, parentId primitive.ObjectID) model.CategoryEntity { return model.CategoryEntity{} }
func (m MockBackupCategoryMapper) DeleteCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity { return model.CategoryEntity{} }
func (m MockBackupCategoryMapper) UpdateCategoryByEntityAndUser(plainId string, updatedEntity model.CategoryEntity, userId primitive.ObjectID) model.CategoryEntity { return model.CategoryEntity{} }
func (m MockBackupCategoryMapper) GetAllCategoriesByUserIncludeDeleted(userId primitive.ObjectID) []model.CategoryEntity { return []model.CategoryEntity{} }
func (m MockBackupCategoryMapper) DeleteAllCategoriesByUser(userId primitive.ObjectID) (int64, error) { return 0, nil }


// MockCashFlowMapper
type MockCashFlowMapper struct {
	cashFlows []model.CashFlowEntity
}
func (m MockCashFlowMapper) GetCashFlowByObjectId(plainId string) model.CashFlowEntity { return model.CashFlowEntity{} }
func (m MockCashFlowMapper) InsertCashFlowByEntity(newEntity model.CashFlowEntity) string { return "" }
func (m MockCashFlowMapper) BulkInsertCashFlows(newEntities []model.CashFlowEntity) ([]string, error) { return []string{}, nil }
func (m MockCashFlowMapper) UpdateCashFlowByEntity(plainId string, updatedEntity model.CashFlowEntity) model.CashFlowEntity { return model.CashFlowEntity{} }
func (m MockCashFlowMapper) GetAllCashFlows(limit, offset int) []model.CashFlowEntity { return []model.CashFlowEntity{} }
func (m MockCashFlowMapper) GetAllCashFlowsIncludeDeleted(limit, offset int) []model.CashFlowEntity { return m.cashFlows }
func (m MockCashFlowMapper) CountAllCashFlows() int64 { return 0 }
func (m MockCashFlowMapper) DeleteCashFlowByObjectId(plainId string) model.CashFlowEntity { return model.CashFlowEntity{} }
func (m MockCashFlowMapper) TruncateCashFlows() error { return nil }
func (m MockCashFlowMapper) GetAllCashFlowsByUser(userId primitive.ObjectID, limit, offset int) []model.CashFlowEntity { return []model.CashFlowEntity{} }
func (m MockCashFlowMapper) GetCashFlowByObjectIdAndUser(id string, userId primitive.ObjectID) model.CashFlowEntity { return model.CashFlowEntity{} }
func (m MockCashFlowMapper) GetCashFlowsByCategoryIdAndUser(categoryId string, userId primitive.ObjectID) []model.CashFlowEntity { return []model.CashFlowEntity{} }
func (m MockCashFlowMapper) GetCashFlowsByDateRangeAndUser(startDate, endDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity { return []model.CashFlowEntity{} }
func (m MockCashFlowMapper) CountAllCashFlowsByUser(userId primitive.ObjectID) int64 { return 0 }
func (m MockCashFlowMapper) GetCashFlowsByObjectIdArray(plainIdList []string) []model.CashFlowEntity { return []model.CashFlowEntity{} }
func (m MockCashFlowMapper) GetCashFlowsByBelongsDate(belongsDate time.Time) []model.CashFlowEntity { return []model.CashFlowEntity{} }
func (m MockCashFlowMapper) DeleteCashFlowByBelongsDate(belongsDate time.Time) []model.CashFlowEntity { return []model.CashFlowEntity{} }
func (m MockCashFlowMapper) GetCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity { return []model.CashFlowEntity{} }
func (m MockCashFlowMapper) DeleteCashFlowByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CashFlowEntity { return model.CashFlowEntity{} }
func (m MockCashFlowMapper) DeleteCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity { return []model.CashFlowEntity{} }
func (m MockCashFlowMapper) DeleteCashFlowsByCategoryIdAndUser(categoryPlainId string, userId primitive.ObjectID) int64 { return 0 }
func (m MockCashFlowMapper) UpdateCashFlowByEntityAndUser(plainId string, updatedEntity model.CashFlowEntity, userId primitive.ObjectID) model.CashFlowEntity { return model.CashFlowEntity{} }
func (m MockCashFlowMapper) GetCashFlowsByFilter(filter model.CashFlowFilter) ([]model.CashFlowEntity, error) { return []model.CashFlowEntity{}, nil }
func (m MockCashFlowMapper) CountCashFlowsByFilter(filter model.CashFlowFilter) (int64, error) { return 0, nil }
func (m MockCashFlowMapper) GetAllCashFlowsByUserIncludeDeleted(userId primitive.ObjectID) []model.CashFlowEntity { return []model.CashFlowEntity{} }
func (m MockCashFlowMapper) DeleteAllCashFlowsByUser(userId primitive.ObjectID) (int64, error) { return 0, nil }

func TestAdminDumpDatabase(t *testing.T) {
	// Save original instances
	origUserMapper := user_mapper.INSTANCE
	origCategoryMapper := category_mapper.INSTANCE
	origCashFlowMapper := cash_flow_mapper.INSTANCE
	defer func() {
		user_mapper.INSTANCE = origUserMapper
		category_mapper.INSTANCE = origCategoryMapper
		cash_flow_mapper.INSTANCE = origCashFlowMapper
	}()

	// Setup mock data
	userId := primitive.NewObjectID()
	deleteUserId := primitive.NewObjectID()
	now := time.Now()

	users := []model.UserEntity{
		{
			Id:           userId,
			Username:     "testuser",
			Nickname:     "Test User",
			AvatarUrl:    "http://avatar.com/1.png",
			BaseEntity: model.BaseEntity{
				IsDelete:     true,
				DeleteUserId: &deleteUserId,
				DeleteTime:   &now,
			},
		},
		{
			Id:           primitive.NewObjectID(),
			Username:     "activeuser",
			Nickname:     "Active User",
			BaseEntity: model.BaseEntity{
				IsDelete: false,
			},
		},
	}

	user_mapper.INSTANCE = MockUserMapper{users: users}
	category_mapper.INSTANCE = MockBackupCategoryMapper{categories: []model.CategoryEntity{}}
	cash_flow_mapper.INSTANCE = MockCashFlowMapper{cashFlows: []model.CashFlowEntity{}}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "backup_test_*.json")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Run dump
	_, err = AdminDumpDatabase(tmpPath)
	if err != nil {
		t.Fatalf("AdminDumpDatabase failed: %v", err)
	}

	// Read and verify
	bytes, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}

	var backup BackupData
	if err := json.Unmarshal(bytes, &backup); err != nil {
		t.Fatal(err)
	}

	if len(backup.Users) != 2 {
		t.Fatalf("Expected 2 users, got %d", len(backup.Users))
	}

	// Verify deleted user
	deletedUser := backup.Users[0]
	if val, ok := deletedUser["nickname"]; !ok || val != "Test User" {
		t.Errorf("Deleted user missing nickname or wrong value: %v", val)
	}
	if val, ok := deletedUser["avatar_url"]; !ok || val != "http://avatar.com/1.png" {
		t.Errorf("Deleted user missing avatar_url or wrong value: %v", val)
	}
	if val, ok := deletedUser["delete_user_id"]; !ok || val != deleteUserId.Hex() {
		t.Errorf("Deleted user missing delete_user_id or wrong value: %v", val)
	}
	// Verify delete_time is present
	if _, ok := deletedUser["delete_time"]; !ok {
		t.Errorf("Deleted user missing delete_time")
	}

	// Verify active user
	activeUser := backup.Users[1]
	if val, ok := activeUser["delete_user_id"]; !ok || val != nil {
		t.Errorf("Active user expected nil delete_user_id, got %v", val)
	}
	if val, ok := activeUser["delete_time"]; !ok || val != nil {
		t.Errorf("Active user expected nil delete_time, got %v", val)
	}
	
	// Verify null handling for empty strings
	if val, ok := activeUser["avatar_url"]; !ok || val != nil {
		t.Errorf("Active user expected nil avatar_url for empty string, got %v", val)
	}
	
	// Test gender validation (mock data setup modification needed if we want to test the mapper logic specifically,
	// but here we are testing backup service logic. The validation happens in mapper.)
	// Since we are mocking the mapper in this test, we can't test the validation logic here directly
	// unless we update the mock to simulate invalid gender.
	// But let's verify that valid gender is passed through.
	if val, ok := deletedUser["gender"]; ok && val == "invalid_gender" {
		t.Errorf("Invalid gender should have been filtered out (if mapper logic was active)")
	}
}
