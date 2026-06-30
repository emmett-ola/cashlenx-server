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
