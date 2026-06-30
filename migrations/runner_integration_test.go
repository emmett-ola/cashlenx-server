//go:build integration

package migrations

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestMySQLMigrationRunnerIntegration(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	t.Run("applies clean migration history", func(t *testing.T) {
		if err := Run(db); err != nil {
			t.Fatal(err)
		}
		items, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		var applied, dirty int
		if err := db.QueryRow("SELECT COUNT(*), COALESCE(SUM(dirty), 0) FROM schema_migrations").Scan(&applied, &dirty); err != nil {
			t.Fatal(err)
		}
		if applied != len(items) || dirty != 0 {
			t.Fatalf("applied=%d dirty=%d, want applied=%d dirty=0", applied, dirty, len(items))
		}
	})

	t.Run("compensates failed migration", func(t *testing.T) {
		item := Migration{
			Version:  999,
			Name:     "999_injected_failure.sql",
			Checksum: "integration-test",
			SQL:      "CREATE TABLE migration_rollback_probe (id INT); INVALID SQL;",
			Down:     "DROP TABLE migration_rollback_probe;",
		}
		err := apply(db, item)
		if err == nil || !strings.Contains(err.Error(), "migration was rolled back") {
			t.Fatalf("apply error = %v, want rollback confirmation", err)
		}
		var tableCount, migrationCount int
		if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'migration_rollback_probe'").Scan(&tableCount); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 999").Scan(&migrationCount); err != nil {
			t.Fatal(err)
		}
		if tableCount != 0 || migrationCount != 0 {
			t.Fatalf("tableCount=%d migrationCount=%d, want both zero", tableCount, migrationCount)
		}
	})

	t.Run("rejects out of order history", func(t *testing.T) {
		if _, err := db.Exec("DELETE FROM schema_migrations WHERE version = 11"); err != nil {
			t.Fatal(err)
		}
		err := Run(db)
		if err == nil || !strings.Contains(err.Error(), "version 011 is missing") {
			t.Fatalf("Run error = %v, want out-of-order rejection", err)
		}
	})
}
