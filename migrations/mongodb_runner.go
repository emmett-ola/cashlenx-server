package migrations

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongoMigrationCollection = "schema_migrations"

//go:embed *.js
var mongoMigrationFiles embed.FS

type MongoMigration struct {
	Version  int
	Name     string
	Checksum string
}

type mongoAppliedMigration struct {
	Version  int        `bson:"_id"`
	Name     string     `bson:"name"`
	Checksum string     `bson:"checksum"`
	Dirty    bool       `bson:"dirty"`
	Started  time.Time  `bson:"started_at"`
	Applied  *time.Time `bson:"applied_at,omitempty"`
}

type MongoMigrationApply func(context.Context, *mongo.Database, MongoMigration) error

func LoadMongo() ([]MongoMigration, error) {
	entries, err := fs.ReadDir(mongoMigrationFiles, ".")
	if err != nil {
		return nil, err
	}
	items := make([]MongoMigration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".js") {
			continue
		}
		prefix, _, ok := strings.Cut(entry.Name(), "_")
		if !ok {
			return nil, fmt.Errorf("invalid MongoDB migration filename %q", entry.Name())
		}
		version, err := strconv.Atoi(prefix)
		if err != nil {
			return nil, fmt.Errorf("invalid MongoDB migration version in %q: %w", entry.Name(), err)
		}
		content, err := mongoMigrationFiles.ReadFile(entry.Name())
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(content)
		items = append(items, MongoMigration{
			Version: version, Name: entry.Name(), Checksum: hex.EncodeToString(sum[:]),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Version < items[j].Version })
	for i := 1; i < len(items); i++ {
		if items[i-1].Version == items[i].Version {
			return nil, fmt.Errorf("duplicate MongoDB migration version %03d", items[i].Version)
		}
	}
	return items, nil
}

func RunMongo(ctx context.Context, db *mongo.Database, apply MongoMigrationApply) error {
	items, err := LoadMongo()
	if err != nil {
		return err
	}
	return runMongo(ctx, db, items, apply)
}

func VerifyMongo(ctx context.Context, db *mongo.Database) error {
	items, err := LoadMongo()
	if err != nil {
		return err
	}
	applied, err := loadMongoApplied(ctx, db)
	if err != nil {
		return err
	}
	if err := validateMongoAppliedHistory(items, applied); err != nil {
		return err
	}
	if len(applied) != len(items) {
		return fmt.Errorf("MongoDB migration history has %d pending migration(s)", len(items)-len(applied))
	}
	return nil
}

func runMongo(ctx context.Context, db *mongo.Database, items []MongoMigration, apply MongoMigrationApply) error {
	if db == nil {
		return errors.New("MongoDB migration database is required")
	}
	if apply == nil {
		return errors.New("MongoDB migration apply function is required")
	}
	applied, err := loadMongoApplied(ctx, db)
	if err != nil {
		return err
	}
	if err := validateMongoAppliedHistory(items, applied); err != nil {
		return err
	}

	collection := db.Collection(mongoMigrationCollection)
	for _, item := range items {
		state, ok := applied[item.Version]
		if ok {
			if state.Dirty {
				return fmt.Errorf("MongoDB migration %03d is dirty; restore or repair it before retrying", item.Version)
			}
			if state.Checksum != item.Checksum {
				return fmt.Errorf("MongoDB migration %03d checksum changed after application", item.Version)
			}
			continue
		}

		started := time.Now().UTC()
		state = mongoAppliedMigration{
			Version: item.Version, Name: item.Name, Checksum: item.Checksum, Dirty: true, Started: started,
		}
		if _, err := collection.InsertOne(ctx, state); err != nil {
			return fmt.Errorf("mark MongoDB migration %03d dirty: %w", item.Version, err)
		}
		if err := apply(ctx, db, item); err != nil {
			return fmt.Errorf("apply MongoDB migration %03d (%s): %w; dirty state retained", item.Version, item.Name, err)
		}
		appliedAt := time.Now().UTC()
		result, err := collection.UpdateOne(ctx, bson.M{"_id": item.Version, "dirty": true}, bson.M{"$set": bson.M{
			"dirty": false, "applied_at": appliedAt,
		}})
		if err != nil {
			return fmt.Errorf("complete MongoDB migration %03d: %w", item.Version, err)
		}
		if result.MatchedCount != 1 {
			return fmt.Errorf("complete MongoDB migration %03d: dirty record no longer matches", item.Version)
		}
	}
	return nil
}

func loadMongoApplied(ctx context.Context, db *mongo.Database) (map[int]mongoAppliedMigration, error) {
	result := map[int]mongoAppliedMigration{}
	cursor, err := db.Collection(mongoMigrationCollection).Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("query MongoDB migrations: %w", err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var state mongoAppliedMigration
		if err := cursor.Decode(&state); err != nil {
			return nil, fmt.Errorf("decode MongoDB migration state: %w", err)
		}
		result[state.Version] = state
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate MongoDB migrations: %w", err)
	}
	return result, nil
}

func validateMongoAppliedHistory(items []MongoMigration, applied map[int]mongoAppliedMigration) error {
	known := make(map[int]int, len(items))
	for index, item := range items {
		known[item.Version] = index
	}

	highestAppliedIndex := -1
	for version, state := range applied {
		index, ok := known[version]
		if !ok {
			return fmt.Errorf("MongoDB migration history contains unknown version %03d", version)
		}
		if state.Name != items[index].Name {
			return fmt.Errorf("MongoDB migration %03d filename changed after application: have %q, expected %q", version, state.Name, items[index].Name)
		}
		if state.Dirty {
			return fmt.Errorf("MongoDB migration %03d is dirty; restore or repair it before retrying", version)
		}
		if state.Checksum != items[index].Checksum {
			return fmt.Errorf("MongoDB migration %03d checksum changed after application", version)
		}
		if index > highestAppliedIndex {
			highestAppliedIndex = index
		}
	}
	for index := 0; index <= highestAppliedIndex; index++ {
		if _, ok := applied[items[index].Version]; !ok {
			return fmt.Errorf("MongoDB migration history is out of order: version %03d is missing before an applied migration", items[index].Version)
		}
	}
	return nil
}
