package manage_service

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoIndexDefinitionsExcludeLegacyFlowType(t *testing.T) {
	for _, index := range mongoCashFlowIndexes() {
		for _, key := range index.Keys.(bson.D) {
			if key.Key == "flow_type" {
				t.Fatal("cash-flow index still contains legacy flow_type")
			}
		}
	}

	if got := len(mongoCashFlowIndexes()); got != 3 {
		t.Fatalf("expected 3 cash-flow indexes, got %d", got)
	}
	if got := len(mongoCategoryIndexes()); got != 3 {
		t.Fatalf("expected 3 category indexes, got %d", got)
	}
}

func TestMySQLIndexDefinitionsUseCurrentTables(t *testing.T) {
	for _, index := range mysqlIndexes() {
		if index.table != "cash_flows" && index.table != "categories" {
			t.Fatalf("unexpected legacy table %q", index.table)
		}
	}
}
