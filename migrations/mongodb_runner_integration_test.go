//go:build integration

package migrations

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoMigrationRunnerIntegration(t *testing.T) {
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
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatal(err)
	}

	items := []MongoMigration{
		{Version: 1, Name: "001_first.js", Checksum: "first"},
		{Version: 3, Name: "003_third.js", Checksum: "third"},
	}
	applyProbe := func(ctx context.Context, db *mongo.Database, item MongoMigration) error {
		_, err := db.Collection("migration_probe").InsertOne(ctx, bson.M{"_id": item.Version})
		return err
	}

	t.Run("fresh existing and repeat-safe sequence", func(t *testing.T) {
		db := integrationMongoDatabase(t, client, "repeat")
		if _, err := db.Collection("users").InsertOne(ctx, bson.M{"username": "existing"}); err != nil {
			t.Fatal(err)
		}
		if err := runMongo(ctx, db, items, applyProbe); err != nil {
			t.Fatal(err)
		}
		if err := runMongo(ctx, db, items, func(context.Context, *mongo.Database, MongoMigration) error {
			return errors.New("repeat unexpectedly reapplied a migration")
		}); err != nil {
			t.Fatal(err)
		}
		for collection, want := range map[string]int64{"schema_migrations": 2, "migration_probe": 2, "users": 1} {
			got, err := db.Collection(collection).CountDocuments(ctx, bson.M{})
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("%s count = %d, want %d", collection, got, want)
			}
		}
	})

	t.Run("checksum change blocks", func(t *testing.T) {
		db := integrationMongoDatabase(t, client, "checksum")
		if err := runMongo(ctx, db, items, applyProbe); err != nil {
			t.Fatal(err)
		}
		changed := append([]MongoMigration(nil), items...)
		changed[0].Checksum = "modified"
		err := runMongo(ctx, db, changed, applyProbe)
		if err == nil || !strings.Contains(err.Error(), "checksum changed") {
			t.Fatalf("runMongo error = %v, want checksum rejection", err)
		}
	})

	t.Run("new lower version blocks reordered history", func(t *testing.T) {
		db := integrationMongoDatabase(t, client, "reordered")
		if err := runMongo(ctx, db, items, applyProbe); err != nil {
			t.Fatal(err)
		}
		reordered := []MongoMigration{
			items[0],
			{Version: 2, Name: "002_inserted.js", Checksum: "inserted"},
			items[1],
		}
		err := runMongo(ctx, db, reordered, applyProbe)
		if err == nil || !strings.Contains(err.Error(), "version 002 is missing") {
			t.Fatalf("runMongo error = %v, want ordering rejection", err)
		}
	})

	t.Run("failure remains dirty and blocks retry", func(t *testing.T) {
		db := integrationMongoDatabase(t, client, "dirty")
		failure := errors.New("injected migration failure")
		err := runMongo(ctx, db, items[:1], func(context.Context, *mongo.Database, MongoMigration) error {
			return failure
		})
		if err == nil || !strings.Contains(err.Error(), failure.Error()) {
			t.Fatalf("runMongo error = %v, want injected failure", err)
		}
		var state mongoAppliedMigration
		if err := db.Collection(mongoMigrationCollection).FindOne(ctx, bson.M{"_id": 1}).Decode(&state); err != nil {
			t.Fatal(err)
		}
		if !state.Dirty {
			t.Fatal("failed migration was not retained as dirty")
		}
		err = runMongo(ctx, db, items[:1], applyProbe)
		if err == nil || !strings.Contains(err.Error(), "is dirty") {
			t.Fatalf("retry error = %v, want dirty-state rejection", err)
		}
	})
}

func integrationMongoDatabase(t *testing.T, client *mongo.Client, suffix string) *mongo.Database {
	t.Helper()
	name := fmt.Sprintf("cashlenx_migration_%s_%d", suffix, time.Now().UnixNano())
	db := client.Database(name)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = db.Drop(ctx)
	})
	return db
}
