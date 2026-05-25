package user_controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserScopedHandlersRequireAuthenticatedUser(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
		body    string
	}{
		{"get profile", GetProfile, http.MethodGet, "/user/profile", ""},
		{"update profile", UpdateProfile, http.MethodPut, "/user/profile", `{"nickname":"Alice"}`},
		{"get configuration", GetConfiguration, http.MethodGet, "/user/configuration", ""},
		{"upsert configuration", UpsertConfiguration, http.MethodPost, "/user/configuration", `{"currency_code":"USD"}`},
		{"change password", ChangePassword, http.MethodPut, "/user/password", `{"old_password":"old","new_password":"new"}`},
		{"request email change", RequestEmailChange, http.MethodPost, "/user/email/change", `{"new_email":"new@example.com"}`},
		{"confirm email change", ConfirmEmailChange, http.MethodPost, "/user/email/confirm", `{"token":"token","password":"password"}`},
		{"delete account", DeleteAccount, http.MethodDelete, "/user/account", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
			}
		})
	}
}

func TestUserControllersRejectInvalidInputBeforeService(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
		body    string
		vars    map[string]string
	}{
		{"admin create invalid json", Create, http.MethodPost, "/admin/user", "{", nil},
		{"admin create missing username", Create, http.MethodPost, "/admin/user", `{"password":"password"}`, nil},
		{"admin get missing id", Get, http.MethodGet, "/admin/user/", "", nil},
		{"admin update missing id", Update, http.MethodPut, "/admin/user/", `{}`, nil},
		{"admin update invalid json", Update, http.MethodPut, "/admin/user/id", "{", map[string]string{"id": "id"}},
		{"admin delete missing id", Delete, http.MethodDelete, "/admin/user/", "", nil},
		{"profile update invalid json", UpdateProfile, http.MethodPut, "/user/profile", "{", nil},
		{"configuration invalid json", UpsertConfiguration, http.MethodPost, "/user/configuration", "{", nil},
		{"password invalid json", ChangePassword, http.MethodPut, "/user/password", "{", nil},
		{"password missing fields", ChangePassword, http.MethodPut, "/user/password", `{}`, nil},
		{"email change invalid json", RequestEmailChange, http.MethodPost, "/user/email/change", "{", nil},
		{"email change missing email", RequestEmailChange, http.MethodPost, "/user/email/change", `{}`, nil},
		{"email confirm invalid json", ConfirmEmailChange, http.MethodPost, "/user/email/confirm", "{", nil},
		{"email confirm missing token", ConfirmEmailChange, http.MethodPost, "/user/email/confirm", `{"password":"password"}`, nil},
		{"reset request invalid json", RequestPasswordReset, http.MethodPost, "/open/auth/reset-password", "{", nil},
		{"reset request missing user", RequestPasswordReset, http.MethodPost, "/open/auth/reset-password", `{}`, nil},
		{"reset confirm invalid json", ConfirmPasswordReset, http.MethodPost, "/open/auth/reset-password/confirm", "{", nil},
		{"reset confirm missing token", ConfirmPasswordReset, http.MethodPost, "/open/auth/reset-password/confirm", `{"password":"password"}`, nil},
		{"reset confirm missing password", ConfirmPasswordReset, http.MethodPost, "/open/auth/reset-password/confirm", `{"token":"token"}`, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req = req.WithContext(context.WithValue(req.Context(), "user_id", "admin-id"))
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func TestGetProfileReturnsLoggedInUser(t *testing.T) {
	userID := primitive.NewObjectID()
	installUserMapperStub(t, model.UserEntity{
		Id:       userID,
		Username: "alice",
		Nickname: "Alice",
		Role:     model.UserRoleUser,
		IsActive: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID.Hex()))
	rec := httptest.NewRecorder()

	GetProfile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestAdminGetReturnsNotFoundWhenUserMissing(t *testing.T) {
	installUserMapperStub(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/user/missing", nil)
	req = mux.SetURLVars(req, map[string]string{"id": primitive.NewObjectID().Hex()})
	rec := httptest.NewRecorder()

	Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestListAllUsesPaginationDefaultsAndReturnsUsers(t *testing.T) {
	installUserMapperStub(t, model.UserEntity{
		Id:       primitive.NewObjectID(),
		Username: "alice",
		Role:     model.UserRoleUser,
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/user?limit=not-number&offset=also-bad", nil)
	rec := httptest.NewRecorder()

	ListAll(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func installUserMapperStub(t *testing.T, users ...model.UserEntity) *userMapperStub {
	t.Helper()
	original := user_mapper.INSTANCE
	stub := &userMapperStub{users: map[string]model.UserEntity{}}
	for _, user := range users {
		stub.users[user.Id.Hex()] = user
	}
	user_mapper.INSTANCE = stub
	t.Cleanup(func() {
		user_mapper.INSTANCE = original
	})
	return stub
}

type userMapperStub struct {
	users map[string]model.UserEntity
}

func (stub *userMapperStub) GetUserByObjectId(plainId string) model.UserEntity {
	return stub.users[plainId]
}

func (stub *userMapperStub) GetUserByUsername(username string) model.UserEntity {
	for _, user := range stub.users {
		if user.Username == username {
			return user
		}
	}
	return model.UserEntity{}
}

func (stub *userMapperStub) GetUserByUsernameIncludeDeleted(username string) model.UserEntity {
	return stub.GetUserByUsername(username)
}

func (stub *userMapperStub) GetUserByEmail(email string) model.UserEntity {
	for _, user := range stub.users {
		if user.EmailAddress == email {
			return user
		}
	}
	return model.UserEntity{}
}

func (stub *userMapperStub) InsertUserByEntity(newEntity model.UserEntity) string {
	stub.users[newEntity.Id.Hex()] = newEntity
	return newEntity.Id.Hex()
}

func (stub *userMapperStub) UpdateUserByEntity(plainId string, updatedEntity model.UserEntity) model.UserEntity {
	stub.users[plainId] = updatedEntity
	return updatedEntity
}

func (stub *userMapperStub) GetAllUsers(limit, offset int) []model.UserEntity {
	users := make([]model.UserEntity, 0, len(stub.users))
	for _, user := range stub.users {
		users = append(users, user)
	}
	return users
}

func (stub *userMapperStub) GetAllUsersIncludeDeleted(limit, offset int) []model.UserEntity {
	return stub.GetAllUsers(limit, offset)
}

func (stub *userMapperStub) GetUsersByRole(role string) []model.UserEntity {
	users := []model.UserEntity{}
	for _, user := range stub.users {
		if user.Role == role {
			users = append(users, user)
		}
	}
	return users
}

func (stub *userMapperStub) CountAllUsers() int64 {
	return int64(len(stub.users))
}

func (stub *userMapperStub) DeleteUserByObjectId(plainId string) model.UserEntity {
	user := stub.users[plainId]
	delete(stub.users, plainId)
	return user
}

func (stub *userMapperStub) TruncateUsers() error {
	stub.users = map[string]model.UserEntity{}
	return nil
}
