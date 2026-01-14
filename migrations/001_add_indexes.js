// MongoDB Index Migration
// Run with: mongosh <connection_string> < 001_add_indexes.js

// Switch to cashlenx database
use cashlenx;

print("Creating indexes for cash_flow collection...");

// Index on belongs_date for date range queries
db.cash_flow.createIndex({ "belongs_date": 1 }, { name: "idx_belongs_date" });
print("✓ Created index: idx_belongs_date");

// Index on flow_type for income/expense filtering
db.cash_flow.createIndex({ "flow_type": 1 }, { name: "idx_flow_type" });
print("✓ Created index: idx_flow_type");

// Compound index on belongs_date and flow_type for filtered date range queries
db.cash_flow.createIndex({ "belongs_date": 1, "flow_type": 1 }, { name: "idx_belongs_date_flow_type" });
print("✓ Created index: idx_belongs_date_flow_type");

// Index on category_id for category-based queries
db.cash_flow.createIndex({ "category_id": 1 }, { name: "idx_category_id" });
print("✓ Created index: idx_category_id");

print("\nCreating indexes for category collection...");

// Unique index on category name
db.category.createIndex({ "name": 1 }, { unique: true, name: "idx_category_name_unique" });
print("✓ Created unique index: idx_category_name_unique");

print("\n=== Index Creation Complete ===");
print("\nVerifying indexes:");

print("\ncash_flow indexes:");
db.cash_flow.getIndexes().forEach(function(idx) {
    print("  - " + idx.name + ": " + JSON.stringify(idx.key));
});

print("\ncategory indexes:");
db.category.getIndexes().forEach(function(idx) {
    print("  - " + idx.name + ": " + JSON.stringify(idx.key));
});
