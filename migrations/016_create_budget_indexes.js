// MongoDB budget index migration.
// Runtime execution is implemented by the Go migration handler; this immutable
// asset defines the ordered migration identity and checksum.

db.budgets.createIndex(
  { belongs_user_id: 1, period: 1, category_id: 1 },
  {
    unique: true,
    name: 'budgets_active_scope_unique_index',
    partialFilterExpression: { is_delete: false }
  }
);
db.budgets.createIndex(
  { belongs_user_id: 1, period: 1, is_delete: 1 },
  { name: 'budgets_user_period_active_index' }
);
