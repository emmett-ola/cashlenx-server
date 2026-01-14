package manage_service

import (
	"context"
	"time"

	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateIndexes creates database indexes for performance optimization
func CreateIndexes() error {
	dbType := util.GetConfigByKey("db.type")

	switch dbType {
	case "mongodb":
		return createMongoDBIndexes()
	case "mysql":
		return createMySQLIndexes()
	default:
		util.Logger.Errorw("unsupported database type", "type", dbType)
		return nil
	}
}

func createMongoDBIndexes() error {
	util.Logger.Info("Creating MongoDB indexes...")

	database.OpenMongoDbConnection(database.CashFlowTableName)
	defer database.CloseMongoDbConnection()

	collection := database.GetMongoDbCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Index on belongs_date
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "belongs_date", Value: 1}},
		Options: options.Index().SetName("idx_belongs_date"),
	})
	if err != nil {
		util.Logger.Errorw("failed to create belongs_date index", "error", err)
		return err
	}
	util.Logger.Info("✓ Created index: idx_belongs_date")

	// Index on flow_type
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "flow_type", Value: 1}},
		Options: options.Index().SetName("idx_flow_type"),
	})
	if err != nil {
		util.Logger.Errorw("failed to create flow_type index", "error", err)
		return err
	}
	util.Logger.Info("✓ Created index: idx_flow_type")

	// Compound index on belongs_date and flow_type
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "belongs_date", Value: 1},
			{Key: "flow_type", Value: 1},
		},
		Options: options.Index().SetName("idx_belongs_date_flow_type"),
	})
	if err != nil {
		util.Logger.Errorw("failed to create compound index", "error", err)
		return err
	}
	util.Logger.Info("✓ Created index: idx_belongs_date_flow_type")

	// Index on category_id
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "category_id", Value: 1}},
		Options: options.Index().SetName("idx_category_id"),
	})
	if err != nil {
		util.Logger.Errorw("failed to create category_id index", "error", err)
		return err
	}
	util.Logger.Info("✓ Created index: idx_category_id")

	// Category collection - unique index on name
	database.CloseMongoDbConnection()
	database.OpenMongoDbConnection(database.CategoryTableName)
	defer database.CloseMongoDbConnection()

	categoryCollection := database.GetMongoDbCollection()
	_, err = categoryCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetName("idx_category_name_unique").SetUnique(true),
	})
	if err != nil {
		util.Logger.Errorw("failed to create category name index", "error", err)
		return err
	}
	util.Logger.Info("✓ Created unique index: idx_category_name_unique")

	util.Logger.Info("All indexes created successfully")
	return nil
}

func createMySQLIndexes() error {
	util.Logger.Info("Creating MySQL indexes...")

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Index on belongs_date
	_, err := connection.Exec("CREATE INDEX IF NOT EXISTS idx_belongs_date ON cash_flow(BELONGS_DATE)")
	if err != nil {
		util.Logger.Errorw("failed to create belongs_date index", "error", err)
		return err
	}
	util.Logger.Info("✓ Created index: idx_belongs_date")

	// Index on flow_type
	_, err = connection.Exec("CREATE INDEX IF NOT EXISTS idx_flow_type ON cash_flow(FLOW_TYPE)")
	if err != nil {
		util.Logger.Errorw("failed to create flow_type index", "error", err)
		return err
	}
	util.Logger.Info("✓ Created index: idx_flow_type")

	// Compound index on belongs_date and flow_type
	_, err = connection.Exec("CREATE INDEX IF NOT EXISTS idx_belongs_date_flow_type ON cash_flow(BELONGS_DATE, FLOW_TYPE)")
	if err != nil {
		util.Logger.Errorw("failed to create compound index", "error", err)
		return err
	}
	util.Logger.Info("✓ Created index: idx_belongs_date_flow_type")

	// Index on category_id
	_, err = connection.Exec("CREATE INDEX IF NOT EXISTS idx_category_id ON cash_flow(CATEGORY_ID)")
	if err != nil {
		util.Logger.Errorw("failed to create category_id index", "error", err)
		return err
	}
	util.Logger.Info("✓ Created index: idx_category_id")

	// Unique index on category name
	_, err = connection.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_category_name_unique ON category(NAME)")
	if err != nil {
		util.Logger.Errorw("failed to create category name index", "error", err)
		return err
	}
	util.Logger.Info("✓ Created unique index: idx_category_name_unique")

	util.Logger.Info("All indexes created successfully")
	return nil
}
