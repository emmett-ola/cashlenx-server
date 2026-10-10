package cash_flow_service

import (
	"sort"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCashFlowServiceSaveExpenseCreatesAndEnrichesFlow(t *testing.T) {
	cashMapper := newCashFlowMapperFake()
	categoryMapper := newCashFlowCategoryMapperFake()
	userID := primitive.NewObjectID()
	categoryID := primitive.NewObjectID()
	categoryMapper.categories[categoryID] = model.CategoryEntity{
		Id:            categoryID,
		BelongsUserId: userID,
		Name:          "Food",
		Type:          model.FlowTypeExpense,
		Emoji:         "🍜",
		BgColor:       "#FF8A65",
	}
	service := NewCashFlowService(cashMapper, categoryMapper)

	created, err := service.SaveExpense("2026-05-12", "Food", 12.345, "lunch", userID.Hex())
	if err != nil {
		t.Fatalf("SaveExpense returned error: %v", err)
	}

	if created.Amount != 12.35 {
		t.Fatalf("Amount = %.2f, want 12.35", created.Amount)
	}
	if created.CategoryName != "Food" || created.CategoryType != model.FlowTypeExpense {
		t.Fatalf("category enrichment = %q/%q", created.CategoryName, created.CategoryType)
	}
	if created.CategoryEmoji != "🍜" || created.CategoryBgColor != "#FF8A65" {
		t.Fatalf("category presentation = %q/%q", created.CategoryEmoji, created.CategoryBgColor)
	}
	if created.BelongsUserId != userID {
		t.Fatalf("BelongsUserId = %s, want %s", created.BelongsUserId.Hex(), userID.Hex())
	}
}

func TestCashFlowServiceSaveIncomeRejectsExpenseCategory(t *testing.T) {
	cashMapper := newCashFlowMapperFake()
	categoryMapper := newCashFlowCategoryMapperFake()
	userID := primitive.NewObjectID()
	categoryID := primitive.NewObjectID()
	categoryMapper.categories[categoryID] = model.CategoryEntity{
		Id:            categoryID,
		BelongsUserId: userID,
		Name:          "Food",
		Type:          model.FlowTypeExpense,
	}
	service := NewCashFlowService(cashMapper, categoryMapper)

	_, err := service.SaveIncome("2026-05-12", "Food", 12, "wrong type", userID.Hex())
	if err == nil {
		t.Fatal("expected category type mismatch error")
	}
	if len(cashMapper.flows) != 0 {
		t.Fatalf("flow count = %d, want 0", len(cashMapper.flows))
	}
}

func TestCashFlowServiceQueryAllForUserEnrichesAndFiltersByType(t *testing.T) {
	cashMapper := newCashFlowMapperFake()
	categoryMapper := newCashFlowCategoryMapperFake()
	userID := primitive.NewObjectID()
	foodID := primitive.NewObjectID()
	salaryID := primitive.NewObjectID()
	categoryMapper.categories[foodID] = model.CategoryEntity{Id: foodID, BelongsUserId: userID, Name: "Food", Type: model.FlowTypeExpense, Emoji: "🍜", BgColor: "#FF8A65"}
	categoryMapper.categories[salaryID] = model.CategoryEntity{Id: salaryID, BelongsUserId: userID, Name: "Salary", Type: model.FlowTypeIncome}
	cashMapper.flows[primitive.NewObjectID().Hex()] = model.CashFlowEntity{
		Id:            primitive.NewObjectID(),
		BelongsUserId: userID,
		CategoryId:    foodID,
		BelongsDate:   time.Date(2026, time.May, 12, 0, 0, 0, 0, time.UTC),
		Amount:        25,
		Description:   "dinner",
	}
	cashMapper.flows[primitive.NewObjectID().Hex()] = model.CashFlowEntity{
		Id:            primitive.NewObjectID(),
		BelongsUserId: userID,
		CategoryId:    salaryID,
		BelongsDate:   time.Date(2026, time.May, 13, 0, 0, 0, 0, time.UTC),
		Amount:        100,
		Description:   "pay",
	}
	service := NewCashFlowService(cashMapper, categoryMapper)

	results, total, err := service.QueryAllForUser(userID.Hex(), model.FlowTypeExpense, "", "", "", "", "", 20, 0)
	if err != nil {
		t.Fatalf("QueryAllForUser returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("total = %d, want type-filtered count 1", total)
	}
	if len(results) != 1 {
		t.Fatalf("results length = %d, want 1", len(results))
	}
	if results[0].CategoryName != "Food" || results[0].CategoryType != model.FlowTypeExpense {
		t.Fatalf("category enrichment = %q/%q", results[0].CategoryName, results[0].CategoryType)
	}
	if results[0].CategoryEmoji != "🍜" || results[0].CategoryBgColor != "#FF8A65" {
		t.Fatalf("category presentation = %q/%q", results[0].CategoryEmoji, results[0].CategoryBgColor)
	}
}

func TestCashFlowServiceFiltersByDescriptionOrCategoryBeforePagination(t *testing.T) {
	cashMapper := newCashFlowMapperFake()
	categoryMapper := newCashFlowCategoryMapperFake()
	userID := primitive.NewObjectID()
	foodID := primitive.NewObjectID()
	salaryID := primitive.NewObjectID()
	categoryMapper.categories[foodID] = model.CategoryEntity{Id: foodID, BelongsUserId: userID, Name: "Food", Type: model.FlowTypeExpense}
	categoryMapper.categories[salaryID] = model.CategoryEntity{Id: salaryID, BelongsUserId: userID, Name: "Salary", Type: model.FlowTypeIncome}

	baseDate := time.Date(2026, time.May, 20, 0, 0, 0, 0, time.UTC)
	for index, fixture := range []struct {
		categoryID  primitive.ObjectID
		description string
	}{
		{foodID, "Market groceries"},
		{foodID, "Dinner"},
		{salaryID, "Monthly pay"},
	} {
		flowID := primitive.NewObjectID()
		cashMapper.flows[flowID.Hex()] = model.CashFlowEntity{
			Id: flowID, BelongsUserId: userID, CategoryId: fixture.categoryID,
			BelongsDate: baseDate.AddDate(0, 0, -index), Description: fixture.description,
		}
	}
	service := NewCashFlowService(cashMapper, categoryMapper)

	categoryPage, categoryTotal, err := service.QueryAllForUser(userID.Hex(), "", "", "foo", "", "", "", 1, 1)
	if err != nil {
		t.Fatalf("category search returned error: %v", err)
	}
	if categoryTotal != 2 || len(categoryPage) != 1 || categoryPage[0].CategoryName != "Food" {
		t.Fatalf("category search page = %+v, total = %d", categoryPage, categoryTotal)
	}

	descriptionPage, descriptionTotal, err := service.QueryAllForUser(userID.Hex(), "", "", "MARKET", "", "", "", 20, 0)
	if err != nil {
		t.Fatalf("description search returned error: %v", err)
	}
	if descriptionTotal != 1 || len(descriptionPage) != 1 || descriptionPage[0].Description != "Market groceries" {
		t.Fatalf("description search page = %+v, total = %d", descriptionPage, descriptionTotal)
	}
}

func TestCashFlowServiceUsesDeletedCategoryPresentationFallback(t *testing.T) {
	cashMapper := newCashFlowMapperFake()
	categoryMapper := newCashFlowCategoryMapperFake()
	userID := primitive.NewObjectID()
	flowID := primitive.NewObjectID()
	cashMapper.flows[flowID.Hex()] = model.CashFlowEntity{
		Id: flowID, BelongsUserId: userID, CategoryId: primitive.NewObjectID(),
	}
	service := NewCashFlowService(cashMapper, categoryMapper)

	result, err := service.QueryByIdForUser(flowID.Hex(), userID.Hex())
	if err != nil {
		t.Fatalf("QueryByIdForUser returned error: %v", err)
	}
	if result.CategoryName != "Unknown" || result.CategoryType != "" {
		t.Fatalf("deleted category fallback = %q/%q", result.CategoryName, result.CategoryType)
	}
	if result.CategoryEmoji != model.DefaultCategoryEmoji || result.CategoryBgColor != model.DefaultCategoryBackgroundColor {
		t.Fatalf("deleted category presentation = %q/%q", result.CategoryEmoji, result.CategoryBgColor)
	}
}

func TestCashFlowServiceQueryAllForUserUsesStableBusinessDateOrder(t *testing.T) {
	cashMapper := newCashFlowMapperFake()
	categoryMapper := newCashFlowCategoryMapperFake()
	userID := primitive.NewObjectID()
	categoryID := primitive.NewObjectID()
	categoryMapper.categories[categoryID] = model.CategoryEntity{
		Id:            categoryID,
		BelongsUserId: userID,
		Name:          "Food",
		Type:          model.FlowTypeExpense,
	}
	day := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	olderID := primitive.NewObjectID()
	newerID := primitive.NewObjectID()
	previousDayID := primitive.NewObjectID()
	cashMapper.flows[olderID.Hex()] = model.CashFlowEntity{
		Id: olderID, BelongsUserId: userID, CategoryId: categoryID,
		BelongsDate: day, BaseEntity: model.BaseEntity{CreateTime: day.Add(time.Hour)},
	}
	cashMapper.flows[newerID.Hex()] = model.CashFlowEntity{
		Id: newerID, BelongsUserId: userID, CategoryId: categoryID,
		BelongsDate: day, BaseEntity: model.BaseEntity{CreateTime: day.Add(2 * time.Hour)},
	}
	cashMapper.flows[previousDayID.Hex()] = model.CashFlowEntity{
		Id: previousDayID, BelongsUserId: userID, CategoryId: categoryID,
		BelongsDate: day.AddDate(0, 0, -1), BaseEntity: model.BaseEntity{CreateTime: day.Add(3 * time.Hour)},
	}

	service := NewCashFlowService(cashMapper, categoryMapper)
	firstPage, total, err := service.QueryAllForUser(userID.Hex(), "", "", "", "", "", "", 2, 0)
	if err != nil {
		t.Fatalf("QueryAllForUser returned error: %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if len(firstPage) != 2 || firstPage[0].Id != newerID || firstPage[1].Id != olderID {
		t.Fatalf("first page order = %#v, want newer same-day flow before older flow", firstPage)
	}
	secondPage, _, err := service.QueryAllForUser(userID.Hex(), "", "", "", "", "", "", 2, 2)
	if err != nil {
		t.Fatalf("second QueryAllForUser returned error: %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].Id != previousDayID {
		t.Fatalf("second page = %#v, want previous-day flow", secondPage)
	}
}

func TestCashFlowServiceGetTotalSummaryForUserAggregatesActiveUserFlows(t *testing.T) {
	cashMapper := newCashFlowMapperFake()
	categoryMapper := newCashFlowCategoryMapperFake()
	userID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()
	foodID := primitive.NewObjectID()
	salaryID := primitive.NewObjectID()
	categoryMapper.categories[foodID] = model.CategoryEntity{Id: foodID, BelongsUserId: userID, Name: "Food", Type: model.FlowTypeExpense}
	categoryMapper.categories[salaryID] = model.CategoryEntity{Id: salaryID, BelongsUserId: userID, Name: "Salary", Type: model.FlowTypeIncome}
	cashMapper.flows[primitive.NewObjectID().Hex()] = model.CashFlowEntity{
		Id:            primitive.NewObjectID(),
		BelongsUserId: userID,
		CategoryId:    foodID,
		BelongsDate:   time.Date(2026, time.May, 12, 0, 0, 0, 0, time.UTC),
		Amount:        12.345,
	}
	cashMapper.flows[primitive.NewObjectID().Hex()] = model.CashFlowEntity{
		Id:            primitive.NewObjectID(),
		BelongsUserId: userID,
		CategoryId:    salaryID,
		BelongsDate:   time.Date(2026, time.May, 13, 0, 0, 0, 0, time.UTC),
		Amount:        100,
	}
	cashMapper.flows[primitive.NewObjectID().Hex()] = model.CashFlowEntity{
		Id:            primitive.NewObjectID(),
		BelongsUserId: otherUserID,
		CategoryId:    foodID,
		BelongsDate:   time.Date(2026, time.May, 14, 0, 0, 0, 0, time.UTC),
		Amount:        99,
	}
	cashMapper.flows[primitive.NewObjectID().Hex()] = model.CashFlowEntity{
		Id:            primitive.NewObjectID(),
		BelongsUserId: userID,
		CategoryId:    foodID,
		BelongsDate:   time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC),
		Amount:        20,
		BaseEntity:    model.BaseEntity{IsDelete: true},
	}
	service := NewCashFlowService(cashMapper, categoryMapper)

	summary, err := service.GetTotalSummaryForUser(userID.Hex())
	if err != nil {
		t.Fatalf("GetTotalSummaryForUser returned error: %v", err)
	}

	if summary.TotalIncome != 100 {
		t.Fatalf("TotalIncome = %.2f, want 100.00", summary.TotalIncome)
	}
	if summary.TotalExpense != 12.35 {
		t.Fatalf("TotalExpense = %.2f, want 12.35", summary.TotalExpense)
	}
	if summary.Balance != 87.66 {
		t.Fatalf("Balance = %.2f, want 87.66", summary.Balance)
	}
	if summary.TransactionCount != 2 {
		t.Fatalf("TransactionCount = %d, want 2", summary.TransactionCount)
	}
	if summary.CategoryBreakdown["Food"] != 12.35 {
		t.Fatalf("Food category = %.2f, want 12.35", summary.CategoryBreakdown["Food"])
	}
}

func TestCashFlowServiceUpdateByIdForUserUsesUserScopedCategory(t *testing.T) {
	cashMapper := newCashFlowMapperFake()
	categoryMapper := newCashFlowCategoryMapperFake()
	userID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()
	foodID := primitive.NewObjectID()
	travelID := primitive.NewObjectID()
	otherTravelID := primitive.NewObjectID()
	flowID := primitive.NewObjectID()
	categoryMapper.categories[foodID] = model.CategoryEntity{Id: foodID, BelongsUserId: userID, Name: "Food", Type: model.FlowTypeExpense}
	categoryMapper.categories[travelID] = model.CategoryEntity{Id: travelID, BelongsUserId: userID, Name: "Travel", Type: model.FlowTypeExpense}
	categoryMapper.categories[otherTravelID] = model.CategoryEntity{Id: otherTravelID, BelongsUserId: otherUserID, Name: "Travel", Type: model.FlowTypeExpense}
	cashMapper.flows[flowID.Hex()] = model.CashFlowEntity{
		Id:            flowID,
		BelongsUserId: userID,
		CategoryId:    foodID,
		BelongsDate:   time.Date(2026, time.May, 12, 0, 0, 0, 0, time.UTC),
		Amount:        25,
		Description:   "dinner",
	}
	service := NewCashFlowService(cashMapper, categoryMapper)

	updated, err := service.UpdateByIdForUser(flowID.Hex(), "", "Travel", 0, "", userID.Hex())
	if err != nil {
		t.Fatalf("UpdateByIdForUser returned error: %v", err)
	}

	if updated.CategoryId != travelID {
		t.Fatalf("CategoryId = %s, want user's Travel category %s", updated.CategoryId.Hex(), travelID.Hex())
	}
	if updated.UpdateUserId != userID {
		t.Fatalf("UpdateUserId = %s, want %s", updated.UpdateUserId.Hex(), userID.Hex())
	}
}

type cashFlowMapperFake struct {
	flows map[string]model.CashFlowEntity
}

func newCashFlowMapperFake() *cashFlowMapperFake {
	return &cashFlowMapperFake{flows: map[string]model.CashFlowEntity{}}
}

func (fake *cashFlowMapperFake) GetCashFlowByObjectId(plainId string) model.CashFlowEntity {
	return fake.flows[plainId]
}

func (fake *cashFlowMapperFake) GetCashFlowsByObjectIdArray(plainIdList []string) []model.CashFlowEntity {
	var flows []model.CashFlowEntity
	for _, id := range plainIdList {
		if flow := fake.flows[id]; !flow.IsEmpty() {
			flows = append(flows, flow)
		}
	}
	return flows
}

func (fake *cashFlowMapperFake) GetCashFlowsByBelongsDate(belongsDate time.Time) []model.CashFlowEntity {
	var flows []model.CashFlowEntity
	for _, flow := range fake.flows {
		if sameDay(flow.BelongsDate, belongsDate) {
			flows = append(flows, flow)
		}
	}
	return flows
}

func (fake *cashFlowMapperFake) InsertCashFlowByEntity(newEntity model.CashFlowEntity) string {
	fake.flows[newEntity.Id.Hex()] = newEntity
	return newEntity.Id.Hex()
}

func (fake *cashFlowMapperFake) BulkInsertCashFlows(entities []model.CashFlowEntity) ([]string, error) {
	ids := make([]string, 0, len(entities))
	for _, entity := range entities {
		ids = append(ids, fake.InsertCashFlowByEntity(entity))
	}
	return ids, nil
}

func (fake *cashFlowMapperFake) UpdateCashFlowByEntity(plainId string, updatedEntity model.CashFlowEntity) model.CashFlowEntity {
	fake.flows[plainId] = updatedEntity
	return updatedEntity
}

func (fake *cashFlowMapperFake) GetAllCashFlows(limit, offset int) []model.CashFlowEntity {
	return pageCashFlows(fake.allFlows(), limit, offset)
}

func (fake *cashFlowMapperFake) GetAllCashFlowsIncludeDeleted(limit, offset int) []model.CashFlowEntity {
	return fake.GetAllCashFlows(limit, offset)
}

func (fake *cashFlowMapperFake) CountAllCashFlows() int64 {
	return int64(len(fake.flows))
}

func (fake *cashFlowMapperFake) DeleteCashFlowByObjectId(plainId string) model.CashFlowEntity {
	flow := fake.flows[plainId]
	delete(fake.flows, plainId)
	return flow
}

func (fake *cashFlowMapperFake) DeleteCashFlowByBelongsDate(belongsDate time.Time) []model.CashFlowEntity {
	var deleted []model.CashFlowEntity
	for id, flow := range fake.flows {
		if sameDay(flow.BelongsDate, belongsDate) {
			deleted = append(deleted, flow)
			delete(fake.flows, id)
		}
	}
	return deleted
}

func (fake *cashFlowMapperFake) TruncateCashFlows() error {
	fake.flows = map[string]model.CashFlowEntity{}
	return nil
}

func (fake *cashFlowMapperFake) GetCashFlowByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CashFlowEntity {
	flow := fake.flows[plainId]
	if flow.BelongsUserId == userId {
		return flow
	}
	return model.CashFlowEntity{}
}

func (fake *cashFlowMapperFake) GetCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	var flows []model.CashFlowEntity
	for _, flow := range fake.flows {
		if flow.BelongsUserId == userId && sameDay(flow.BelongsDate, belongsDate) {
			flows = append(flows, flow)
		}
	}
	return flows
}

func (fake *cashFlowMapperFake) GetCashFlowsByDateRangeAndUser(from, to time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	var flows []model.CashFlowEntity
	for _, flow := range fake.flows {
		if flow.BelongsUserId == userId && !flow.BelongsDate.Before(from) && flow.BelongsDate.Before(to) {
			flows = append(flows, flow)
		}
	}
	return flows
}

func (fake *cashFlowMapperFake) GetCashFlowsByCategoryIdAndUser(categoryPlainId string, userId primitive.ObjectID) []model.CashFlowEntity {
	var flows []model.CashFlowEntity
	for _, flow := range fake.flows {
		if flow.BelongsUserId == userId && flow.CategoryId.Hex() == categoryPlainId {
			flows = append(flows, flow)
		}
	}
	return flows
}

func (fake *cashFlowMapperFake) GetAllCashFlowsByUser(userId primitive.ObjectID, limit, offset int) []model.CashFlowEntity {
	var flows []model.CashFlowEntity
	for _, flow := range fake.flows {
		if flow.BelongsUserId == userId {
			flows = append(flows, flow)
		}
	}
	return pageCashFlows(flows, limit, offset)
}

func (fake *cashFlowMapperFake) CountAllCashFlowsByUser(userId primitive.ObjectID) int64 {
	return int64(len(fake.GetAllCashFlowsByUser(userId, 0, 0)))
}

func (fake *cashFlowMapperFake) DeleteCashFlowByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CashFlowEntity {
	flow := fake.GetCashFlowByObjectIdAndUser(plainId, userId)
	if !flow.IsEmpty() {
		delete(fake.flows, plainId)
	}
	return flow
}

func (fake *cashFlowMapperFake) DeleteCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	var deleted []model.CashFlowEntity
	for id, flow := range fake.flows {
		if flow.BelongsUserId == userId && sameDay(flow.BelongsDate, belongsDate) {
			deleted = append(deleted, flow)
			delete(fake.flows, id)
		}
	}
	return deleted
}

func (fake *cashFlowMapperFake) DeleteCashFlowsByCategoryIdAndUser(categoryPlainId string, userId primitive.ObjectID) int64 {
	var deleted int64
	for id, flow := range fake.flows {
		if flow.BelongsUserId == userId && flow.CategoryId.Hex() == categoryPlainId {
			delete(fake.flows, id)
			deleted++
		}
	}
	return deleted
}

func (fake *cashFlowMapperFake) UpdateCashFlowByEntityAndUser(plainId string, updatedEntity model.CashFlowEntity, userId primitive.ObjectID) model.CashFlowEntity {
	if existing := fake.GetCashFlowByObjectIdAndUser(plainId, userId); existing.IsEmpty() {
		return model.CashFlowEntity{}
	}
	fake.flows[plainId] = updatedEntity
	return updatedEntity
}

func (fake *cashFlowMapperFake) GetCashFlowsByFilter(filter model.CashFlowFilter) ([]model.CashFlowEntity, error) {
	var flows []model.CashFlowEntity
	for _, flow := range fake.flows {
		if flow.IsDelete {
			continue
		}
		if !filter.UserId.IsZero() && flow.BelongsUserId != filter.UserId {
			continue
		}
		if filter.CategoryId != "" && flow.CategoryId.Hex() != filter.CategoryId {
			continue
		}
		if filter.ExactDescription != "" && flow.Description != filter.ExactDescription {
			continue
		}
		if filter.Description != "" && flow.Description != filter.Description {
			continue
		}
		if !filter.FromDate.IsZero() && flow.BelongsDate.Before(filter.FromDate) {
			continue
		}
		if !filter.ToDate.IsZero() && flow.BelongsDate.After(filter.ToDate) {
			continue
		}
		flows = append(flows, flow)
	}
	sort.Slice(flows, func(i, j int) bool {
		if !flows[i].BelongsDate.Equal(flows[j].BelongsDate) {
			return flows[i].BelongsDate.After(flows[j].BelongsDate)
		}
		if !flows[i].CreateTime.Equal(flows[j].CreateTime) {
			return flows[i].CreateTime.After(flows[j].CreateTime)
		}
		return flows[i].Id.Hex() > flows[j].Id.Hex()
	})
	return pageCashFlows(flows, filter.Limit, filter.Offset), nil
}

func (fake *cashFlowMapperFake) CountCashFlowsByFilter(filter model.CashFlowFilter) (int64, error) {
	limit, offset := filter.Limit, filter.Offset
	filter.Limit, filter.Offset = 0, 0
	flows, err := fake.GetCashFlowsByFilter(filter)
	filter.Limit, filter.Offset = limit, offset
	return int64(len(flows)), err
}

func (fake *cashFlowMapperFake) GetAllCashFlowsByUserIncludeDeleted(userId primitive.ObjectID) []model.CashFlowEntity {
	return fake.GetAllCashFlowsByUser(userId, 0, 0)
}

func (fake *cashFlowMapperFake) DeleteAllCashFlowsByUser(userId primitive.ObjectID) (int64, error) {
	var deleted int64
	for id, flow := range fake.flows {
		if flow.BelongsUserId == userId {
			delete(fake.flows, id)
			deleted++
		}
	}
	return deleted, nil
}

func (fake *cashFlowMapperFake) allFlows() []model.CashFlowEntity {
	flows := make([]model.CashFlowEntity, 0, len(fake.flows))
	for _, flow := range fake.flows {
		flows = append(flows, flow)
	}
	return flows
}

type cashFlowCategoryMapperFake struct {
	categories map[primitive.ObjectID]model.CategoryEntity
}

func newCashFlowCategoryMapperFake() *cashFlowCategoryMapperFake {
	return &cashFlowCategoryMapperFake{categories: map[primitive.ObjectID]model.CategoryEntity{}}
}

func (fake *cashFlowCategoryMapperFake) GetCategoryByObjectId(plainId string) model.CategoryEntity {
	id, _ := primitive.ObjectIDFromHex(plainId)
	return fake.categories[id]
}

func (fake *cashFlowCategoryMapperFake) GetCategoryByName(categoryName string) model.CategoryEntity {
	for _, category := range fake.categories {
		if category.Name == categoryName {
			return category
		}
	}
	return model.CategoryEntity{}
}

func (fake *cashFlowCategoryMapperFake) GetCategoryByParentId(parentPlainId string) []model.CategoryEntity {
	return nil
}

func (fake *cashFlowCategoryMapperFake) InsertCategoryByEntity(newEntity model.CategoryEntity) string {
	fake.categories[newEntity.Id] = newEntity
	return newEntity.Id.Hex()
}

func (fake *cashFlowCategoryMapperFake) UpdateCategoryByEntity(plainId string, updatedEntity model.CategoryEntity) model.CategoryEntity {
	return updatedEntity
}

func (fake *cashFlowCategoryMapperFake) GetAllCategories(limit, offset int) []model.CategoryEntity {
	return nil
}

func (fake *cashFlowCategoryMapperFake) GetAllCategoriesIncludeDeleted(limit, offset int) []model.CategoryEntity {
	return nil
}

func (fake *cashFlowCategoryMapperFake) CountAllCategories() int64 {
	return int64(len(fake.categories))
}

func (fake *cashFlowCategoryMapperFake) CountCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) (int64, error) {
	return int64(len(fake.typeCategories(userId, categoryType))), nil
}

func (fake *cashFlowCategoryMapperFake) DeleteCategoryByObjectId(plainId string) model.CategoryEntity {
	return model.CategoryEntity{}
}

func (fake *cashFlowCategoryMapperFake) TruncateCategories() error {
	return nil
}

func (fake *cashFlowCategoryMapperFake) GetCategoriesByUserAndType(userId primitive.ObjectID, categoryType string, limit, offset int) ([]model.CategoryEntity, error) {
	return fake.typeCategories(userId, categoryType), nil
}

func (fake *cashFlowCategoryMapperFake) GetRootCategoriesByUser(userId primitive.ObjectID) ([]model.CategoryEntity, error) {
	return nil, nil
}

func (fake *cashFlowCategoryMapperFake) GetRootCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) {
	return nil, nil
}

func (fake *cashFlowCategoryMapperFake) GetCategoriesByParentIdAndUser(parentId primitive.ObjectID, userId primitive.ObjectID) ([]model.CategoryEntity, error) {
	return nil, nil
}

func (fake *cashFlowCategoryMapperFake) GetCategoriesByParentIdUserAndType(parentId primitive.ObjectID, userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) {
	return nil, nil
}

func (fake *cashFlowCategoryMapperFake) GetCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity {
	category := fake.GetCategoryByObjectId(plainId)
	if category.BelongsUserId == userId {
		return category
	}
	return model.CategoryEntity{}
}

func (fake *cashFlowCategoryMapperFake) GetCategoryByNameAndUser(categoryName string, userId primitive.ObjectID) model.CategoryEntity {
	for _, category := range fake.categories {
		if category.Name == categoryName && category.BelongsUserId == userId {
			return category
		}
	}
	return model.CategoryEntity{}
}

func (fake *cashFlowCategoryMapperFake) GetCategoryByNameUserAndType(categoryName string, userId primitive.ObjectID, categoryType string) model.CategoryEntity {
	for _, category := range fake.categories {
		if category.Name == categoryName && category.BelongsUserId == userId && category.Type == categoryType {
			return category
		}
	}
	return model.CategoryEntity{}
}

func (fake *cashFlowCategoryMapperFake) GetCategoryByNameUserTypeAndParent(categoryName string, userId primitive.ObjectID, categoryType string, parentId primitive.ObjectID) model.CategoryEntity {
	return fake.GetCategoryByNameUserAndType(categoryName, userId, categoryType)
}

func (fake *cashFlowCategoryMapperFake) DeleteCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity {
	return model.CategoryEntity{}
}

func (fake *cashFlowCategoryMapperFake) UpdateCategoryByEntityAndUser(plainId string, updatedEntity model.CategoryEntity, userId primitive.ObjectID) model.CategoryEntity {
	return updatedEntity
}

func (fake *cashFlowCategoryMapperFake) GetAllCategoriesByUser(userId primitive.ObjectID, limit, offset int) []model.CategoryEntity {
	return nil
}

func (fake *cashFlowCategoryMapperFake) CountAllCategoriesByUser(userId primitive.ObjectID) int64 {
	return 0
}

func (fake *cashFlowCategoryMapperFake) GetAllCategoriesByUserIncludeDeleted(userId primitive.ObjectID) []model.CategoryEntity {
	return nil
}

func (fake *cashFlowCategoryMapperFake) DeleteAllCategoriesByUser(userId primitive.ObjectID) (int64, error) {
	return 0, nil
}

func (fake *cashFlowCategoryMapperFake) typeCategories(userId primitive.ObjectID, categoryType string) []model.CategoryEntity {
	var categories []model.CategoryEntity
	for _, category := range fake.categories {
		if category.BelongsUserId == userId && category.Type == categoryType {
			categories = append(categories, category)
		}
	}
	return categories
}

func sameDay(left, right time.Time) bool {
	ly, lm, ld := left.Date()
	ry, rm, rd := right.Date()
	return ly == ry && lm == rm && ld == rd
}

func pageCashFlows(flows []model.CashFlowEntity, limit, offset int) []model.CashFlowEntity {
	if offset > len(flows) {
		return []model.CashFlowEntity{}
	}
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		return flows[offset:]
	}
	end := offset + limit
	if end > len(flows) {
		end = len(flows)
	}
	return flows[offset:end]
}
