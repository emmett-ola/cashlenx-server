package manage_service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPreflightRestoreRejectsBrokenRelationships(t *testing.T) {
	userID := "507f1f77bcf86cd799439011"
	categoryID := "507f1f77bcf86cd799439012"
	backup := BackupData{
		Version: "2.0.0", Timestamp: time.Now().UTC().Format(time.RFC3339),
		Users:     []map[string]interface{}{{"id": userID, "username": "u", "password_hash": "h", "create_time": "2026-01-01T00:00:00Z", "update_time": "2026-01-01T00:00:00Z", "is_active": true, "role": "user"}},
		CashFlows: []map[string]interface{}{{"id": "507f1f77bcf86cd799439013", "belongs_user_id": userID, "category_id": categoryID, "belongs_date": "2026-01-01T00:00:00Z", "amount": 1.0, "description": "x", "create_time": "2026-01-01T00:00:00Z", "update_time": "2026-01-01T00:00:00Z"}},
	}
	path := filepath.Join(t.TempDir(), "backup.json")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(file).Encode(backup); err != nil {
		t.Fatal(err)
	}
	file.Close()
	if _, err := PreflightRestore(path); err == nil {
		t.Fatal("expected relationship validation error")
	}
}

func TestPreflightBackupDestinationRejectsDirectory(t *testing.T) {
	if err := PreflightBackupDestination(t.TempDir()); err == nil {
		t.Fatal("expected directory rejection")
	}
}
