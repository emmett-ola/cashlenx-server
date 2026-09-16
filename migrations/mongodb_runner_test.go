package migrations

import (
	"strings"
	"testing"
)

func TestLoadMongoMigrationsInVersionOrder(t *testing.T) {
	items, err := LoadMongo()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 || items[0].Version != 1 || items[1].Version != 10 || items[2].Version != 16 {
		t.Fatalf("unexpected MongoDB migration sequence: %#v", items)
	}
	for _, item := range items {
		if len(item.Checksum) != 64 {
			t.Fatalf("migration %03d has invalid checksum %q", item.Version, item.Checksum)
		}
	}
}

func TestValidateMongoAppliedHistory(t *testing.T) {
	items := []MongoMigration{
		{Version: 1, Name: "001_first.js", Checksum: "first"},
		{Version: 3, Name: "003_third.js", Checksum: "third"},
		{Version: 7, Name: "007_seventh.js", Checksum: "seventh"},
	}
	tests := []struct {
		name    string
		applied map[int]mongoAppliedMigration
		wantErr string
	}{
		{name: "empty", applied: map[int]mongoAppliedMigration{}},
		{name: "ordered prefix", applied: map[int]mongoAppliedMigration{
			1: {Version: 1, Name: "001_first.js", Checksum: "first"},
			3: {Version: 3, Name: "003_third.js", Checksum: "third"},
		}},
		{name: "gap", applied: map[int]mongoAppliedMigration{
			1: {Version: 1, Name: "001_first.js", Checksum: "first"},
			7: {Version: 7, Name: "007_seventh.js", Checksum: "seventh"},
		}, wantErr: "version 003 is missing"},
		{name: "unknown", applied: map[int]mongoAppliedMigration{
			99: {Version: 99, Name: "099_unknown.js", Checksum: "unknown"},
		}, wantErr: "unknown version 099"},
		{name: "filename drift", applied: map[int]mongoAppliedMigration{
			1: {Version: 1, Name: "001_renamed.js", Checksum: "first"},
		}, wantErr: "filename changed"},
		{name: "checksum drift", applied: map[int]mongoAppliedMigration{
			1: {Version: 1, Name: "001_first.js", Checksum: "changed"},
		}, wantErr: "checksum changed"},
		{name: "dirty", applied: map[int]mongoAppliedMigration{
			1: {Version: 1, Name: "001_first.js", Checksum: "first", Dirty: true},
		}, wantErr: "is dirty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMongoAppliedHistory(items, tt.applied)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateMongoAppliedHistory returned error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}
