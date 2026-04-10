// MongoDB initialization script for CashLenX - SCHEMA ONLY
// This script creates the database, collections, and inserts basic default categories
// Demo/test data is available in init-mongo-demo.js (import manually via CLI: cashlenx manage import)

print('Starting MongoDB initialization for CashLenX...');

// Resolve database name from the Mongo container initialization environment.
// In compose.yml, MONGO_INITDB_DATABASE is sourced from DB_NAME, so DB_NAME remains
// the single user-facing configuration field in .env.
const dbName = process.env.MONGO_INITDB_DATABASE || 'cashlenx';

// Switch to target database
db = db.getSiblingDB(dbName);

// Create collections
db.createCollection('users');
db.createCollection('auth_token_refresh');
db.createCollection('auth_token_password_reset');
db.createCollection('cash_flows');
db.createCollection('categories');

print('Collections created successfully');

// Note: Default categories are automatically created for each user when they register
// See config/default_categories.json for the list of default categories

// Create indexes for better query performance
// Users indexes
db.users.createIndex({ username: 1 }, { unique: true });
db.users.createIndex({ role: 1 });
db.users.createIndex({ is_delete: 1 });

// Auth token refresh indexes
db.auth_token_refresh.createIndex({ token: 1 }, { unique: true });
db.auth_token_refresh.createIndex({ user_id: 1 });
db.auth_token_refresh.createIndex({ expires_at: 1 });
db.auth_token_refresh.createIndex({ is_delete: 1 });
db.auth_token_refresh.createIndex({ create_time: 1 });

// Auth token password reset indexes
db.auth_token_password_reset.createIndex({ token: 1 }, { unique: true });
db.auth_token_password_reset.createIndex({ user_id: 1 });
db.auth_token_password_reset.createIndex({ expires_at: 1 });
db.auth_token_password_reset.createIndex({ is_delete: 1 });
db.auth_token_password_reset.createIndex({ create_time: 1 });
db.auth_token_password_reset.createIndex({ ip_address: 1 });

// Cash flows indexes
db.cash_flows.createIndex({ belongs_date: -1 });
db.cash_flows.createIndex({ category_id: 1 });
db.cash_flows.createIndex({ flow_type: 1 });
db.cash_flows.createIndex({ belongs_date: -1, flow_type: 1 });
db.cash_flows.createIndex({ belongs_user_id: 1 });

// Categories indexes
db.categories.createIndex({ belongs_user_id: 1 });
db.categories.createIndex({ belongs_user_id: 1, name: 1 }, { unique: true });

print('Indexes created successfully');

// Print initialization summary
print('\n=== CashLenX MongoDB Initialized ===');
print(`Database: ${dbName}`);
print(`Users: ${db.users.countDocuments()}`);
print(`Auth Token Refresh: ${db.auth_token_refresh.countDocuments()}`);
print(`Auth Token Password Reset: ${db.auth_token_password_reset.countDocuments()}`);
print(`Categories: ${db.categories.countDocuments()}`);
print(`Cash Flows: ${db.cash_flows.countDocuments()}`);
print('');
print('Schema only - no demo data loaded');
print('Admin user will be auto-created on first server start');
print('Default categories will be auto-created for each new user');
print('Load demo data via: cashlenx manage import -i demo-data.xlsx');
print('=====================================\n');

print('MongoDB initialization completed successfully!');
