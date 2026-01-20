package cash_flow_service

import (
	"errors"
	"strings"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SaveExpense(belongsDate, categoryName string, amount float64, description string, userId string) (model.CashFlowEntity, error) {
	// Validate inputs
	if err := validation.ValidateCategoryName(categoryName); err != nil {
		return model.CashFlowEntity{}, err
	}

	if err := validation.ValidateAmount(amount); err != nil {
		return model.CashFlowEntity{}, err
	}

	if belongsDate != "" {
		if err := validation.ValidateDate(belongsDate); err != nil {
			return model.CashFlowEntity{}, err
		}
	}

	if err := validation.ValidateDescription(description); err != nil {
		return model.CashFlowEntity{}, err
	}

	// Validate user ID
	if userId == "" {
		return model.CashFlowEntity{}, errors.New("user ID is required")
	}

	// Convert user ID to ObjectID
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return model.CashFlowEntity{}, errors.New("invalid user ID format")
	}

	// Round to 2 decimal places
	amount, _ = decimal.NewFromFloat(amount).Round(2).Float64()

	// Required parameter: category
	categoryEntity := category_mapper.INSTANCE.GetCategoryByNameAndUser(categoryName, userObjectId)
	if categoryEntity.IsEmpty() {
		return model.CashFlowEntity{}, errors.New("category does not exist or access denied")
	}

	// Validate category type matches cash flow type
	if !strings.EqualFold(categoryEntity.Type, model.FlowTypeExpense) {
		return model.CashFlowEntity{}, errors.New("category type mismatch: expected EXPENSE category")
	}

	// Optional parameter: date (default to today)
	var date time.Time
	if belongsDate != "" {
		// Parse the provided date using our multi-format parser
		parsedDate, err := util.ParseDate(belongsDate)
		if err != nil {
			return model.CashFlowEntity{}, errors.New("belongs_date error, try format like 19700101, 1970-01-01, or 1970/01/01")
		}
		// Use UTC time for consistent storage
		date = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		// Use today's date in UTC
		today := time.Now().UTC()
		date = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	}

	// Pre-generate ID before insertion
	newId := primitive.NewObjectID()

	insertedId := cash_flow_mapper.INSTANCE.InsertCashFlowByEntity(model.CashFlowEntity{
		Id:            newId,
		CategoryId:    categoryEntity.Id,
		BelongsDate:   date,
		Amount:        amount,
		Description:   description,
		BelongsUserId: userObjectId,
		BaseEntity: model.BaseEntity{
			CreateUserId: userObjectId,
			UpdateUserId: userObjectId,
		},
	})
	if insertedId == "" {
		return model.CashFlowEntity{}, errors.New("cash_flow create failed")
	}

	newCashFlow := cash_flow_mapper.INSTANCE.GetCashFlowByObjectId(newId.Hex())
	if !newCashFlow.IsEmpty() {
		newCashFlow.CategoryName = categoryEntity.Name
		newCashFlow.CategoryType = categoryEntity.Type
	}
	return newCashFlow, nil
}

func IsExpenseRequiredFiledSatisfied(categoryName string, amount float64) bool {
	if categoryName == "" {
		return false
	}
	if amount == 0 {
		return false
	}

	return true
}