package user_controller

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildUserProfileResponseIncludesNicknameAndInactiveStatus(t *testing.T) {
	userID := primitive.NewObjectID()
	response := buildUserProfileResponse(model.UserEntity{
		Id:       userID,
		Username: "alice",
		Nickname: "Alice Chen",
		Gender:   model.GenderFemale,
		IsActive: false,
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
