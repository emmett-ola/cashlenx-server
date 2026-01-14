package manage_service

import (
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/macar-x/cashlenx-server/util"
)

// InitializeDemoData initializes the database with demo categories and transactions
func InitializeDemoData() error {
	userId := "demo_user"
	// Create default categories with their types
	categories := []struct {
		name  string
		type_ string
	}{
		{"Salary", "income"},
		{"Freelance", "income"},
		{"Food & Dining", "expense"},
		{"Transportation", "expense"},
		{"Shopping", "expense"},
		{"Entertainment", "expense"},
		{"Healthcare", "expense"},
		{"Utilities", "expense"},
	}

	for _, cat := range categories {
		// Check if category already exists
		_, err := category_service.CreateForUser(cat.name, cat.type_, "", "", userId)
		if err != nil {
			util.Logger.Warnw("category creation skipped", "category", cat.name, "error", err)
		}
	}

	// Create sample transactions for the past week
	today := time.Now()

	// Sample income
	_, _ = cash_flow_service.SaveIncome(
		today.AddDate(0, 0, -7).Format(model.DateFormatYYYYMMDD),
		"Salary",
		5000.00,
		"Monthly salary",
		userId,
	)

	// Sample expenses
	expenses := []struct {
		daysAgo     int
		category    string
		amount      float64
		description string
	}{
		{1, "Food & Dining", 45.50, "Lunch"},
		{1, "Transportation", 20.00, "Bus fare"},
		{2, "Food & Dining", 32.00, "Groceries"},
		{3, "Entertainment", 50.00, "Movie tickets"},
		{4, "Shopping", 120.00, "Clothes"},
		{5, "Healthcare", 80.00, "Pharmacy"},
		{6, "Utilities", 150.00, "Electricity bill"},
	}

	for _, exp := range expenses {
		date := today.AddDate(0, 0, -exp.daysAgo).Format(model.DateFormatYYYYMMDD)
		_, _ = cash_flow_service.SaveExpense(date, exp.category, exp.amount, exp.description, userId)
	}

	return nil
}
