# Database Migrations

This directory contains database migration scripts for CashLenX.

## Available Migrations

### 001_add_indexes.js
Creates performance indexes on frequently queried fields.

**MongoDB Usage**:
```bash
mongosh <connection_string> < 001_add_indexes.js
```

**Or use CLI command**:
```bash
cashlenx manage indexes
```

**Indexes Created**:
- `cash_flow.belongs_date` - For date range queries
- `cash_flow.flow_type` - For income/expense filtering
- `cash_flow(belongs_date, flow_type)` - Compound index for filtered queries
- `cash_flow.category_id` - For category-based queries
- `category.name` - Unique index for category lookups

**Expected Performance Improvement**:
- Date queries: 10-100x faster
- Category lookups: 10x faster
- Type filtering: 50x faster

## Migration Guidelines

1. **Always backup** your database before running migrations
2. **Test migrations** on a copy of production data first
3. **Run migrations** during low-traffic periods
4. **Monitor performance** after migration
5. **Have a rollback plan** ready

## Rollback

To remove indexes created by 001_add_indexes.js:

```javascript
use cashlenx;
db.cash_flow.dropIndex("idx_belongs_date");
db.cash_flow.dropIndex("idx_flow_type");
db.cash_flow.dropIndex("idx_belongs_date_flow_type");
db.cash_flow.dropIndex("idx_category_id");
db.category.dropIndex("idx_category_name_unique");
```

## Future Migrations

Future migrations will be numbered sequentially:
- 002_*.js
- 003_*.js
- etc.

Always run migrations in order.
