package user_controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildUserProfileResponseIncludesNicknameAndInactiveStatus(t *testing.T) {
	userID := primitive.NewObjectID()
	response := buildUserProfileResponse(model.UserEntity{
		Id:          userID,
		Username:    "alice",
		Nickname:    "Alice Chen",
		Gender:      model.GenderFemale,
		PhoneNumber: "+65 6123 4567",
		Location:    "Singapore",
		BirthDate:   "1992-08-21",
		IsActive:    false,
	})

	if response.Id != userID.Hex() {
		t.Fatalf("Id = %q, want %q", response.Id, userID.Hex())
	}
	if response.Nickname != "Alice Chen" {
		t.Fatalf("Nickname = %q, want Alice Chen", response.Nickname)
	}
	if response.Gender != model.GenderFemale {
		t.Fatalf("Gender = %q, want %q", response.Gender, model.GenderFemale)
	}
	if response.PhoneNumber != "+65 6123 4567" || response.Location != "Singapore" || response.BirthDate != "1992-08-21" {
		t.Fatalf("extended profile response = %+v", response)
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if !json.Valid(payload) {
		t.Fatal("expected valid JSON")
	}
	if !strings.Contains(string(payload), `"is_active":false`) {
		t.Fatalf("expected inactive status to be emitted, got %s", payload)
	}
}

func TestGetProfilePassesAuthenticatedUserToService(t *testing.T) {
	userID := primitive.NewObjectID()
	original := getProfileUser
	getProfileUser = func(serviceUserID string) model.UserEntity {
		if serviceUserID != userID.Hex() {
			t.Fatalf("service user id = %q, want %q", serviceUserID, userID.Hex())
		}
		return model.UserEntity{
			Id:       userID,
			Username: "alice",
			Nickname: "Alice",
			IsActive: true,
			Role:     model.UserRoleUser,
		}
	}
	t.Cleanup(func() { getProfileUser = original })

	req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID.Hex()))
	rec := httptest.NewRecorder()

	GetProfile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var response struct {
		Data model.UserProfileResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("response JSON decode failed: %v", err)
	}
	if response.Data.Id != userID.Hex() || response.Data.Nickname != "Alice" {
		t.Fatalf("profile response = %+v", response.Data)
	}
}

func TestUpdateProfilePassesRequestToService(t *testing.T) {
	userID := primitive.NewObjectID()
	var got model.UserProfileUpdateRequest
	original := updateProfileUser
	updateProfileUser = func(serviceUserID string, request model.UserProfileUpdateRequest) (model.UserEntity, error) {
		if serviceUserID != userID.Hex() {
			t.Fatalf("service user id = %q, want %q", serviceUserID, userID.Hex())
		}
		got = request
		return model.UserEntity{
			Id:          userID,
			Username:    "alice",
			Nickname:    request.Nickname,
			AvatarUrl:   request.AvatarUrl,
			Gender:      request.Gender,
			PhoneNumber: request.PhoneNumber,
			Location:    request.Location,
			BirthDate:   request.BirthDate,
			IsActive:    true,
		}, nil
	}
	t.Cleanup(func() { updateProfileUser = original })

	req := httptest.NewRequest(http.MethodPut, "/user/profile", strings.NewReader(`{"nickname":"Alice","avatar_url":"https://example.test/a.png","gender":"female","phone_number":"+65 6123 4567","location":"Singapore","birth_date":"1992-08-21"}`))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID.Hex()))
	rec := httptest.NewRecorder()

	UpdateProfile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got.Nickname != "Alice" || got.AvatarUrl != "https://example.test/a.png" || got.Gender != model.GenderFemale {
		t.Fatalf("service request = %+v", got)
	}
	if got.PhoneNumber != "+65 6123 4567" || got.Location != "Singapore" || got.BirthDate != "1992-08-21" {
		t.Fatalf("extended service request = %+v", got)
	}
}
