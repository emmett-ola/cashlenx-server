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

func TestPasswordResetHandlersDelegateSuccessfulProcedure(t *testing.T) {
	var requestedIdentity, requestedIP string
	var confirmedToken, confirmedPassword string
	originalRequest := requestUserPasswordReset
	originalConfirm := confirmUserPasswordReset
	requestUserPasswordReset = func(identity, ipAddress string) error {
		requestedIdentity, requestedIP = identity, ipAddress
		return nil
	}
	confirmUserPasswordReset = func(token, password string) error {
		confirmedToken, confirmedPassword = token, password
		return nil
	}
	t.Cleanup(func() {
		requestUserPasswordReset = originalRequest
		confirmUserPasswordReset = originalConfirm
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/reset-password", strings.NewReader(`{"email_or_username":"alice@example.test"}`))
	request.RemoteAddr = "192.0.2.10:1234"
	requestRecorder := httptest.NewRecorder()
	RequestPasswordReset(requestRecorder, request)
	if requestRecorder.Code != http.StatusOK {
		t.Fatalf("request status = %d, want %d; body=%s", requestRecorder.Code, http.StatusOK, requestRecorder.Body.String())
	}

	confirm := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/reset-password/confirm", strings.NewReader(`{"token":"reset-token","password":"NewPass123!"}`))
	confirmRecorder := httptest.NewRecorder()
	ConfirmPasswordReset(confirmRecorder, confirm)
	if confirmRecorder.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, want %d; body=%s", confirmRecorder.Code, http.StatusOK, confirmRecorder.Body.String())
	}

	if requestedIdentity != "alice@example.test" || requestedIP != "192.0.2.10" {
		t.Fatalf("reset request args = %q/%q", requestedIdentity, requestedIP)
	}
	if confirmedToken != "reset-token" || confirmedPassword != "NewPass123!" {
		t.Fatalf("reset confirm args = %q/%q", confirmedToken, confirmedPassword)
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

func TestUserControllerServiceDelegationPaths(t *testing.T) {
	adminID := primitive.NewObjectID().Hex()
	userID := primitive.NewObjectID()
	var createReq, updateReq model.UserDTO
	var createActor *string
	var getCreatedID, updateID, deleteAdminID string
	var configGetID, configUpsertID string
	var passwordArgs, deleteCurrentArgs struct {
		userID string
		first  string
		second string
	}
	var configReq model.UserConfigurationRequest

	originalCreate := createUserByAdmin
	createUserByAdmin = func(req model.UserDTO, actor *string) (string, error) {
		createReq = req
		createActor = actor
		return userID.Hex(), nil
	}
	originalGet := getUserByID
	getUserByID = func(id string) model.UserEntity {
		getCreatedID = id
		return model.UserEntity{Id: userID, Username: "alice", Role: model.UserRoleUser, IsActive: true}
	}
	originalUpdate := updateUserByAdmin
	updateUserByAdmin = func(id string, req model.UserDTO) (model.UserEntity, error) {
		updateID = id
		updateReq = req
		return model.UserEntity{Id: userID, Username: req.Username, Role: model.UserRoleUser, IsActive: true}, nil
	}
	originalDeleteAdmin := deleteUserByAdmin
	deleteUserByAdmin = func(id string) error {
		deleteAdminID = id
		return nil
	}
	originalGetConfig := getUserConfiguration
	getUserConfiguration = func(serviceUserID string) (model.UserConfigurationEntity, error) {
		configGetID = serviceUserID
		return testUserControllerConfig(userID), nil
	}
	originalUpsertConfig := upsertUserConfiguration
	upsertUserConfiguration = func(serviceUserID string, req model.UserConfigurationRequest) (model.UserConfigurationEntity, error) {
		configUpsertID = serviceUserID
		configReq = req
		return testUserControllerConfig(userID), nil
	}
	originalPassword := changeUserPassword
	changeUserPassword = func(serviceUserID, oldPassword, newPassword string) error {
		passwordArgs.userID = serviceUserID
		passwordArgs.first = oldPassword
		passwordArgs.second = newPassword
		return nil
	}
	originalDeleteCurrent := deleteCurrentUser
	deleteCurrentUser = func(serviceUserID string) error {
		deleteCurrentArgs.userID = serviceUserID
		return nil
	}
	t.Cleanup(func() {
		createUserByAdmin = originalCreate
		getUserByID = originalGet
		updateUserByAdmin = originalUpdate
		deleteUserByAdmin = originalDeleteAdmin
		getUserConfiguration = originalGetConfig
		upsertUserConfiguration = originalUpsertConfig
		changeUserPassword = originalPassword
		deleteCurrentUser = originalDeleteCurrent
	})

	createReqHTTP := httptest.NewRequest(http.MethodPost, "/admin/user", strings.NewReader(`{"username":"alice","password":"secret","email_address":"alice@example.test"}`))
	createReqHTTP = createReqHTTP.WithContext(context.WithValue(createReqHTTP.Context(), "user_id", adminID))
	createRec := httptest.NewRecorder()
	Create(createRec, createReqHTTP)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body=%s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}

	updateReqHTTP := httptest.NewRequest(http.MethodPut, "/admin/user/"+userID.Hex(), strings.NewReader(`{"username":"alice2","email_address":"alice2@example.test"}`))
	updateReqHTTP = mux.SetURLVars(updateReqHTTP, map[string]string{"id": userID.Hex()})
	updateRec := httptest.NewRecorder()
	Update(updateRec, updateReqHTTP)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d; body=%s", updateRec.Code, http.StatusOK, updateRec.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/admin/user/"+userID.Hex(), nil)
	deleteReq = mux.SetURLVars(deleteReq, map[string]string{"id": userID.Hex()})
	deleteRec := httptest.NewRecorder()
	Delete(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want %d; body=%s", deleteRec.Code, http.StatusOK, deleteRec.Body.String())
	}

	configGetReq := httptest.NewRequest(http.MethodGet, "/user/configuration", nil)
	configGetReq = configGetReq.WithContext(context.WithValue(configGetReq.Context(), "user_id", userID.Hex()))
	configGetRec := httptest.NewRecorder()
	GetConfiguration(configGetRec, configGetReq)
	if configGetRec.Code != http.StatusOK {
		t.Fatalf("config get status = %d, want %d; body=%s", configGetRec.Code, http.StatusOK, configGetRec.Body.String())
	}

	configUpsertReq := httptest.NewRequest(http.MethodPost, "/user/configuration", strings.NewReader(`{"display_language":"en-US","currency_code":"USD","active_theme_color":"blue"}`))
	configUpsertReq = configUpsertReq.WithContext(context.WithValue(configUpsertReq.Context(), "user_id", userID.Hex()))
	configUpsertRec := httptest.NewRecorder()
	UpsertConfiguration(configUpsertRec, configUpsertReq)
	if configUpsertRec.Code != http.StatusOK {
		t.Fatalf("config upsert status = %d, want %d; body=%s", configUpsertRec.Code, http.StatusOK, configUpsertRec.Body.String())
	}

	passwordReq := httptest.NewRequest(http.MethodPut, "/user/password", strings.NewReader(`{"old_password":"old","new_password":"new"}`))
	passwordReq = passwordReq.WithContext(context.WithValue(passwordReq.Context(), "user_id", userID.Hex()))
	passwordRec := httptest.NewRecorder()
	ChangePassword(passwordRec, passwordReq)
	if passwordRec.Code != http.StatusOK {
		t.Fatalf("password status = %d, want %d; body=%s", passwordRec.Code, http.StatusOK, passwordRec.Body.String())
	}

	accountReq := httptest.NewRequest(http.MethodDelete, "/user/account", nil)
	accountReq = accountReq.WithContext(context.WithValue(accountReq.Context(), "user_id", userID.Hex()))
	accountRec := httptest.NewRecorder()
	DeleteAccount(accountRec, accountReq)
	if accountRec.Code != http.StatusOK {
		t.Fatalf("account status = %d, want %d; body=%s", accountRec.Code, http.StatusOK, accountRec.Body.String())
	}

	if createReq.Username != "alice" || createReq.Password != "secret" || createReq.EmailAddress != "alice@example.test" || createActor == nil || *createActor != adminID || getCreatedID != userID.Hex() {
		t.Fatalf("create req=%+v actor=%v getCreatedID=%q", createReq, createActor, getCreatedID)
	}
	if updateID != userID.Hex() || updateReq.Username != "alice2" || updateReq.EmailAddress != "alice2@example.test" {
		t.Fatalf("update id=%q req=%+v", updateID, updateReq)
	}
	if deleteAdminID != userID.Hex() {
		t.Fatalf("delete admin id = %q", deleteAdminID)
	}
	if configGetID != userID.Hex() || configUpsertID != userID.Hex() {
		t.Fatalf("config ids = get %q upsert %q", configGetID, configUpsertID)
	}
	if configReq.DisplayLanguage == nil || *configReq.DisplayLanguage != "en-US" || configReq.CurrencyCode == nil || *configReq.CurrencyCode != "USD" || configReq.ActiveThemeColor == nil || *configReq.ActiveThemeColor != "blue" {
		t.Fatalf("config req = %+v", configReq)
	}
	if passwordArgs.userID != userID.Hex() || passwordArgs.first != "old" || passwordArgs.second != "new" {
		t.Fatalf("password args = %+v", passwordArgs)
	}
	if deleteCurrentArgs.userID != userID.Hex() {
		t.Fatalf("delete current args = %+v", deleteCurrentArgs)
	}
}

func testUserControllerConfig(userID primitive.ObjectID) model.UserConfigurationEntity {
	return model.UserConfigurationEntity{
		Id:               primitive.NewObjectID(),
		BelongsUserId:    userID,
		DisplayLanguage:  "en-US",
		CurrencyCode:     "USD",
		ActiveThemeColor: "blue",
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
