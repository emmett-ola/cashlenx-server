package manage_service

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAdminRestoreCompensatesFailedDestructiveRestore(t *testing.T) {
	input := writeMinimalAdminBackup(t)
	inputData, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}

	var restoreCalls int
	var phases []string
	stats, err := adminRestoreDatabaseWithDependencies(
		input,
		func(progress Progress) { phases = append(phases, progress.Phase) },
		func(snapshotPath string) (OperationStats, error) {
			return OperationStats{}, os.WriteFile(snapshotPath, inputData, 0o600)
		},
		func(path string, _ ProgressFunc) (OperationStats, error) {
			restoreCalls++
			if restoreCalls == 1 {
				return OperationStats{Users: EntityStats{Failed: 1}}, errors.New("injected restore failure")
			}
			if path == input {
				t.Fatal("rollback must use the generated snapshot")
			}
			return OperationStats{}, nil
		},
	)
	if err == nil || !strings.Contains(err.Error(), "database was rolled back") {
		t.Fatalf("restore error = %v, want successful rollback report", err)
	}
	if stats.Users.Failed != 1 || restoreCalls != 2 {
		t.Fatalf("stats=%+v restoreCalls=%d", stats, restoreCalls)
	}
	if !containsString(phases, "snapshot") || !containsString(phases, "rollback") {
		t.Fatalf("progress phases = %v", phases)
	}
}

func TestAdminRestoreReportsRollbackFailure(t *testing.T) {
	input := writeMinimalAdminBackup(t)
	inputData, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}

	var restoreCalls int
	_, err = adminRestoreDatabaseWithDependencies(
		input,
		nil,
		func(snapshotPath string) (OperationStats, error) {
			return OperationStats{}, os.WriteFile(snapshotPath, inputData, 0o600)
		},
		func(string, ProgressFunc) (OperationStats, error) {
			restoreCalls++
			if restoreCalls == 1 {
				return OperationStats{}, errors.New("injected restore failure")
			}
			return OperationStats{}, errors.New("injected rollback failure")
		},
	)
	if err == nil || !strings.Contains(err.Error(), "rollback failed: injected rollback failure") {
		t.Fatalf("restore error = %v, want rollback failure", err)
	}
}

func writeMinimalAdminBackup(t *testing.T) string {
	t.Helper()
	backup := BackupData{
		Version:   "2.0.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Users: []map[string]interface{}{{
			"id":            "507f1f77bcf86cd799439011",
			"username":      "admin",
			"password_hash": "hash",
			"create_time":   "2026-01-01T00:00:00Z",
			"update_time":   "2026-01-01T00:00:00Z",
			"is_active":     true,
			"role":          "admin",
		}},
	}
	path := filepath.Join(t.TempDir(), "backup.json")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(file).Encode(backup); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
