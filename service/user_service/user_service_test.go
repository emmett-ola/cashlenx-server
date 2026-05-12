package user_service

import (
	"testing"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateServiceCreatesUserAndInitializesCategories(t *testing.T) {
	repo := installUserServiceTestDeps(t)
	var initializedUserID string
	initializeDefaultCategoriesForUser = func(userId string) error {
		initializedUserID = userId
		return nil
	}

	userID, err := CreateService(model.UserDTO{
		Username: "alice",
		Password: "StrongPass123!",
		Nickname: "Alice",
		Gender:   model.GenderFemale,
	}, nil)
	if err != nil {
		t.Fatalf("CreateService returned error: %v", err)
	}

	created := repo.users[userID]
	if created.Id.IsZero() {
		t.Fatal("expected user to be inserted")
	}
	if created.Role != model.UserRoleUser {
		t.Fatalf("Role = %q, want %q", created.Role, model.UserRoleUser)
	}
	if !created.IsActive {
		t.Fatal("expected created user to be active")
	}
	if created.CreateUserId != created.Id || created.UpdateUserId != created.Id {
		t.Fatal("expected self-registration audit ids to point to created user")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(created.PasswordHash), []byte("StrongPass123!")); err != nil {
		t.Fatalf("password hash does not match original password: %v", err)
	}
	if initializedUserID != userID {
		t.Fatalf("initializedUserID = %q, want %q", initializedUserID, userID)
	}
}

func TestCreateServiceIgnoresRequestedAdminRole(t *testing.T) {
	repo := installUserServiceTestDeps(t)

	userID, err := CreateService(model.UserDTO{
		Username: "alice",
		Password: "StrongPass123!",
		Role:     model.UserRoleAdmin,
	}, nil)
	if err != nil {
		t.Fatalf("CreateService returned error: %v", err)
	}

	created := repo.users[userID]
	if created.Role != model.UserRoleUser {
		t.Fatalf("Role = %q, want %q", created.Role, model.UserRoleUser)
	}
}

func TestCreateServiceRejectsExistingUsername(t *testing.T) {
	repo := installUserServiceTestDeps(t)
	existingID := primitive.NewObjectID()
	repo.users[existingID.Hex()] = model.UserEntity{Id: existingID, Username: "alice"}

	_, err := CreateService(model.UserDTO{
		Username: "alice",
		Password: "StrongPass123!",
	}, nil)
	if err == nil {
		t.Fatal("expected duplicate username error")
	}
	if len(repo.insertedIDs) != 0 {
		t.Fatalf("inserted count = %d, want 0", len(repo.insertedIDs))
	}
}

func TestUpdateServiceRejectsRoleChange(t *testing.T) {
	repo := installUserServiceTestDeps(t)
	userID := primitive.NewObjectID()
	repo.users[userID.Hex()] = model.UserEntity{
		Id:       userID,
		Username: "alice",
		Role:     model.UserRoleUser,
	}

	_, err := UpdateService(userID.Hex(), model.UserDTO{
		Role: model.UserRoleAdmin,
	})
	if err == nil {
		t.Fatal("expected role change error")
	}
	if repo.users[userID.Hex()].Role != model.UserRoleUser {
		t.Fatalf("Role = %q, want %q", repo.users[userID.Hex()].Role, model.UserRoleUser)
	}
}

func TestUpdateProfileServiceUpdatesAllowedProfileFields(t *testing.T) {
	repo := installUserServiceTestDeps(t)
	userID := primitive.NewObjectID()
	repo.users[userID.Hex()] = model.UserEntity{
		Id:       userID,
		Username: "alice",
		Gender:   model.GenderFemale,
	}

	updated, err := UpdateProfileService(userID.Hex(), model.UserProfileUpdateRequest{
		Nickname:  "Al",
		AvatarUrl: "https://example.test/avatar.png",
		Gender:    model.GenderOthers,
	})
	if err != nil {
		t.Fatalf("UpdateProfileService returned error: %v", err)
	}

	if updated.Nickname != "Al" {
		t.Fatalf("Nickname = %q, want Al", updated.Nickname)
	}
	if updated.AvatarUrl != "https://example.test/avatar.png" {
		t.Fatalf("AvatarUrl = %q", updated.AvatarUrl)
	}
	if updated.Gender != model.GenderOthers {
		t.Fatalf("Gender = %q, want %q", updated.Gender, model.GenderOthers)
	}
	if updated.UpdateUserId != userID {
		t.Fatalf("UpdateUserId = %s, want %s", updated.UpdateUserId.Hex(), userID.Hex())
	}
}

func TestChangePasswordServiceRequiresOldPasswordAndRevokesTokens(t *testing.T) {
	repo := installUserServiceTestDeps(t)
	var revokedUserID string
	revokeAllRefreshTokens = func(userId string) error {
		revokedUserID = userId
		return nil
	}
	userID := primitive.NewObjectID()
	oldHash, err := bcrypt.GenerateFromPassword([]byte("OldPass123!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash old password: %v", err)
	}
	repo.users[userID.Hex()] = model.UserEntity{
		Id:           userID,
		Username:     "alice",
		PasswordHash: string(oldHash),
	}

	err = ChangePasswordService(userID.Hex(), "OldPass123!", "NewPass123!")
	if err != nil {
		t.Fatalf("ChangePasswordService returned error: %v", err)
	}

	updated := repo.users[userID.Hex()]
	if err := bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("NewPass123!")); err != nil {
		t.Fatalf("new password hash does not match: %v", err)
	}
	if revokedUserID != userID.Hex() {
		t.Fatalf("revokedUserID = %q, want %q", revokedUserID, userID.Hex())
	}
}

func TestChangePasswordServiceRejectsWrongOldPassword(t *testing.T) {
	repo := installUserServiceTestDeps(t)
	var revoked bool
	revokeAllRefreshTokens = func(userId string) error {
		revoked = true
		return nil
	}
	userID := primitive.NewObjectID()
	oldHash, err := bcrypt.GenerateFromPassword([]byte("OldPass123!"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash old password: %v", err)
	}
	repo.users[userID.Hex()] = model.UserEntity{
		Id:           userID,
		PasswordHash: string(oldHash),
	}

	err = ChangePasswordService(userID.Hex(), "WrongPass123!", "NewPass123!")
	if err == nil {
		t.Fatal("expected invalid old password error")
	}
	if revoked {
		t.Fatal("expected no token revocation when old password is wrong")
	}
}

func TestDeleteServiceRejectsAdminAndRevokesUserTokens(t *testing.T) {
	t.Run("admin deletion rejected", func(t *testing.T) {
		repo := installUserServiceTestDeps(t)
		adminID := primitive.NewObjectID()
		repo.users[adminID.Hex()] = model.UserEntity{Id: adminID, Role: model.UserRoleAdmin}

		err := DeleteService(adminID.Hex())
		if err == nil {
			t.Fatal("expected admin deletion error")
		}
		if repo.deletedIDs[adminID.Hex()] {
			t.Fatal("expected admin user not to be deleted")
		}
	})

	t.Run("normal user deleted and tokens revoked", func(t *testing.T) {
		repo := installUserServiceTestDeps(t)
		var revokedUserID string
		revokeAllRefreshTokens = func(userId string) error {
			revokedUserID = userId
			return nil
		}
		userID := primitive.NewObjectID()
		repo.users[userID.Hex()] = model.UserEntity{Id: userID, Role: model.UserRoleUser}

		err := DeleteService(userID.Hex())
		if err != nil {
			t.Fatalf("DeleteService returned error: %v", err)
		}
		if !repo.deletedIDs[userID.Hex()] {
			t.Fatal("expected user to be deleted")
		}
		if revokedUserID != userID.Hex() {
			t.Fatalf("revokedUserID = %q, want %q", revokedUserID, userID.Hex())
		}
	})
}

func TestInitAdminUserCreatesAdminOnlyWhenNoAdminExists(t *testing.T) {
	t.Run("creates bootstrap admin", func(t *testing.T) {
		repo := installUserServiceTestDeps(t)
		var initializedUserID string
		initializeDefaultCategoriesForUser = func(userId string) error {
			initializedUserID = userId
			return nil
		}

		InitAdminUser()

		if len(repo.insertedIDs) != 1 {
			t.Fatalf("inserted count = %d, want 1", len(repo.insertedIDs))
		}
		created := repo.users[repo.insertedIDs[0]]
		if created.Role != model.UserRoleAdmin {
			t.Fatalf("Role = %q, want %q", created.Role, model.UserRoleAdmin)
		}
		if initializedUserID != created.Id.Hex() {
			t.Fatalf("initializedUserID = %q, want %q", initializedUserID, created.Id.Hex())
		}
	})

	t.Run("skips when admin already exists", func(t *testing.T) {
		repo := installUserServiceTestDeps(t)
		adminID := primitive.NewObjectID()
		repo.users[adminID.Hex()] = model.UserEntity{
			Id:       adminID,
			Username: "admin",
			Role:     model.UserRoleAdmin,
		}

		InitAdminUser()

		if len(repo.insertedIDs) != 0 {
			t.Fatalf("inserted count = %d, want 0", len(repo.insertedIDs))
		}
	})
}

func installUserServiceTestDeps(t *testing.T) *userRepoStub {
	t.Helper()

	originalRepo := userRepo
	originalInitCategories := initializeDefaultCategoriesForUser
	originalRevokeAll := revokeAllRefreshTokens

	stub := &userRepoStub{
		users:      map[string]model.UserEntity{},
		deletedIDs: map[string]bool{},
	}
	userRepo = stub
	initializeDefaultCategoriesForUser = func(userId string) error { return nil }
	revokeAllRefreshTokens = func(userId string) error { return nil }

	t.Cleanup(func() {
		userRepo = originalRepo
		initializeDefaultCategoriesForUser = originalInitCategories
		revokeAllRefreshTokens = originalRevokeAll
	})
	return stub
}

type userRepoStub struct {
	users       map[string]model.UserEntity
	insertErr   bool
	updateErr   bool
	insertedIDs []string
	deletedIDs  map[string]bool
}

func (stub *userRepoStub) GetUserByObjectId(plainId string) model.UserEntity {
	return stub.users[plainId]
}

func (stub *userRepoStub) GetUserByUsername(username string) model.UserEntity {
	for _, user := range stub.users {
		if user.Username == username {
			return user
		}
	}
	return model.UserEntity{}
}

func (stub *userRepoStub) GetUserByUsernameIncludeDeleted(username string) model.UserEntity {
	return stub.GetUserByUsername(username)
}

func (stub *userRepoStub) GetUserByEmail(email string) model.UserEntity {
	for _, user := range stub.users {
		if user.EmailAddress == email {
			return user
		}
	}
	return model.UserEntity{}
}

func (stub *userRepoStub) InsertUserByEntity(newEntity model.UserEntity) string {
	if stub.insertErr {
		return ""
	}
	stub.users[newEntity.Id.Hex()] = newEntity
	stub.insertedIDs = append(stub.insertedIDs, newEntity.Id.Hex())
	return newEntity.Id.Hex()
}

func (stub *userRepoStub) UpdateUserByEntity(plainId string, updatedEntity model.UserEntity) model.UserEntity {
	if stub.updateErr {
		return model.UserEntity{}
	}
	stub.users[plainId] = updatedEntity
	return updatedEntity
}

func (stub *userRepoStub) GetAllUsers(limit, offset int) []model.UserEntity {
	users := make([]model.UserEntity, 0, len(stub.users))
	for _, user := range stub.users {
		users = append(users, user)
	}
	return users
}

func (stub *userRepoStub) GetUsersByRole(role string) []model.UserEntity {
	var users []model.UserEntity
	for _, user := range stub.users {
		if user.Role == role {
			users = append(users, user)
		}
	}
	return users
}

func (stub *userRepoStub) CountAllUsers() int64 {
	return int64(len(stub.users))
}

func (stub *userRepoStub) DeleteUserByObjectId(plainId string) model.UserEntity {
	user := stub.users[plainId]
	if user.Id.IsZero() {
		return model.UserEntity{}
	}
	stub.deletedIDs[plainId] = true
	return user
}
