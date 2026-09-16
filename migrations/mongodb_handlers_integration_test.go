//go:build integration

package migrations_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/migrations"
	"github.com/macar-x/cashlenx-server/service/manage_service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoMigrationHandlersIntegration(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database(fmt.Sprintf("cashlenx_migration_handlers_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		_ = db.Drop(cleanupCtx)
	})

	if _, err := db.Collection("cash_flows").InsertOne(ctx, bson.M{"belongs_user_id": "existing", "is_delete": false}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Collection("cash_flows").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "flow_type", Value: 1}}, Options: options.Index().SetName("flow_type_1"),
	}); err != nil {
		t.Fatal(err)
	}

	if err := migrations.RunMongo(ctx, db, manage_service.ApplyMongoMigration); err != nil {
		t.Fatal(err)
	}
	if err := migrations.RunMongo(ctx, db, manage_service.ApplyMongoMigration); err != nil {
		t.Fatalf("repeat run failed: %v", err)
	}

	var applied, dirty int64
	applied, err = db.Collection("schema_migrations").CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatal(err)
	}
	dirty, err = db.Collection("schema_migrations").CountDocuments(ctx, bson.M{"dirty": true})
	if err != nil {
		t.Fatal(err)
	}
	if applied != 3 || dirty != 0 {
		t.Fatalf("applied=%d dirty=%d, want applied=3 dirty=0", applied, dirty)
	}

	assertMongoIndex(t, ctx, db.Collection("cash_flows"), "cash_flows_user_date_index", true)
	assertMongoIndex(t, ctx, db.Collection("cash_flows"), "flow_type_1", false)
	assertMongoIndex(t, ctx, db.Collection("categories"), "categories_active_scope_unique_index", true)
	assertMongoIndex(t, ctx, db.Collection("operation_confirm_codes"), "verification_token_1", true)
	assertMongoIndex(t, ctx, db.Collection("budgets"), "budgets_active_scope_unique_index", true)
}

func assertMongoIndex(t *testing.T, ctx context.Context, collection *mongo.Collection, name string, want bool) {
	t.Helper()
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cursor.Close(ctx)
	found := false
	for cursor.Next(ctx) {
		var index struct {
			Name string `bson:"name"`
		}
		if err := cursor.Decode(&index); err != nil {
			t.Fatal(err)
		}
		if index.Name == name {
			found = true
		}
	}
	if err := cursor.Err(); err != nil {
		t.Fatal(err)
	}
	if found != want {
		t.Fatalf("index %s on %s found=%v, want %v", name, collection.Name(), found, want)
	}
}
