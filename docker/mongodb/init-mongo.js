// MongoDB initialization script for CashLenX - SCHEMA ONLY
// This script creates the database, collections, and inserts basic default categories
// Demo/test data is available in init-mongo-demo.js (import manually via CLI: cashlenx manage import)

print('Starting MongoDB initialization for CashLenX...');

// Switch to cashlenx database
db = db.getSiblingDB('cashlenx');

// Create collections
db.createCollection('cash_flows');
db.createCollection('categories');

print('Collections created successfully');

// Insert basic default categories (auto-loaded on init)
// Aligned with Go CategoryEntity model:
// - _id: MongoDB ObjectId
// - belongs_user_id: Owner user ID (system default)
// - parent_id: for hierarchical categories (optional)
// - name: category name
// - type: INCOME or EXPENSE
// - remark: additional notes
// - create_time: creation timestamp
// - modify_time: last modification timestamp
// - is_delete: soft delete flag
const categories = [
  {
    _id: ObjectId(),
    belongs_user_id: ObjectId("000000000000000000000000"),
    name: 'Salary',
    type: 'INCOME',
    remark: 'Income from employment',
    create_user_id: ObjectId("000000000000000000000000"),
    create_time: new Date(),
    update_user_id: ObjectId("000000000000000000000000"),
    modify_time: new Date(),
    is_delete: false
  },
  {
    _id: ObjectId(),
    belongs_user_id: ObjectId("000000000000000000000000"),
    name: 'Freelance',
    type: 'INCOME',
    remark: 'Income from freelance work',
    create_user_id: ObjectId("000000000000000000000000"),
    create_time: new Date(),
    update_user_id: ObjectId("000000000000000000000000"),
    modify_time: new Date(),
    is_delete: false
  },
  {
    _id: ObjectId(),
    belongs_user_id: ObjectId("000000000000000000000000"),
    name: 'Investment',
    type: 'INCOME',
    remark: 'Income from investments and dividends',
    create_user_id: ObjectId("000000000000000000000000"),
    create_time: new Date(),
    update_user_id: ObjectId("000000000000000000000000"),
    modify_time: new Date(),
    is_delete: false
  },
  {
    _id: ObjectId(),
    belongs_user_id: ObjectId("000000000000000000000000"),
    name: 'Other Income',
    type: 'INCOME',
    remark: 'Other income sources',
    create_user_id: ObjectId("000000000000000000000000"),
    create_time: new Date(),
    update_user_id: ObjectId("000000000000000000000000"),
    modify_time: new Date(),
    is_delete: false
  },
  {
    _id: ObjectId(),
    belongs_user_id: ObjectId("000000000000000000000000"),
    name: 'Food & Dining',
    type: 'EXPENSE',
    remark: 'Restaurants, groceries, food delivery',
    create_user_id: ObjectId("000000000000000000000000"),
    create_time: new Date(),
    update_user_id: ObjectId("000000000000000000000000"),
    modify_time: new Date(),
    is_delete: false
  },
  {
    _id: ObjectId(),
    belongs_user_id: ObjectId("000000000000000000000000"),
    name: 'Transportation',
    type: 'EXPENSE',
    remark: 'Gas, public transport, car maintenance',
    create_user_id: ObjectId("000000000000000000000000"),
    create_time: new Date(),
    update_user_id: ObjectId("000000000000000000000000"),
    modify_time: new Date(),
    is_delete: false
  },
  {
    _id: ObjectId(),
    belongs_user_id: ObjectId("000000000000000000000000"),
    name: 'Shopping',
    type: 'EXPENSE',
    remark: 'Retail purchases, online shopping',
    create_user_id: ObjectId("000000000000000000000000"),
    create_time: new Date(),
    update_user_id: ObjectId("000000000000000000000000"),
    modify_time: new Date(),
    is_delete: false
  },
  {
    _id: ObjectId(),
    belongs_user_id: ObjectId("000000000000000000000000"),
    name: 'Entertainment',
    type: 'EXPENSE',
    remark: 'Movies, games, hobbies',
    create_user_id: ObjectId("000000000000000000000000"),
    create_time: new Date(),
    update_user_id: ObjectId("000000000000000000000000"),
    modify_time: new Date(),
    is_delete: false
  },
  {
    _id: ObjectId(),
    belongs_user_id: ObjectId("000000000000000000000000"),
    name: 'Healthcare',
    type: 'EXPENSE',
    remark: 'Medical expenses, pharmacy, fitness',
    create_user_id: ObjectId("000000000000000000000000"),
    create_time: new Date(),
    update_user_id: ObjectId("000000000000000000000000"),
    modify_time: new Date(),
    is_delete: false
  },
  {
    _id: ObjectId(),
    belongs_user_id: ObjectId("000000000000000000000000"),
    name: 'Utilities',
    type: 'EXPENSE',
    remark: 'Electricity, water, internet, phone',
    create_user_id: ObjectId("000000000000000000000000"),
    create_time: new Date(),
    update_user_id: ObjectId("000000000000000000000000"),
    modify_time: new Date(),
    is_delete: false
  }
];

db.categories.insertMany(categories);
print(`Inserted ${categories.length} default categories`);

// Create indexes for better query performance
db.cash_flows.createIndex({ belongs_date: -1 });
db.cash_flows.createIndex({ category_id: 1 });
db.cash_flows.createIndex({ flow_type: 1 });
db.cash_flows.createIndex({ belongs_date: -1, flow_type: 1 });
db.cash_flows.createIndex({ belongs_user_id: 1 });
db.categories.createIndex({ belongs_user_id: 1 });
db.categories.createIndex({ belongs_user_id: 1, name: 1 }, { unique: true });

print('Indexes created successfully');

// Print initialization summary
print('\n=== CashLenX MongoDB Initialized ===');
print(`Categories: ${db.categories.countDocuments()}`);
print(`Cash Flows: ${db.cash_flows.countDocuments()}`);
print('Schema only - no demo data loaded');
print('Load demo data via: cashlenx manage import -i demo-data.xlsx');
print('=====================================\n');

print('MongoDB initialization completed successfully!');
