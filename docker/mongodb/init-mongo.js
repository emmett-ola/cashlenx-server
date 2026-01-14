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
// - parent_id: for hierarchical categories (optional)
// - name: category name
// - remark: additional notes
// - create_time: creation timestamp
// - modify_time: last modification timestamp
const categories = [
  {
    _id: ObjectId(),
    name: 'Salary',
    remark: 'Income from employment',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    name: 'Freelance',
    remark: 'Income from freelance work',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    name: 'Investment',
    remark: 'Income from investments and dividends',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    name: 'Other Income',
    remark: 'Other income sources',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    name: 'Food & Dining',
    remark: 'Restaurants, groceries, food delivery',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    name: 'Transportation',
    remark: 'Gas, public transport, car maintenance',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    name: 'Shopping',
    remark: 'Retail purchases, online shopping',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    name: 'Entertainment',
    remark: 'Movies, games, hobbies',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    name: 'Healthcare',
    remark: 'Medical expenses, pharmacy, fitness',
    create_time: new Date(),
    modify_time: new Date()
  },
  {
    _id: ObjectId(),
    name: 'Utilities',
    remark: 'Electricity, water, internet, phone',
    create_time: new Date(),
    modify_time: new Date()
  }
];

db.categories.insertMany(categories);
print(`Inserted ${categories.length} default categories`);

// Create indexes for better query performance
db.cash_flows.createIndex({ belongs_date: -1 });
db.cash_flows.createIndex({ category_id: 1 });
db.cash_flows.createIndex({ flow_type: 1 });
db.cash_flows.createIndex({ belongs_date: -1, flow_type: 1 });

print('Indexes created successfully');

// Print initialization summary
print('\n=== CashLenX MongoDB Initialized ===');
print(`Categories: ${db.categories.countDocuments()}`);
print(`Cash Flows: ${db.cash_flows.countDocuments()}`);
print('Schema only - no demo data loaded');
print('Load demo data via: cashlenx manage import -i demo-data.xlsx');
print('=====================================\n');

print('MongoDB initialization completed successfully!');
