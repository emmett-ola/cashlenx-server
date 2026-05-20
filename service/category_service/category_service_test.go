package category_service

import (
	"testing"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateForUserCreatesRootCategory(t *testing.T) {
	service, stub := newCategoryServiceStub()
	userID := primitive.NewObjectID()

	created, err := service.CreateForUser("Salary", "INCOME", "monthly pay", "", userID.Hex())
	if err != nil {
		t.Fatalf("CreateForUser returned error: %v", err)
	}

	if created.Id.IsZero() {
		t.Fatal("expected created category to have an id")
	}
	if created.BelongsUserId != userID {
		t.Fatalf("BelongsUserId = %s, want %s", created.BelongsUserId.Hex(), userID.Hex())
	}
	if created.Type != "income" {
		t.Fatalf("Type = %q, want income", created.Type)
	}
	if !created.ParentId.IsZero() {
		t.Fatalf("ParentId = %s, want zero ObjectID", created.ParentId.Hex())
	}
	if len(stub.inserted) != 1 {
		t.Fatalf("inserted count = %d, want 1", len(stub.inserted))
	}
}

func TestCreateForUserRejectsDuplicateUnderSameParent(t *testing.T) {
	service, stub := newCategoryServiceStub()
	userID := primitive.NewObjectID()
	parentID := primitive.NewObjectID()
	stub.categories[parentID] = model.CategoryEntity{
		Id:            parentID,
		BelongsUserId: userID,
		Name:          "Food",
		Type:          "expense",
	}
	stub.categories[primitive.NewObjectID()] = model.CategoryEntity{
		Id:            primitive.NewObjectID(),
		BelongsUserId: userID,
		ParentId:      parentID,
		Name:          "Lunch",
		Type:          "expense",
	}

	_, err := service.CreateForUser("Lunch", "expense", "", parentID.Hex(), userID.Hex())
	if err == nil {
		t.Fatal("expected duplicate category error")
	}
	if len(stub.inserted) != 0 {
		t.Fatalf("inserted count = %d, want 0", len(stub.inserted))
	}
}

func TestCreateForUserRejectsParentTypeMismatch(t *testing.T) {
	service, stub := newCategoryServiceStub()
	userID := primitive.NewObjectID()
	parentID := primitive.NewObjectID()
	stub.categories[parentID] = model.CategoryEntity{
		Id:            parentID,
		BelongsUserId: userID,
		Name:          "Pay",
		Type:          "income",
	}

	_, err := service.CreateForUser("Lunch", "expense", "", parentID.Hex(), userID.Hex())
	if err == nil {
		t.Fatal("expected parent type mismatch error")
	}
	if len(stub.inserted) != 0 {
		t.Fatalf("inserted count = %d, want 0", len(stub.inserted))
	}
}

func TestGetCategoryTreeByUserBuildsFilteredTree(t *testing.T) {
	service, stub := newCategoryServiceStub()
	userID := primitive.NewObjectID()
	rootID := primitive.NewObjectID()
	childID := primitive.NewObjectID()
	otherTypeID := primitive.NewObjectID()
	stub.categories[rootID] = model.CategoryEntity{
		Id:            rootID,
		BelongsUserId: userID,
		Name:          "Food",
		Type:          "expense",
	}
	stub.categories[childID] = model.CategoryEntity{
		Id:            childID,
		BelongsUserId: userID,
		ParentId:      rootID,
		Name:          "Lunch",
		Type:          "expense",
	}
	stub.categories[otherTypeID] = model.CategoryEntity{
		Id:            otherTypeID,
		BelongsUserId: userID,
		Name:          "Salary",
		Type:          "income",
	}

	tree, err := service.GetCategoryTreeByUser(userID.Hex(), "expense")
	if err != nil {
		t.Fatalf("GetCategoryTreeByUser returned error: %v", err)
	}
	if len(tree) != 1 {
		t.Fatalf("tree length = %d, want 1", len(tree))
	}
	if tree[0].Name != "Food" {
		t.Fatalf("root name = %q, want Food", tree[0].Name)
	}
	if len(tree[0].Children) != 1 {
		t.Fatalf("child count = %d, want 1", len(tree[0].Children))
	}
	if tree[0].Children[0].Name != "Lunch" {
		t.Fatalf("child name = %q, want Lunch", tree[0].Children[0].Name)
	}
}

func newCategoryServiceStub() (*CategoryService, *categoryMapperStub) {
	stub := &categoryMapperStub{categories: map[primitive.ObjectID]model.CategoryEntity{}}
	return NewCategoryService(stub, nil), stub
}

type categoryMapperStub struct {
	categories map[primitive.ObjectID]model.CategoryEntity
	inserted   []model.CategoryEntity
}

func (stub *categoryMapperStub) GetCategoryByObjectId(plainId string) model.CategoryEntity {
	id, _ := primitive.ObjectIDFromHex(plainId)
	return stub.categories[id]
}

func (stub *categoryMapperStub) GetCategoryByName(categoryName string) model.CategoryEntity {
	for _, category := range stub.categories {
		if category.Name == categoryName {
			return category
		}
	}
	return model.CategoryEntity{}
}

func (stub *categoryMapperStub) GetCategoryByParentId(parentPlainId string) []model.CategoryEntity {
	parentID, _ := primitive.ObjectIDFromHex(parentPlainId)
	return stub.children(parentID, primitive.NilObjectID, "")
}

func (stub *categoryMapperStub) InsertCategoryByEntity(newEntity model.CategoryEntity) string {
	stub.categories[newEntity.Id] = newEntity
	stub.inserted = append(stub.inserted, newEntity)
	return newEntity.Id.Hex()
}

func (stub *categoryMapperStub) UpdateCategoryByEntity(plainId string, updatedEntity model.CategoryEntity) model.CategoryEntity {
	id, _ := primitive.ObjectIDFromHex(plainId)
	updatedEntity.Id = id
	stub.categories[id] = updatedEntity
	return updatedEntity
}

func (stub *categoryMapperStub) GetAllCategories(limit, offset int) []model.CategoryEntity {
	return stub.all(primitive.NilObjectID, "", limit, offset)
}

func (stub *categoryMapperStub) GetAllCategoriesIncludeDeleted(limit, offset int) []model.CategoryEntity {
	return stub.GetAllCategories(limit, offset)
}

func (stub *categoryMapperStub) CountAllCategories() int64 {
	return int64(len(stub.categories))
}

func (stub *categoryMapperStub) CountCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) (int64, error) {
	return int64(len(stub.all(userId, categoryType, 0, 0))), nil
}

func (stub *categoryMapperStub) DeleteCategoryByObjectId(plainId string) model.CategoryEntity {
	id, _ := primitive.ObjectIDFromHex(plainId)
	category := stub.categories[id]
	delete(stub.categories, id)
	return category
}

func (stub *categoryMapperStub) TruncateCategories() error {
	stub.categories = map[primitive.ObjectID]model.CategoryEntity{}
	return nil
}

func (stub *categoryMapperStub) GetCategoriesByUserAndType(userId primitive.ObjectID, categoryType string, limit, offset int) ([]model.CategoryEntity, error) {
	return stub.all(userId, categoryType, limit, offset), nil
}

func (stub *categoryMapperStub) GetRootCategoriesByUser(userId primitive.ObjectID) ([]model.CategoryEntity, error) {
	return stub.roots(userId, ""), nil
}

func (stub *categoryMapperStub) GetRootCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) {
	return stub.roots(userId, categoryType), nil
}

func (stub *categoryMapperStub) GetCategoriesByParentIdAndUser(parentId primitive.ObjectID, userId primitive.ObjectID) ([]model.CategoryEntity, error) {
	return stub.children(parentId, userId, ""), nil
}

func (stub *categoryMapperStub) GetCategoriesByParentIdUserAndType(parentId primitive.ObjectID, userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) {
	return stub.children(parentId, userId, categoryType), nil
}

func (stub *categoryMapperStub) GetCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity {
	category := stub.GetCategoryByObjectId(plainId)
	if category.BelongsUserId == userId {
		return category
	}
	return model.CategoryEntity{}
}

func (stub *categoryMapperStub) GetCategoryByNameAndUser(categoryName string, userId primitive.ObjectID) model.CategoryEntity {
	for _, category := range stub.categories {
		if category.Name == categoryName && category.BelongsUserId == userId {
			return category
		}
	}
	return model.CategoryEntity{}
}

func (stub *categoryMapperStub) GetCategoryByNameUserAndType(categoryName string, userId primitive.ObjectID, categoryType string) model.CategoryEntity {
	for _, category := range stub.categories {
		if category.Name == categoryName && category.BelongsUserId == userId && category.Type == categoryType {
			return category
		}
	}
	return model.CategoryEntity{}
}

func (stub *categoryMapperStub) GetCategoryByNameUserTypeAndParent(categoryName string, userId primitive.ObjectID, categoryType string, parentId primitive.ObjectID) model.CategoryEntity {
	for _, category := range stub.categories {
		if category.Name == categoryName && category.BelongsUserId == userId && category.Type == categoryType && category.ParentId == parentId {
			return category
		}
	}
	return model.CategoryEntity{}
}

func (stub *categoryMapperStub) DeleteCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity {
	category := stub.GetCategoryByObjectIdAndUser(plainId, userId)
	if !category.IsEmpty() {
		delete(stub.categories, category.Id)
	}
	return category
}

func (stub *categoryMapperStub) UpdateCategoryByEntityAndUser(plainId string, updatedEntity model.CategoryEntity, userId primitive.ObjectID) model.CategoryEntity {
	id, _ := primitive.ObjectIDFromHex(plainId)
	existing := stub.categories[id]
	if existing.BelongsUserId != userId {
		return model.CategoryEntity{}
	}
	updatedEntity.Id = id
	updatedEntity.BelongsUserId = userId
	stub.categories[id] = updatedEntity
	return updatedEntity
}

func (stub *categoryMapperStub) GetAllCategoriesByUser(userId primitive.ObjectID, limit, offset int) []model.CategoryEntity {
	return stub.all(userId, "", limit, offset)
}

func (stub *categoryMapperStub) CountAllCategoriesByUser(userId primitive.ObjectID) int64 {
	return int64(len(stub.all(userId, "", 0, 0)))
}

func (stub *categoryMapperStub) GetAllCategoriesByUserIncludeDeleted(userId primitive.ObjectID) []model.CategoryEntity {
	return stub.all(userId, "", 0, 0)
}

func (stub *categoryMapperStub) DeleteAllCategoriesByUser(userId primitive.ObjectID) (int64, error) {
	var deleted int64
	for id, category := range stub.categories {
		if category.BelongsUserId == userId {
			delete(stub.categories, id)
			deleted++
		}
	}
	return deleted, nil
}

func (stub *categoryMapperStub) roots(userId primitive.ObjectID, categoryType string) []model.CategoryEntity {
	var categories []model.CategoryEntity
	for _, category := range stub.categories {
		if category.BelongsUserId != userId || !category.ParentId.IsZero() {
			continue
		}
		if categoryType != "" && category.Type != categoryType {
			continue
		}
		categories = append(categories, category)
	}
	return categories
}

func (stub *categoryMapperStub) children(parentId, userId primitive.ObjectID, categoryType string) []model.CategoryEntity {
	var categories []model.CategoryEntity
	for _, category := range stub.categories {
		if category.ParentId != parentId {
			continue
		}
		if !userId.IsZero() && category.BelongsUserId != userId {
			continue
		}
		if categoryType != "" && category.Type != categoryType {
			continue
		}
		categories = append(categories, category)
	}
	return categories
}

func (stub *categoryMapperStub) all(userId primitive.ObjectID, categoryType string, limit, offset int) []model.CategoryEntity {
	var categories []model.CategoryEntity
	for _, category := range stub.categories {
		if !userId.IsZero() && category.BelongsUserId != userId {
			continue
		}
		if categoryType != "" && category.Type != categoryType {
			continue
		}
		categories = append(categories, category)
	}
	if offset > len(categories) {
		return []model.CategoryEntity{}
	}
	if limit <= 0 {
		return categories[offset:]
	}
	end := offset + limit
	if end > len(categories) {
		end = len(categories)
	}
	return categories[offset:end]
}
