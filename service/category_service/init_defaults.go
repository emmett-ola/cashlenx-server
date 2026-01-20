package category_service

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/macar-x/cashlenx-server/util"
)

// DefaultCategory represents a default category from configuration
type DefaultCategory struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Remark string `json:"remark"`
}

// DefaultCategoriesConfig represents the structure of default_categories.json
type DefaultCategoriesConfig struct {
	Categories []DefaultCategory `json:"categories"`
}

// LoadDefaultCategories loads default categories from config file
func LoadDefaultCategories() ([]DefaultCategory, error) {
	// Get the executable directory or working directory
	configPath := util.GetConfigByKey("default_categories.path")
	if configPath == "" {
		// Default path relative to the project root
		configPath = "config/default_categories.json"
	}

	// Try to read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		util.Logger.Warnw("Failed to read default categories config, using hardcoded defaults", "path", configPath, "error", err)
		return getHardcodedDefaultCategories(), nil
	}

	var config DefaultCategoriesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		util.Logger.Errorw("Failed to parse default categories config, using hardcoded defaults", "error", err)
		return getHardcodedDefaultCategories(), nil
	}

	if len(config.Categories) == 0 {
		util.Logger.Warn("No categories found in config, using hardcoded defaults")
		return getHardcodedDefaultCategories(), nil
	}

	util.Logger.Infof("Loaded %d default categories from config", len(config.Categories))
	return config.Categories, nil
}

// getHardcodedDefaultCategories returns hardcoded default categories as fallback
func getHardcodedDefaultCategories() []DefaultCategory {
	return []DefaultCategory{
		{Name: "Salary", Type: "INCOME", Remark: "Income from employment"},
		{Name: "Freelance", Type: "INCOME", Remark: "Income from freelance work"},
		{Name: "Investment", Type: "INCOME", Remark: "Income from investments and dividends"},
		{Name: "Other Income", Type: "INCOME", Remark: "Other income sources"},
		{Name: "Food & Dining", Type: "EXPENSE", Remark: "Restaurants, groceries, food delivery"},
		{Name: "Transportation", Type: "EXPENSE", Remark: "Gas, public transport, car maintenance"},
		{Name: "Shopping", Type: "EXPENSE", Remark: "Retail purchases, online shopping"},
		{Name: "Entertainment", Type: "EXPENSE", Remark: "Movies, games, hobbies"},
		{Name: "Healthcare", Type: "EXPENSE", Remark: "Medical expenses, pharmacy, fitness"},
		{Name: "Utilities", Type: "EXPENSE", Remark: "Electricity, water, internet, phone"},
	}
}

// InitializeDefaultCategoriesForUser creates default categories for a new user
func InitializeDefaultCategoriesForUser(userId string) error {
	// Load default categories from config
	defaultCategories, err := LoadDefaultCategories()
	if err != nil {
		return err
	}

	// Create each default category for the user
	successCount := 0
	for _, cat := range defaultCategories {
		_, err := CreateForUser(cat.Name, cat.Type, cat.Remark, "", userId)
		if err != nil {
			util.Logger.Warnw("Failed to create default category for user",
				"userId", userId,
				"category", cat.Name,
				"error", err)
			// Continue creating other categories even if one fails
			continue
		}
		successCount++
	}

	util.Logger.Infow("Initialized default categories for user",
		"userId", userId,
		"total", len(defaultCategories),
		"success", successCount)

	return nil
}

// GetDefaultCategoriesConfigPath returns the absolute path to default categories config
func GetDefaultCategoriesConfigPath() string {
	configPath := util.GetConfigByKey("default_categories.path")
	if configPath == "" {
		configPath = "config/default_categories.json"
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return configPath
	}
	return absPath
}
