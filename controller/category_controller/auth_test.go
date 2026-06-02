package category_controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCategoryHandlersRequireAuthenticatedUser(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
		body    string
		vars    map[string]string
	}{
		{"create", Create, http.MethodPost, "/category", `{"name":"Food","type":"expense"}`, nil},
		{"list", ListAll, http.MethodGet, "/category", "", nil},
		{"tree", Tree, http.MethodGet, "/category/tree", "", nil},
		{"query by id", QueryById, http.MethodGet, "/category/id", "", map[string]string{"id": "id"}},
		{"query by name", QueryByName, http.MethodGet, "/category/name/Food", "", map[string]string{"name": "Food"}},
		{"query children", QueryChildren, http.MethodGet, "/category/id/children", "", map[string]string{"parent_id": "id"}},
		{"update", UpdateById, http.MethodPut, "/category/id", "", map[string]string{"id": "id"}},
		{"delete", DeleteById, http.MethodDelete, "/category/id", "", map[string]string{"id": "id"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
			}
		})
	}
}

func TestTreeRejectsInvalidCategoryTypeBeforeService(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/category/tree?type=invalid", nil)
	req = req.WithContext(contextWithUserID(req.Context(), "user-id"))
	rec := httptest.NewRecorder()

	Tree(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCategoryCreateRejectsInvalidJSONBeforeAuth(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/category", strings.NewReader(`{"name":`))
	rec := httptest.NewRecorder()

	Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCategoryUpdateRejectsMissingIDAndInvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
		vars map[string]string
	}{
		{name: "missing id", body: `{}`, vars: nil},
		{name: "invalid json", body: `{"name":`, vars: map[string]string{"id": "category-id"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/category/id", strings.NewReader(tc.body))
			req = req.WithContext(contextWithUserID(req.Context(), "user-id"))
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			rec := httptest.NewRecorder()

			UpdateById(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func TestCategoryCreatePassesRequestAndUserToService(t *testing.T) {
	userID := primitive.NewObjectID().Hex()
	parentID := primitive.NewObjectID().Hex()
	var got struct {
		name     string
		typ      string
		remark   string
		parentID string
		userID   string
	}
	original := createCategoryForUser
	createCategoryForUser = func(name, categoryType, remark, serviceParentID, serviceUserID string) (model.CategoryEntity, error) {
		got.name = name
		got.typ = categoryType
		got.remark = remark
		got.parentID = serviceParentID
		got.userID = serviceUserID
		return model.CategoryEntity{Id: primitive.NewObjectID(), Name: name, Type: categoryType, Remark: remark}, nil
	}
	t.Cleanup(func() { createCategoryForUser = original })

	req := httptest.NewRequest(http.MethodPost, "/category", strings.NewReader(`{"name":"Food","type":"expense","remark":"daily","parent_id":"`+parentID+`"}`))
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()

	Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if got.name != "Food" || got.typ != "expense" || got.remark != "daily" || got.parentID != parentID || got.userID != userID {
		t.Fatalf("service args = %+v", got)
	}
}

func TestCategoryListPassesQueryAndUserToService(t *testing.T) {
	userID := primitive.NewObjectID().Hex()
	var got struct {
		userID string
		typ    string
		limit  int
		offset int
	}
	original := queryCategoriesForUser
	queryCategoriesForUser = func(serviceUserID, categoryType string, limit, offset int) ([]model.CategoryEntity, int64, error) {
		got.userID = serviceUserID
		got.typ = categoryType
		got.limit = limit
		got.offset = offset
		return []model.CategoryEntity{{Id: primitive.NewObjectID(), Name: "Food", Type: categoryType}}, 3, nil
	}
	t.Cleanup(func() { queryCategoriesForUser = original })

	req := httptest.NewRequest(http.MethodGet, "/category?type=expense&limit=7&offset=14", nil)
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()

	ListAll(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got.userID != userID || got.typ != "expense" || got.limit != 7 || got.offset != 14 {
		t.Fatalf("service args = %+v", got)
	}
	var response struct {
		Meta map[string]interface{} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("response JSON decode failed: %v", err)
	}
	if response.Meta["total_count"].(float64) != 3 {
		t.Fatalf("meta = %+v", response.Meta)
	}
}

func TestCategoryQueryHandlersPassRouteValuesAndUserToService(t *testing.T) {
	userID := primitive.NewObjectID().Hex()
	categoryID := primitive.NewObjectID().Hex()
	parentID := primitive.NewObjectID().Hex()

	originalByID := queryCategoryByIDForUser
	queryCategoryByIDForUser = func(serviceCategoryID, serviceUserID string) (model.CategoryEntity, error) {
		if serviceCategoryID != categoryID || serviceUserID != userID {
			t.Fatalf("by id args = %q, %q", serviceCategoryID, serviceUserID)
		}
		return model.CategoryEntity{Id: primitive.NewObjectID(), Name: "Food"}, nil
	}
	originalByName := queryCategoryByNameForUser
	queryCategoryByNameForUser = func(name, serviceUserID string) (model.CategoryEntity, error) {
		if name != "Food" || serviceUserID != userID {
			t.Fatalf("by name args = %q, %q", name, serviceUserID)
		}
		return model.CategoryEntity{Id: primitive.NewObjectID(), Name: name}, nil
	}
	originalChildren := queryChildCategoriesForUser
	queryChildCategoriesForUser = func(serviceParentID, serviceUserID, categoryType string) ([]model.CategoryEntity, error) {
		if serviceParentID != parentID || serviceUserID != userID || categoryType != "income" {
			t.Fatalf("children args = %q, %q, %q", serviceParentID, serviceUserID, categoryType)
		}
		return []model.CategoryEntity{{Id: primitive.NewObjectID(), Name: "Salary", Type: categoryType}}, nil
	}
	t.Cleanup(func() {
		queryCategoryByIDForUser = originalByID
		queryCategoryByNameForUser = originalByName
		queryChildCategoriesForUser = originalChildren
	})

	cases := []struct {
		name    string
		handler http.HandlerFunc
		path    string
		vars    map[string]string
	}{
		{"by id", QueryById, "/category/" + categoryID, map[string]string{"id": categoryID}},
		{"by name", QueryByName, "/category/name/Food", map[string]string{"name": "Food"}},
		{"children", QueryChildren, "/category/" + parentID + "/children?type=income", map[string]string{"parent_id": parentID}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req = req.WithContext(contextWithUserID(req.Context(), userID))
			req = mux.SetURLVars(req, tc.vars)
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
			}
		})
	}
}

func TestCategoryUpdateAndDeletePassRouteBodyQueryAndUserToService(t *testing.T) {
	userID := primitive.NewObjectID().Hex()
	categoryID := primitive.NewObjectID().Hex()
	parentID := primitive.NewObjectID().Hex()

	originalUpdate := updateCategoryForUser
	updateCategoryForUser = func(serviceCategoryID, name, categoryType, remark, serviceParentID, serviceUserID string) (model.CategoryEntity, error) {
		if serviceCategoryID != categoryID || name != "Meals" || categoryType != "expense" || remark != "updated" || serviceParentID != parentID || serviceUserID != userID {
			t.Fatalf("update args = %q, %q, %q, %q, %q, %q", serviceCategoryID, name, categoryType, remark, serviceParentID, serviceUserID)
		}
		return model.CategoryEntity{Id: primitive.NewObjectID(), Name: name, Type: categoryType, Remark: remark}, nil
	}
	originalDelete := deleteCategoryForUser
	deleteCategoryForUser = func(serviceCategoryID, serviceUserID string, force bool) (model.CategoryEntity, error) {
		if serviceCategoryID != categoryID || serviceUserID != userID || !force {
			t.Fatalf("delete args = %q, %q, %t", serviceCategoryID, serviceUserID, force)
		}
		return model.CategoryEntity{Id: primitive.NewObjectID(), Name: "Meals"}, nil
	}
	t.Cleanup(func() {
		updateCategoryForUser = originalUpdate
		deleteCategoryForUser = originalDelete
	})

	updateReq := httptest.NewRequest(http.MethodPut, "/category/"+categoryID, strings.NewReader(`{"name":"Meals","type":"expense","remark":"updated","parent_id":"`+parentID+`"}`))
	updateReq = updateReq.WithContext(contextWithUserID(updateReq.Context(), userID))
	updateReq = mux.SetURLVars(updateReq, map[string]string{"id": categoryID})
	updateRec := httptest.NewRecorder()
	UpdateById(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d; body=%s", updateRec.Code, http.StatusOK, updateRec.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/category/"+categoryID+"?force=true", nil)
	deleteReq = deleteReq.WithContext(contextWithUserID(deleteReq.Context(), userID))
	deleteReq = mux.SetURLVars(deleteReq, map[string]string{"id": categoryID})
	deleteRec := httptest.NewRecorder()
	DeleteById(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want %d; body=%s", deleteRec.Code, http.StatusOK, deleteRec.Body.String())
	}
}

func TestCategoryTreePassesUserAndTypeToService(t *testing.T) {
	userID := primitive.NewObjectID().Hex()
	original := queryCategoryTreeForUser
	queryCategoryTreeForUser = func(serviceUserID, categoryType string) ([]model.CategoryTree, error) {
		if serviceUserID != userID || categoryType != "income" {
			t.Fatalf("tree args = %q, %q", serviceUserID, categoryType)
		}
		return []model.CategoryTree{{Id: primitive.NewObjectID().Hex(), Name: "Salary", Type: categoryType}}, nil
	}
	t.Cleanup(func() { queryCategoryTreeForUser = original })

	req := httptest.NewRequest(http.MethodGet, "/category/tree?type=income", nil)
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()

	Tree(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func contextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, "user_id", userID)
}
