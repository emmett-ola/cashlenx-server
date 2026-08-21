package budget_service

import (
	"fmt"
	"math"
	"regexp"
	"time"

	appErrors "github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/mapper/budget_mapper"
	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var periodPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

type Service struct {
	budgets    budget_mapper.BudgetMapper
	cashFlows  cash_flow_mapper.CashFlowMapper
	categories category_mapper.CategoryMapper
}

func New(budgetMapper budget_mapper.BudgetMapper, cashFlowMapper cash_flow_mapper.CashFlowMapper, categoryMapper category_mapper.CategoryMapper) *Service {
	return &Service{budgets: budgetMapper, cashFlows: cashFlowMapper, categories: categoryMapper}
}

func defaultService() *Service {
	return New(budget_mapper.INSTANCE, cash_flow_mapper.INSTANCE, category_mapper.INSTANCE)
}

func CreateForUser(request model.UpsertBudgetRequest, userID string) (model.BudgetView, error) {
	return defaultService().CreateForUser(request, userID)
}

func ListForUser(period, userID string) ([]model.BudgetView, error) {
	return defaultService().ListForUser(period, userID)
}

func GetForUser(id, userID string) (model.BudgetView, error) {
	return defaultService().GetForUser(id, userID)
}

func UpdateForUser(id string, request model.UpsertBudgetRequest, userID string) (model.BudgetView, error) {
	return defaultService().UpdateForUser(id, request, userID)
}

func DeleteForUser(id, userID string) error { return defaultService().DeleteForUser(id, userID) }

func validateRequest(request model.UpsertBudgetRequest) error {
	if _, err := time.Parse("2006-01", request.Period); err != nil || !periodPattern.MatchString(request.Period) {
		return appErrors.NewFieldValidationError("period", "period must use YYYY-MM")
	}
	if math.IsNaN(request.LimitAmount) || math.IsInf(request.LimitAmount, 0) || request.LimitAmount <= 0 {
		return appErrors.NewFieldValidationError("limit_amount", "limit amount must be greater than zero")
	}
	if _, err := primitive.ObjectIDFromHex(request.CategoryId); err != nil {
		return appErrors.NewFieldValidationError("category_id", "invalid category ID")
	}
	return nil
}

func objectID(value, field string) (primitive.ObjectID, error) {
	id, err := primitive.ObjectIDFromHex(value)
	if err != nil {
		return primitive.NilObjectID, appErrors.NewFieldValidationError(field, fmt.Sprintf("invalid %s", field))
	}
	return id, nil
}

func (s *Service) validateCategory(categoryID, userID primitive.ObjectID) (model.CategoryEntity, error) {
	category := s.categories.GetCategoryByObjectIdAndUser(categoryID.Hex(), userID)
	if category.IsEmpty() {
		return model.CategoryEntity{}, appErrors.NewNotFoundError("expense category not found")
	}
	if category.Type != "expense" {
		return model.CategoryEntity{}, appErrors.NewInvalidInputError("budgets can only target expense categories")
	}
	return category, nil
}

func (s *Service) CreateForUser(request model.UpsertBudgetRequest, plainUserID string) (model.BudgetView, error) {
	if err := validateRequest(request); err != nil {
		return model.BudgetView{}, err
	}
	userID, err := objectID(plainUserID, "user ID")
	if err != nil {
		return model.BudgetView{}, err
	}
	categoryID, _ := primitive.ObjectIDFromHex(request.CategoryId)
	if _, err := s.validateCategory(categoryID, userID); err != nil {
		return model.BudgetView{}, err
	}
	existing, err := s.budgets.GetByScope(userID, categoryID, request.Period)
	if err != nil {
		return model.BudgetView{}, appErrors.NewDatabaseError("query budget", err)
	}
	if !existing.IsEmpty() {
		return model.BudgetView{}, appErrors.NewAlreadyExistsError("budget already exists for this category and period")
	}
	now := time.Now().UTC()
	entity, err := s.budgets.Insert(model.BudgetEntity{Id: primitive.NewObjectID(), BelongsUserId: userID, CategoryId: categoryID, Period: request.Period, LimitAmount: request.LimitAmount, BaseEntity: model.BaseEntity{CreateUserId: userID, CreateTime: now, UpdateUserId: userID, UpdateTime: now}})
	if err != nil {
		return model.BudgetView{}, appErrors.NewDatabaseError("create budget", err)
	}
	return s.view(entity, userID), nil
}

func (s *Service) ListForUser(period, plainUserID string) ([]model.BudgetView, error) {
	if period == "" {
		period = time.Now().Format("2006-01")
	}
	if !periodPattern.MatchString(period) {
		return nil, appErrors.NewFieldValidationError("period", "period must use YYYY-MM")
	}
	userID, err := objectID(plainUserID, "user ID")
	if err != nil {
		return nil, err
	}
	entities, err := s.budgets.ListByUserAndPeriod(userID, period)
	if err != nil {
		return nil, appErrors.NewDatabaseError("list budgets", err)
	}
	views := make([]model.BudgetView, 0, len(entities))
	for _, entity := range entities {
		views = append(views, s.view(entity, userID))
	}
	return views, nil
}

func (s *Service) GetForUser(plainID, plainUserID string) (model.BudgetView, error) {
	id, err := objectID(plainID, "budget ID")
	if err != nil {
		return model.BudgetView{}, err
	}
	userID, err := objectID(plainUserID, "user ID")
	if err != nil {
		return model.BudgetView{}, err
	}
	entity, err := s.budgets.GetByIDAndUser(id, userID)
	if err != nil {
		return model.BudgetView{}, appErrors.NewDatabaseError("query budget", err)
	}
	if entity.IsEmpty() {
		return model.BudgetView{}, appErrors.NewNotFoundError("budget not found")
	}
	return s.view(entity, userID), nil
}

func (s *Service) UpdateForUser(plainID string, request model.UpsertBudgetRequest, plainUserID string) (model.BudgetView, error) {
	if err := validateRequest(request); err != nil {
		return model.BudgetView{}, err
	}
	id, err := objectID(plainID, "budget ID")
	if err != nil {
		return model.BudgetView{}, err
	}
	userID, err := objectID(plainUserID, "user ID")
	if err != nil {
		return model.BudgetView{}, err
	}
	entity, err := s.budgets.GetByIDAndUser(id, userID)
	if err != nil {
		return model.BudgetView{}, appErrors.NewDatabaseError("query budget", err)
	}
	if entity.IsEmpty() {
		return model.BudgetView{}, appErrors.NewNotFoundError("budget not found")
	}
	categoryID, _ := primitive.ObjectIDFromHex(request.CategoryId)
	if _, err := s.validateCategory(categoryID, userID); err != nil {
		return model.BudgetView{}, err
	}
	conflict, err := s.budgets.GetByScope(userID, categoryID, request.Period)
	if err != nil {
		return model.BudgetView{}, appErrors.NewDatabaseError("query budget", err)
	}
	if !conflict.IsEmpty() && conflict.Id != entity.Id {
		return model.BudgetView{}, appErrors.NewAlreadyExistsError("budget already exists for this category and period")
	}
	entity.CategoryId = categoryID
	entity.Period = request.Period
	entity.LimitAmount = request.LimitAmount
	entity.UpdateUserId = userID
	entity.UpdateTime = time.Now().UTC()
	entity, err = s.budgets.Update(entity)
	if err != nil || entity.IsEmpty() {
		return model.BudgetView{}, appErrors.NewDatabaseError("update budget", err)
	}
	return s.view(entity, userID), nil
}

func (s *Service) DeleteForUser(plainID, plainUserID string) error {
	id, err := objectID(plainID, "budget ID")
	if err != nil {
		return err
	}
	userID, err := objectID(plainUserID, "user ID")
	if err != nil {
		return err
	}
	deleted, err := s.budgets.Delete(id, userID, userID)
	if err != nil {
		return appErrors.NewDatabaseError("delete budget", err)
	}
	if !deleted {
		return appErrors.NewNotFoundError("budget not found")
	}
	return nil
}

func (s *Service) view(entity model.BudgetEntity, userID primitive.ObjectID) model.BudgetView {
	category := s.categories.GetCategoryByObjectIdAndUser(entity.CategoryId.Hex(), userID)
	start, _ := time.Parse("2006-01", entity.Period)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	spent := 0.0
	for _, flow := range s.cashFlows.GetCashFlowsByDateRangeAndUser(start, end, userID) {
		if flow.CategoryId == entity.CategoryId {
			spent += flow.Amount
		}
	}
	remaining := entity.LimitAmount - spent
	progress := 0.0
	if entity.LimitAmount > 0 {
		progress = spent / entity.LimitAmount
	}
	return model.BudgetView{Id: entity.Id.Hex(), CategoryId: entity.CategoryId.Hex(), CategoryName: category.Name, Period: entity.Period, LimitAmount: entity.LimitAmount, SpentAmount: spent, Remaining: remaining, Progress: progress}
}
