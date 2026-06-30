package manage_service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const maxBackupSize = 100 << 20

type Progress struct {
	Phase     string
	Entity    string
	Completed int
	Total     int
	Message   string
}

type ProgressFunc func(Progress)

type PreflightReport struct {
	Version            string
	Users              int
	UserConfigurations int
	Categories         int
	CashFlows          int
}

func emit(progress ProgressFunc, phase, entity string, completed, total int, message string) {
	if progress != nil {
		progress(Progress{Phase: phase, Entity: entity, Completed: completed, Total: total, Message: message})
	}
}

func PreflightBackupDestination(filePath string) error {
	if strings.TrimSpace(filePath) == "" {
		return fmt.Errorf("backup path cannot be empty")
	}
	abs, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("resolve backup path: %w", err)
	}
	parent := filepath.Dir(abs)
	info, err := os.Stat(parent)
	if err != nil {
		return fmt.Errorf("backup directory is unavailable: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("backup parent is not a directory: %s", parent)
	}
	if existing, err := os.Stat(abs); err == nil && existing.IsDir() {
		return fmt.Errorf("backup path is a directory: %s", abs)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("inspect backup destination: %w", err)
	}
	return nil
}

func PreflightRestore(filePath string) (PreflightReport, error) {
	var report PreflightReport
	info, err := os.Stat(filePath)
	if err != nil {
		return report, fmt.Errorf("inspect backup: %w", err)
	}
	if !info.Mode().IsRegular() {
		return report, fmt.Errorf("backup is not a regular file")
	}
	if info.Size() == 0 || info.Size() > maxBackupSize {
		return report, fmt.Errorf("backup size must be between 1 byte and %d bytes", maxBackupSize)
	}
	file, err := os.Open(filePath)
	if err != nil {
		return report, fmt.Errorf("open backup: %w", err)
	}
	defer file.Close()

	var backup BackupData
	decoder := json.NewDecoder(io.LimitReader(file, maxBackupSize+1))
	if err := decoder.Decode(&backup); err != nil {
		return report, fmt.Errorf("decode backup: %w", err)
	}
	if backup.Version != "1.0.0" && backup.Version != "2.0.0" {
		return report, fmt.Errorf("unsupported backup version %q", backup.Version)
	}
	if _, err := time.Parse(time.RFC3339, backup.Timestamp); err != nil {
		return report, fmt.Errorf("invalid backup timestamp: %w", err)
	}
	if len(backup.Users) == 0 {
		return report, fmt.Errorf("backup contains no users")
	}

	users := map[string]bool{}
	categories := map[string]string{}
	parents := map[string]string{}
	flows := map[string]bool{}
	for i, user := range backup.Users {
		id, err := requireObjectID(user, "id")
		if err != nil {
			return report, fmt.Errorf("users[%d]: %w", i, err)
		}
		if users[id] {
			return report, fmt.Errorf("users[%d]: duplicate id %s", i, id)
		}
		users[id] = true
		if err := requireStrings(user, "username", "password_hash", "create_time", "update_time", "role"); err != nil {
			return report, fmt.Errorf("users[%d]: %w", i, err)
		}
		if _, ok := user["is_active"].(bool); !ok {
			return report, fmt.Errorf("users[%d]: is_active must be a boolean", i)
		}
	}
	for i, config := range backup.UserConfigs {
		owner, err := requireObjectID(config, "belongs_user_id")
		if err != nil {
			return report, fmt.Errorf("user_configurations[%d]: %w", i, err)
		}
		if !users[owner] {
			return report, fmt.Errorf("user_configurations[%d]: unknown owner %s", i, owner)
		}
	}
	for i, category := range backup.Categories {
		id, err := requireObjectID(category, "id")
		if err != nil {
			return report, fmt.Errorf("categories[%d]: %w", i, err)
		}
		owner, err := requireObjectID(category, "belongs_user_id")
		if err != nil {
			return report, fmt.Errorf("categories[%d]: %w", i, err)
		}
		if !users[owner] {
			return report, fmt.Errorf("categories[%d]: unknown owner %s", i, owner)
		}
		if _, duplicate := categories[id]; duplicate {
			return report, fmt.Errorf("categories[%d]: duplicate id %s", i, id)
		}
		categories[id] = owner
		if err := requireStrings(category, "name", "type", "parent_id", "create_time", "update_time"); err != nil {
			return report, fmt.Errorf("categories[%d]: %w", i, err)
		}
		parents[id], _ = category["parent_id"].(string)
	}
	for id, parent := range parents {
		if parent != primitive.NilObjectID.Hex() && parent != "" && categories[parent] != categories[id] {
			return report, fmt.Errorf("category %s has missing or cross-user parent %s", id, parent)
		}
	}
	for i, flow := range backup.CashFlows {
		id, err := requireObjectID(flow, "id")
		if err != nil {
			return report, fmt.Errorf("cash_flows[%d]: %w", i, err)
		}
		if flows[id] {
			return report, fmt.Errorf("cash_flows[%d]: duplicate id %s", i, id)
		}
		flows[id] = true
		owner, err := requireObjectID(flow, "belongs_user_id")
		if err != nil {
			return report, fmt.Errorf("cash_flows[%d]: %w", i, err)
		}
		category, err := requireObjectID(flow, "category_id")
		if err != nil {
			return report, fmt.Errorf("cash_flows[%d]: %w", i, err)
		}
		if categories[category] != owner {
			return report, fmt.Errorf("cash_flows[%d]: category %s is missing or owned by another user", i, category)
		}
		if err := requireStrings(flow, "belongs_date", "description", "create_time", "update_time"); err != nil {
			return report, fmt.Errorf("cash_flows[%d]: %w", i, err)
		}
		if _, ok := flow["amount"].(float64); !ok {
			return report, fmt.Errorf("cash_flows[%d]: amount must be a number", i)
		}
	}
	report = PreflightReport{Version: backup.Version, Users: len(backup.Users), UserConfigurations: len(backup.UserConfigs), Categories: len(backup.Categories), CashFlows: len(backup.CashFlows)}
	return report, nil
}

func requireObjectID(values map[string]interface{}, field string) (string, error) {
	value, ok := values[field].(string)
	if !ok || !primitive.IsValidObjectID(value) {
		return "", fmt.Errorf("%s must be a valid object ID", field)
	}
	return value, nil
}

func requireStrings(values map[string]interface{}, fields ...string) error {
	for _, field := range fields {
		if _, ok := values[field].(string); !ok {
			return fmt.Errorf("%s must be a string", field)
		}
	}
	return nil
}
