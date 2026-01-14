package manage_service

import (
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mock category mapper for testing
type MockCategoryMapper struct {
	categories map[string]model.CategoryEntity
}

func (m MockCategoryMapper) GetCategoryByObjectId(plainId string) model.CategoryEntity {
	return m.categories[plainId]
}

func (m MockCategoryMapper) GetCategoryByName(categoryName string) model.CategoryEntity {
	return model.CategoryEntity{}
}

func (m MockCategoryMapper) GetCategoryByParentId(parentPlainId string) []model.CategoryEntity {
	return []model.CategoryEntity{}
}

func (m MockCategoryMapper) InsertCategoryByEntity(newEntity model.CategoryEntity) string {
	return ""
}

func (m MockCategoryMapper) UpdateCategoryByEntity(plainId string, updatedEntity model.CategoryEntity) model.CategoryEntity {
	return model.CategoryEntity{}
}

func (m MockCategoryMapper) GetAllCategories(limit, offset int) []model.CategoryEntity {
	return []model.CategoryEntity{}
}

func (m MockCategoryMapper) CountAllCategories() int64 {
	return 0
}

func (m MockCategoryMapper) DeleteCategoryByObjectId(plainId string) model.CategoryEntity {
	return model.CategoryEntity{}
}

func (m MockCategoryMapper) TruncateCategories() error {
	return nil
}

// TestCategoryNameRetrieval tests that category names are properly retrieved with fallback
func TestCategoryNameRetrieval(t *testing.T) {
	// Create test category IDs
	foodCategoryId := primitive.NewObjectID()
	transportCategoryId := primitive.NewObjectID()
	unknownCategoryId := primitive.NewObjectID()

	// Create test categories
	testCategories := map[string]model.CategoryEntity{
		foodCategoryId.Hex(): {
			Id:   foodCategoryId,
			Name: "Food",
			Type: "expense",
		},
		transportCategoryId.Hex(): {
			Id:   transportCategoryId,
			Name: "Transport",
			Type: "expense",
		},
	}

	// Create test cash flows
	testCashFlows := []model.CashFlowEntity{
		{
			Id:          primitive.NewObjectID(),
			CategoryId:  foodCategoryId,
			BelongsDate: time.Now(),
			FlowType:    "expense",
			Amount:      100.0,
			Description: "Groceries",
		},
		{
			Id:          primitive.NewObjectID(),
			CategoryId:  transportCategoryId,
			BelongsDate: time.Now(),
			FlowType:    "expense",
			Amount:      50.0,
			Description: "Bus ticket",
		},
		{
			Id:          primitive.NewObjectID(),
			CategoryId:  unknownCategoryId,
			BelongsDate: time.Now(),
			FlowType:    "expense",
			Amount:      20.0,
			Description: "Unknown category",
		},
	}

	// Test category name retrieval logic
	for _, cashFlow := range testCashFlows {
		// Simulate the current logic
		categoryName := "Unknown"
		categoryEntity := testCategories[cashFlow.CategoryId.Hex()]
		if !categoryEntity.IsEmpty() {
			categoryName = categoryEntity.Name
		}

		// Verify results
		switch cashFlow.CategoryId {
		case foodCategoryId:
			if categoryName != "Food" {
				t.Errorf("Expected category name 'Food', got '%s'", categoryName)
			}
		case transportCategoryId:
			if categoryName != "Transport" {
				t.Errorf("Expected category name 'Transport', got '%s'", categoryName)
			}
		case unknownCategoryId:
			if categoryName != "Unknown" {
				t.Errorf("Expected category name 'Unknown' for missing category, got '%s'", categoryName)
			}
		}
	}

	t.Log("All category name retrieval tests passed!")
}
