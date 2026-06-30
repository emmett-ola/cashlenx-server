package migrations

import (
	"strings"
	"testing"
)

func TestLoadMigrationsInVersionOrder(t *testing.T) {
	items, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 || items[0].Version != 2 || items[len(items)-1].Version != 12 {
		t.Fatalf("unexpected migration range: %#v", items)
	}
	if items[len(items)-1].Down == "" {
		t.Fatal("migration 012 must provide rollback SQL")
	}
	for i := 1; i < len(items); i++ {
		if items[i-1].Version >= items[i].Version {
			t.Fatal("migrations are not strictly ordered")
		}
	}
}

func TestSplitSQLPreservesQuotedSemicolonsAndDropsComments(t *testing.T) {
	statements := splitSQL("-- comment\nSELECT 'a;b'; INSERT INTO `x` VALUES (1);")
	if len(statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(statements))
	}
	if !strings.Contains(statements[0], "a;b") || strings.Contains(statements[0], "comment") {
		t.Fatalf("unexpected first statement %q", statements[0])
	}
}

func TestValidateAppliedHistory(t *testing.T) {
	items := []Migration{
		{Version: 2, Name: "002_create.sql"},
		{Version: 4, Name: "004_alter.sql"},
		{Version: 7, Name: "007_index.sql"},
	}
	tests := []struct {
		name    string
		applied map[int]appliedMigration
		wantErr string
	}{
		{name: "empty", applied: map[int]appliedMigration{}},
		{name: "ordered prefix", applied: map[int]appliedMigration{
			2: {name: "002_create.sql"},
			4: {name: "004_alter.sql"},
		}},
		{name: "gap", applied: map[int]appliedMigration{
			2: {name: "002_create.sql"},
			7: {name: "007_index.sql"},
		}, wantErr: "version 004 is missing"},
		{name: "unknown version", applied: map[int]appliedMigration{
			2:  {name: "002_create.sql"},
			99: {name: "099_unknown.sql"},
		}, wantErr: "unknown version 099"},
		{name: "filename drift", applied: map[int]appliedMigration{
			2: {name: "002_renamed.sql"},
		}, wantErr: "filename changed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAppliedHistory(items, tt.applied)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateAppliedHistory returned error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateAppliedHistory error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}
