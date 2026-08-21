package manage_service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateIndexes reconciles runtime indexes with the current multi-user schema.
func CreateIndexes() error {
	switch util.GetConfigByKey("db.type") {
	case "mongodb":
		return createMongoDBIndexes()
	case "mysql":
		return createMySQLIndexes()
	default:
		return fmt.Errorf("unsupported database type %q", util.GetConfigByKey("db.type"))
	}
}

func mongoCashFlowIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "belongs_user_id", Value: 1}, {Key: "belongs_date", Value: -1}}, Options: options.Index().SetName("cash_flows_user_date_index")},
		{Keys: bson.D{{Key: "belongs_user_id", Value: 1}, {Key: "category_id", Value: 1}}, Options: options.Index().SetName("cash_flows_user_category_index")},
		{Keys: bson.D{{Key: "belongs_user_id", Value: 1}, {Key: "is_delete", Value: 1}}, Options: options.Index().SetName("cash_flows_user_active_index")},
	}
}

func mongoCategoryIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "belongs_user_id", Value: 1}, {Key: "type", Value: 1}, {Key: "parent_id", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().SetName("categories_active_scope_unique_index").SetUnique(true).
				SetPartialFilterExpression(bson.D{{Key: "is_delete", Value: false}}),
		},
		{Keys: bson.D{{Key: "belongs_user_id", Value: 1}, {Key: "type", Value: 1}, {Key: "is_delete", Value: 1}}, Options: options.Index().SetName("categories_user_type_active_index")},
		{Keys: bson.D{{Key: "belongs_user_id", Value: 1}, {Key: "parent_id", Value: 1}, {Key: "is_delete", Value: 1}}, Options: options.Index().SetName("categories_user_parent_active_index")},
	}
}

func mongoBudgetIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.D{{Key: "belongs_user_id", Value: 1}, {Key: "period", Value: 1}, {Key: "category_id", Value: 1}}, Options: options.Index().SetName("budgets_active_scope_unique_index").SetUnique(true).SetPartialFilterExpression(bson.D{{Key: "is_delete", Value: false}})},
		{Keys: bson.D{{Key: "belongs_user_id", Value: 1}, {Key: "period", Value: 1}, {Key: "is_delete", Value: 1}}, Options: options.Index().SetName("budgets_user_period_active_index")},
	}
}

func createMongoDBIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cashFlows := database.GetMongoCollection(database.CashFlowTableName)
	if _, err := cashFlows.Indexes().CreateMany(ctx, mongoCashFlowIndexes()); err != nil {
		return fmt.Errorf("create cash-flow indexes: %w", err)
	}
	categories := database.GetMongoCollection(database.CategoryTableName)
	if _, err := categories.Indexes().CreateMany(ctx, mongoCategoryIndexes()); err != nil {
		return fmt.Errorf("create category indexes: %w", err)
	}
	budgets := database.GetMongoCollection(database.BudgetTableName)
	if _, err := budgets.Indexes().CreateMany(ctx, mongoBudgetIndexes()); err != nil {
		return fmt.Errorf("create budget indexes: %w", err)
	}
	if err := dropMongoIndexes(ctx, cashFlows, "idx_flow_type", "idx_belongs_date_flow_type", "flow_type_1", "belongs_date_-1_flow_type_1"); err != nil {
		return err
	}
	if err := dropMongoIndexes(ctx, categories, "idx_category_name_unique", "belongs_user_id_1_name_1", "name_1"); err != nil {
		return err
	}
	return nil
}

func dropMongoIndexes(ctx context.Context, collection *mongo.Collection, names ...string) error {
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return fmt.Errorf("list indexes for %s: %w", collection.Name(), err)
	}
	defer cursor.Close(ctx)

	existing := map[string]bool{}
	for cursor.Next(ctx) {
		var index struct {
			Name string `bson:"name"`
		}
		if err := cursor.Decode(&index); err != nil {
			return fmt.Errorf("decode index for %s: %w", collection.Name(), err)
		}
		existing[index.Name] = true
	}
	for _, name := range names {
		if existing[name] {
			if _, err := collection.Indexes().DropOne(ctx, name); err != nil {
				return fmt.Errorf("drop obsolete index %s.%s: %w", collection.Name(), name, err)
			}
		}
	}
	return cursor.Err()
}

type mysqlIndexDefinition struct {
	name    string
	table   string
	columns string
}

func mysqlIndexes() []mysqlIndexDefinition {
	return []mysqlIndexDefinition{
		{name: "cash_flows_user_date_index", table: database.CashFlowTableName, columns: "belongs_user_id, belongs_date"},
		{name: "cash_flows_user_category_index", table: database.CashFlowTableName, columns: "belongs_user_id, category_id"},
		{name: "cash_flows_user_active_index", table: database.CashFlowTableName, columns: "belongs_user_id, is_delete"},
		{name: "categories_user_scope_index", table: database.CategoryTableName, columns: "belongs_user_id, type, parent_id, name, is_delete"},
		{name: "budgets_user_period_active_index", table: database.BudgetTableName, columns: "belongs_user_id, period, is_delete"},
	}
}

func createMySQLIndexes() error {
	db := database.GetMySqlConnection()
	for _, index := range mysqlIndexes() {
		exists, err := mysqlIndexExists(db, index.table, index.name)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		query := fmt.Sprintf("CREATE INDEX `%s` ON `%s` (%s)", index.name, index.table, index.columns)
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("create MySQL index %s: %w", index.name, err)
		}
	}
	return nil
}

func mysqlIndexExists(db *sql.DB, table, index string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?`, table, index).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query MySQL index %s: %w", index, err)
	}
	return count > 0, nil
}
