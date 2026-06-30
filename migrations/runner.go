package migrations

import (
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed *.sql
var migrationFiles embed.FS

type Migration struct {
	Version  int
	Name     string
	SQL      string
	Checksum string
	Down     string
}

var baselineTables = []string{"users", "categories", "cash_flows", "operation_confirm_codes", "refresh_tokens", "user_configurations"}

func Load() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationFiles, ".")
	if err != nil {
		return nil, err
	}
	items := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") || strings.HasSuffix(entry.Name(), ".down.sql") {
			continue
		}
		prefix, _, ok := strings.Cut(entry.Name(), "_")
		if !ok {
			return nil, fmt.Errorf("invalid migration filename %q", entry.Name())
		}
		version, err := strconv.Atoi(prefix)
		if err != nil {
			return nil, fmt.Errorf("invalid migration version in %q: %w", entry.Name(), err)
		}
		content, err := migrationFiles.ReadFile(entry.Name())
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(content)
		downName := strings.TrimSuffix(entry.Name(), ".sql") + ".down.sql"
		down, downErr := migrationFiles.ReadFile(downName)
		if downErr != nil && !errors.Is(downErr, fs.ErrNotExist) {
			return nil, downErr
		}
		items = append(items, Migration{Version: version, Name: entry.Name(), SQL: string(content), Checksum: hex.EncodeToString(sum[:]), Down: string(down)})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Version < items[j].Version })
	for i := 1; i < len(items); i++ {
		if items[i-1].Version == items[i].Version {
			return nil, fmt.Errorf("duplicate SQL migration version %03d", items[i].Version)
		}
	}
	return items, nil
}

func Run(db *sql.DB) error {
	items, err := Load()
	if err != nil {
		return err
	}
	if err := ensureTrackingTable(db); err != nil {
		return err
	}
	applied, err := loadApplied(db)
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		if err := initializeBaseline(db, items); err != nil {
			return err
		}
		applied, err = loadApplied(db)
		if err != nil {
			return err
		}
	}
	if err := validateAppliedHistory(items, applied); err != nil {
		return err
	}
	for _, item := range items {
		state, ok := applied[item.Version]
		if ok {
			if state.dirty {
				return fmt.Errorf("migration %03d is dirty; restore or repair it before retrying", item.Version)
			}
			if state.checksum != item.Checksum {
				return fmt.Errorf("migration %03d checksum changed after application", item.Version)
			}
			continue
		}
		if err := apply(db, item); err != nil {
			return err
		}
	}
	return nil
}

type appliedMigration struct {
	name     string
	checksum string
	dirty    bool
}

func validateAppliedHistory(items []Migration, applied map[int]appliedMigration) error {
	known := make(map[int]int, len(items))
	for index, item := range items {
		known[item.Version] = index
	}

	highestAppliedIndex := -1
	for version, state := range applied {
		index, ok := known[version]
		if !ok {
			return fmt.Errorf("schema_migrations contains unknown version %03d", version)
		}
		if state.name != items[index].Name {
			return fmt.Errorf("migration %03d filename changed after application: have %q, expected %q", version, state.name, items[index].Name)
		}
		if index > highestAppliedIndex {
			highestAppliedIndex = index
		}
	}

	for index := 0; index <= highestAppliedIndex; index++ {
		if _, ok := applied[items[index].Version]; !ok {
			return fmt.Errorf("migration history is out of order: version %03d is missing before an applied migration", items[index].Version)
		}
	}
	return nil
}

func ensureTrackingTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
        version INT NOT NULL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        checksum CHAR(64) NOT NULL,
        dirty BOOLEAN NOT NULL DEFAULT TRUE,
        applied_at TIMESTAMP NULL
    ) ENGINE=InnoDB`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	return nil
}

func loadApplied(db *sql.DB) (map[int]appliedMigration, error) {
	rows, err := db.Query("SELECT version, name, checksum, dirty FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("query schema migrations: %w", err)
	}
	defer rows.Close()
	result := map[int]appliedMigration{}
	for rows.Next() {
		var version int
		var state appliedMigration
		if err := rows.Scan(&version, &state.name, &state.checksum, &state.dirty); err != nil {
			return nil, err
		}
		result[version] = state
	}
	return result, rows.Err()
}

func initializeBaseline(db *sql.DB, items []Migration) error {
	existing := 0
	for _, table := range baselineTables {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`, table).Scan(&count); err != nil {
			return err
		}
		existing += count
	}
	if existing == 0 {
		return nil
	}
	if existing != len(baselineTables) {
		return fmt.Errorf("refusing to baseline partial MySQL schema: found %d of %d core tables", existing, len(baselineTables))
	}
	baselineVersion := 11
	var activeScopeColumn int
	if err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'categories' AND column_name = 'active_scope_key'`).Scan(&activeScopeColumn); err != nil {
		return err
	}
	if activeScopeColumn == 1 {
		baselineVersion = 12
	}
	for _, item := range items {
		if item.Version > baselineVersion {
			continue
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (version, name, checksum, dirty, applied_at) VALUES (?, ?, ?, FALSE, ?)`, item.Version, item.Name, item.Checksum, time.Now().UTC()); err != nil {
			return fmt.Errorf("baseline migration %03d: %w", item.Version, err)
		}
	}
	return nil
}

func apply(db *sql.DB, item Migration) error {
	if _, err := db.Exec(`INSERT INTO schema_migrations (version, name, checksum, dirty) VALUES (?, ?, ?, TRUE)`, item.Version, item.Name, item.Checksum); err != nil {
		return fmt.Errorf("mark migration %03d dirty: %w", item.Version, err)
	}
	for _, statement := range splitSQL(item.SQL) {
		upper := strings.ToUpper(strings.TrimSpace(statement))
		if strings.HasPrefix(upper, "USE ") || strings.HasPrefix(upper, "CREATE SCHEMA ") || strings.HasPrefix(upper, "CREATE DATABASE ") {
			continue
		}
		if _, err := db.Exec(statement); err != nil {
			applyErr := fmt.Errorf("apply migration %03d (%s): %w", item.Version, item.Name, err)
			if rollbackErr := rollback(db, item); rollbackErr != nil {
				return fmt.Errorf("%w; automatic rollback also failed: %v", applyErr, rollbackErr)
			}
			return fmt.Errorf("%w; migration was rolled back", applyErr)
		}
	}
	if _, err := db.Exec(`UPDATE schema_migrations SET dirty = FALSE, applied_at = ? WHERE version = ?`, time.Now().UTC(), item.Version); err != nil {
		return fmt.Errorf("complete migration %03d: %w", item.Version, err)
	}
	return nil
}

func rollback(db *sql.DB, item Migration) error {
	if strings.TrimSpace(item.Down) == "" {
		return fmt.Errorf("migration %03d has no down script", item.Version)
	}
	var rollbackErrors []string
	for _, statement := range splitSQL(item.Down) {
		if _, err := db.Exec(statement); err != nil && !ignorableRollbackError(err) {
			rollbackErrors = append(rollbackErrors, err.Error())
		}
	}
	if len(rollbackErrors) > 0 {
		return fmt.Errorf("migration %03d rollback errors: %s", item.Version, strings.Join(rollbackErrors, "; "))
	}
	if _, err := db.Exec("DELETE FROM schema_migrations WHERE version = ? AND dirty = TRUE", item.Version); err != nil {
		return fmt.Errorf("clear rolled-back migration %03d: %w", item.Version, err)
	}
	return nil
}

func ignorableRollbackError(err error) bool {
	for current := err; current != nil; current = errors.Unwrap(current) {
		value := reflect.ValueOf(current)
		if value.Kind() == reflect.Ptr && !value.IsNil() {
			value = value.Elem()
		}
		if value.Kind() != reflect.Struct {
			continue
		}
		field := value.FieldByName("Number")
		if field.IsValid() && field.CanUint() {
			number := field.Uint()
			// 1091: object to drop does not exist. 1061: duplicate index name.
			return number == 1091 || number == 1061
		}
	}
	return false
}

func splitSQL(input string) []string {
	var statements []string
	var current strings.Builder
	var quote rune
	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if quote == 0 && r == '-' && i+1 < len(runes) && runes[i+1] == '-' {
			for i < len(runes) && runes[i] != '\n' {
				i++
			}
			current.WriteRune('\n')
			continue
		}
		if r == '\'' || r == '"' || r == '`' {
			if quote == 0 {
				quote = r
			} else if quote == r {
				if i+1 < len(runes) && runes[i+1] == r {
					current.WriteRune(r)
					i++
				} else {
					quote = 0
				}
			}
		}
		if r == ';' && quote == 0 {
			if statement := strings.TrimSpace(current.String()); statement != "" {
				statements = append(statements, statement)
			}
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}
	if statement := strings.TrimSpace(current.String()); statement != "" {
		statements = append(statements, statement)
	}
	return statements
}
