package admin_cmd

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/auth/provider"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/manage_service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAdminRequiresAdminSession(t *testing.T) {
	t.Setenv("CASHLENX_CLI_SESSION_FILE", filepath.Join(t.TempDir(), "session.json"))
	adminSessionUserId = ""

	if err := AdminCmd.PersistentPreRunE(AdminCmd, nil); err == nil {
		t.Fatal("expected missing CLI session error")
	}
	if adminSessionUserId != "" {
		t.Fatalf("adminSessionUserId = %q, want empty", adminSessionUserId)
	}
}

func TestAdminPrintHelpers(t *testing.T) {
	printAdminUser(model.UserEntity{
		Id:              primitive.NewObjectID(),
		Username:        "admin",
		Role:            model.UserRoleAdmin,
		IsActive:        true,
		Nickname:        "Administrator",
		EmailAddress:    "admin@example.test",
		IsEmailVerified: true,
		Gender:          "others",
		BaseEntity: model.BaseEntity{
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		},
	})

	if got := truncateUserField("abcdef", 3); got != "abc" {
		t.Fatalf("truncateUserField tiny limit = %q, want abc", got)
	}
	if got := truncateUserField("abcdefghijkl", 8); got != "abcde..." {
		t.Fatalf("truncateUserField ellipsis = %q, want abcde...", got)
	}
	if got := truncateUserField("abc", 8); got != "abc" {
		t.Fatalf("truncateUserField short = %q, want abc", got)
	}
}

func TestAdminPersistentPreRunStoresAdminUserID(t *testing.T) {
	const adminID = "507f1f77bcf86cd799439011"
	original := requireAdminSession
	requireAdminSession = func() (*provider.Claims, error) {
		return &provider.Claims{UserID: adminID, Username: "admin", Role: model.UserRoleAdmin}, nil
	}
	t.Cleanup(func() {
		requireAdminSession = original
		adminSessionUserId = ""
	})

	if err := AdminCmd.PersistentPreRunE(AdminCmd, nil); err != nil {
		t.Fatalf("PersistentPreRunE returned error: %v", err)
	}
	if adminSessionUserId != adminID {
		t.Fatalf("adminSessionUserId = %q, want %q", adminSessionUserId, adminID)
	}
}

func TestAdminUserCommandsPassInputsToServices(t *testing.T) {
	adminSessionUserId = primitive.NewObjectID().Hex()
	userID := primitive.NewObjectID()
	var createReq model.UserDTO
	var createActor *string
	var listLimit, listOffset int
	var getID, updateID, deleteID string
	var updateReq model.UserDTO

	originalCreate := createAdminUser
	createAdminUser = func(req model.UserDTO, actor *string) (string, error) {
		createReq = req
		createActor = actor
		return userID.Hex(), nil
	}
	originalGet := getAdminUser
	getAdminUser = func(id string) model.UserEntity {
		getID = id
		return testAdminUser(userID, "alice")
	}
	originalList := listAdminUsers
	listAdminUsers = func(limit, offset int) []model.UserEntity {
		listLimit = limit
		listOffset = offset
		return []model.UserEntity{testAdminUser(userID, "alice")}
	}
	originalCount := countAdminUsers
	countAdminUsers = func() int64 { return 1 }
	originalUpdate := updateAdminUser
	updateAdminUser = func(id string, req model.UserDTO) (model.UserEntity, error) {
		updateID = id
		updateReq = req
		return testAdminUser(userID, req.Username), nil
	}
	originalDelete := deleteAdminUser
	deleteAdminUser = func(id string) error {
		deleteID = id
		return nil
	}
	t.Cleanup(func() {
		createAdminUser = originalCreate
		getAdminUser = originalGet
		listAdminUsers = originalList
		countAdminUsers = originalCount
		updateAdminUser = originalUpdate
		deleteAdminUser = originalDelete
		resetAdminCommandState()
	})

	adminUserUsername, adminUserPassword, adminUserEmail, adminUserGender, adminUserEmailVerified = "alice", "secret", "alice@example.test", model.GenderFemale, true
	if err := userCreateCmd.RunE(userCreateCmd, nil); err != nil {
		t.Fatalf("user create RunE returned error: %v", err)
	}
	adminUserLimit, adminUserOffset = 7, 14
	if err := userListCmd.RunE(userListCmd, nil); err != nil {
		t.Fatalf("user list RunE returned error: %v", err)
	}
	adminUserId = userID.Hex()
	if err := userGetCmd.RunE(userGetCmd, nil); err != nil {
		t.Fatalf("user get RunE returned error: %v", err)
	}
	adminUserUsername, adminUserEmail, adminUserEmailVerified, adminUserEmailVerifiedOn = "alice2", "alice2@example.test", true, true
	if err := userUpdateCmd.RunE(userUpdateCmd, nil); err != nil {
		t.Fatalf("user update RunE returned error: %v", err)
	}
	if err := userDeleteCmd.RunE(userDeleteCmd, nil); err != nil {
		t.Fatalf("user delete RunE returned error: %v", err)
	}

	if createReq.Username != "alice" || createReq.Password != "secret" || createReq.EmailAddress != "alice@example.test" || createActor == nil || *createActor != adminSessionUserId {
		t.Fatalf("create req = %+v actor=%v", createReq, createActor)
	}
	if listLimit != 7 || listOffset != 14 {
		t.Fatalf("list args = %d, %d", listLimit, listOffset)
	}
	if getID != userID.Hex() {
		t.Fatalf("get id = %q", getID)
	}
	if updateID != userID.Hex() || updateReq.Username != "alice2" || updateReq.EmailAddress != "alice2@example.test" || !updateReq.IsEmailVerified {
		t.Fatalf("update id=%q req=%+v", updateID, updateReq)
	}
	if deleteID != userID.Hex() {
		t.Fatalf("delete id = %q", deleteID)
	}
}

func TestAdminBackupRestoreCommandsPassPathsToServices(t *testing.T) {
	var dumpPath, restoreServicePath string
	originalDump := adminDumpWithProgress
	adminDumpWithProgress = func(path string, _ manage_service.ProgressFunc) (manage_service.OperationStats, error) {
		dumpPath = path
		return manage_service.OperationStats{Users: manage_service.EntityStats{Success: 1}}, nil
	}
	originalRestore := adminRestoreWithProgress
	adminRestoreWithProgress = func(path string, _ manage_service.ProgressFunc) (manage_service.OperationStats, error) {
		restoreServicePath = path
		return manage_service.OperationStats{Categories: manage_service.EntityStats{Success: 1}}, nil
	}
	t.Cleanup(func() {
		adminDumpWithProgress = originalDump
		adminRestoreWithProgress = originalRestore
		resetAdminCommandState()
	})

	backupPath = "backup.json"
	if err := backupCmd.RunE(backupCmd, nil); err != nil {
		t.Fatalf("backup RunE returned error: %v", err)
	}
	restorePath, forceRestore = "restore.json", true
	if err := restoreBackupCmd.RunE(restoreBackupCmd, nil); err != nil {
		t.Fatalf("restore RunE returned error: %v", err)
	}

	if dumpPath != "backup.json" {
		t.Fatalf("dump path = %q", dumpPath)
	}
	if restoreServicePath != "restore.json" {
		t.Fatalf("restore path = %q", restoreServicePath)
	}
}

func testAdminUser(id primitive.ObjectID, username string) model.UserEntity {
	return model.UserEntity{
		Id:           id,
		Username:     username,
		Role:         model.UserRoleUser,
		IsActive:     true,
		EmailAddress: username + "@example.test",
	}
}

func resetAdminCommandState() {
	adminSessionUserId = ""
	adminUserId = ""
	adminUserUsername = ""
	adminUserPassword = ""
	adminUserNickname = ""
	adminUserAvatarURL = ""
	adminUserEmail = ""
	adminUserGender = ""
	adminUserEmailVerified = false
	adminUserEmailVerifiedOn = false
	adminUserLimit = 20
	adminUserOffset = 0
	backupPath = ""
	restorePath = ""
	forceRestore = false
}
