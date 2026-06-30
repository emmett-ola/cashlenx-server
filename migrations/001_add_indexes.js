// MongoDB index reconciliation for the current multi-user schema.
// Run against the selected application database, for example:
// mongosh "$MONGO_DB_URI" migrations/001_add_indexes.js

print(`Reconciling indexes in database: ${db.getName()}`);

function dropIndexIfPresent(collection, indexName) {
  if (collection.getIndexes().some((index) => index.name === indexName)) {
    collection.dropIndex(indexName);
    print(`Dropped obsolete index: ${collection.getName()}.${indexName}`);
  }
}

// Remove legacy names from both the old singular collections and the current
// plural collections. Dropping an absent index is intentionally a no-op.
['cash_flow', 'cash_flows'].forEach((collectionName) => {
  const collection = db.getCollection(collectionName);
  dropIndexIfPresent(collection, 'idx_flow_type');
  dropIndexIfPresent(collection, 'idx_belongs_date_flow_type');
});

['category', 'categories'].forEach((collectionName) => {
  const collection = db.getCollection(collectionName);
  dropIndexIfPresent(collection, 'idx_category_name_unique');
  dropIndexIfPresent(collection, 'belongs_user_id_1_name_1');
});

db.cash_flows.createIndex(
  { belongs_user_id: 1, belongs_date: -1 },
  { name: 'cash_flows_user_date_index' }
);
db.cash_flows.createIndex(
  { belongs_user_id: 1, category_id: 1 },
  { name: 'cash_flows_user_category_index' }
);
db.cash_flows.createIndex(
  { belongs_user_id: 1, is_delete: 1 },
  { name: 'cash_flows_user_active_index' }
);

db.categories.createIndex(
  { belongs_user_id: 1, type: 1, parent_id: 1, name: 1 },
  {
    unique: true,
    name: 'categories_active_scope_unique_index',
    partialFilterExpression: { is_delete: false }
  }
);
db.categories.createIndex(
  { belongs_user_id: 1, type: 1, is_delete: 1 },
  { name: 'categories_user_type_active_index' }
);
db.categories.createIndex(
  { belongs_user_id: 1, parent_id: 1, is_delete: 1 },
  { name: 'categories_user_parent_active_index' }
);

db.user_configurations.createIndex(
  { belongs_user_id: 1 },
  { unique: true, name: 'user_configurations_belongs_user_id_unique_index' }
);

print('MongoDB index reconciliation complete.');
