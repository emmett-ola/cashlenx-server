# MongoDB Initialization Files

## Overview

This directory contains MongoDB initialization scripts for CashLenX. The scripts are separated into schema initialization and optional demo data to provide flexibility for different deployment scenarios.

## Files

### 1. `init-mongo.js` (Primary - Auto-Executed)
**Purpose**: Schema initialization with basic default categories

**Executed Automatically On**: `docker-compose up` (first run)

**What It Does**:
- Creates `cash_flows` collection
- Creates `categories` collection
- Inserts 10 default categories (Salary, Freelance, Investment, Food & Dining, Transportation, Shopping, Entertainment, Healthcare, Utilities, Other Income)
- Creates indexes for query performance
- No transaction/demo data (clean slate)

**Data Model Alignment**:
- Matches Go `CategoryEntity` model with optional `parent_id` for hierarchical categories
- Matches Go `CashFlowEntity` model with proper field names (`category_id`, `belongs_date`, `flow_type`, `remark`)
- Uses `create_time` and `modify_time` instead of `created_at`/`updated_at`

### 2. `init-mongo-demo.js` (Optional - Manual Import)
**Purpose**: Demo transaction data for testing and development

**When to Use**:
- Development environment testing
- Demo/POC deployments
- Testing queries and analytics

**How to Load**:

**Option A: Manual database import**
```bash
docker exec -i cashlenx-mongodb mongosh -u cashlenx -p cashlenx123 \
  --authenticationDatabase admin cashlenx < docker/mongodb/init-mongo-demo.js
```

**Option B: Via CLI (Recommended)**
1. Convert demo data to Excel format
2. Run: `cashlenx manage import -i demo-data.xlsx`

**Data Included**:
- 15 sample transactions spread across the past month
- Mix of income and expense transactions
- Realistic amounts and categories
- Includes today, yesterday, this week, and earlier month data

## Schema Design

### Categories Collection
```javascript
{
  _id: ObjectId(),              // MongoDB ID
  parent_id: ObjectId(),        // Optional: hierarchical structure
  name: String,                 // Category name
  remark: String,               // Additional notes
  create_time: Date,            // Creation timestamp
  modify_time: Date             // Last modified timestamp
}
```

**Default Categories**:
- **Income**: Salary, Freelance, Investment, Other Income
- **Expense**: Food & Dining, Transportation, Shopping, Entertainment, Healthcare, Utilities

### Cash Flows Collection
```javascript
{
  _id: ObjectId(),              // MongoDB ID
  category_id: ObjectId(),      // Reference to categories
  belongs_date: String,         // Date string (YYYY-MM-DD format)
  flow_type: String,            // 'income' or 'expense'
  amount: Number,               // Transaction amount
  description: String,          // Transaction description
  remark: String,               // Additional notes
  create_time: Date,            // Creation timestamp
  modify_time: Date             // Last modified timestamp
}
```

### Indexes
- `belongs_date: -1` - Query by date (descending for latest first)
- `category_id: 1` - Filter by category
- `flow_type: 1` - Filter by income/expense
- `belongs_date: -1, flow_type: 1` - Common combined query

## Improvements Over Original

### 1. Entity Model Alignment
- **Before**: Used `date`, `type`, `category` (string reference)
- **After**: Uses `belongs_date`, `flow_type`, `category_id` (ObjectId reference)

### 2. Better Category References
- **Before**: String-based category names
- **After**: ObjectId references for data integrity and query performance

### 3. Hierarchical Categories
- **Before**: Flat category structure
- **After**: `parent_id` field enables category hierarchies

### 4. Field Name Consistency
- **Before**: `created_at`
- **After**: `create_time`, `modify_time` (matches Go models)

### 5. Separated Demo Data
- **Before**: Schema and demo data mixed
- **After**: Clean separation for production readiness

### 6. Enhanced Indexes
- Better index strategy for common query patterns
- Index on hierarchical parent_id lookups (for future use)

### 7. Better Documentation
- Clear comments explaining field purposes
- Schema alignment documentation

## Usage Examples

### Start with Schema Only (Default)
```bash
# Docker Compose automatically uses init-mongo.js
docker-compose up -d mongodb

# Verify collections and indexes
docker exec -it cashlenx-mongodb mongosh -u cashlenx -p cashlenx123 --authenticationDatabase admin cashlenx
# Then: db.getCollectionNames()
```

### Load Demo Data Manually
```bash
# Option 1: Direct MongoDB script
docker exec -i cashlenx-mongodb mongosh -u cashlenx -p cashlenx123 \
  --authenticationDatabase admin cashlenx < docker/mongodb/init-mongo-demo.js

# Option 2: Via CLI (after building cashlenx)
cd backend
go build -o cashlenx main.go

# Create Excel from demo JS (manual step) or use CLI:
./cashlenx manage import -i path/to/demo-data.xlsx
```

### Query Examples
```bash
# View categories
docker exec -it cashlenx-mongodb mongosh -u cashlenx -p cashlenx123 \
  --authenticationDatabase admin cashlenx \
  -e "db.categories.find({}, {name: 1, _id: 1}).pretty()"

# View transactions for today
docker exec -it cashlenx-mongodb mongosh -u cashlenx -p cashlenx123 \
  --authenticationDatabase admin cashlenx \
  -e "db.cash_flows.find({belongs_date: new Date().toISOString().split('T')[0]}).pretty()"

# Summary statistics
docker exec -it cashlenx-mongodb mongosh -u cashlenx -p cashlenx123 \
  --authenticationDatabase admin cashlenx \
  -e "db.cash_flows.aggregate([
    { \$group: { _id: '\$flow_type', total: { \$sum: '\$amount' }, count: { \$sum: 1 } } }
  ]).pretty()"
```

## Migration from Old Schema

If migrating from old schema with different field names:

```javascript
// Rename fields in cash_flows
db.cash_flows.updateMany(
  {},
  [
    {
      $set: {
        belongs_date: '$date',
        flow_type: '$type',
        create_time: '$created_at',
        modify_time: '$updated_at'
      }
    },
    {
      $unset: ['date', 'type', 'created_at', 'updated_at']
    }
  ]
);

// Add category_id from category name (requires joining logic)
db.categories.find({}).forEach(cat => {
  const categoryName = cat.name;
  db.cash_flows.updateMany(
    { category: categoryName },
    {
      $set: { category_id: cat._id },
      $unset: { category: 1 }
    }
  );
});
```

## Production Considerations

1. **Backup volumes** before production:
   ```bash
   docker run --rm -v cashlenx_mongodb_data:/data -v $(pwd):/backup ubuntu \
     tar czf /backup/mongodb-backup.tar.gz /data
   ```

2. **Use strong passwords** instead of defaults:
   ```yaml
   environment:
     MONGO_INITDB_ROOT_PASSWORD: <strong-password>
   ```

3. **Enable authentication** (already enabled by default):
   ```yaml
   environment:
     MONGO_INITDB_ROOT_USERNAME: <username>
     MONGO_INITDB_ROOT_PASSWORD: <strong-password>
   ```

4. **Configure backups** in cron or Docker backup service

5. **Monitor** indexes and query performance

6. **Use replica sets** for production (configure in docker-compose)

## Related Files

- Docker Compose: `../../docker-compose.yml`
- MySQL init: `../mysql/init-mysql.sql`
- Backend models: `../../backend/model/`
- CLI commands: `../../backend/cmd/manage_cmd/`

## See Also

- [MySQL README](../mysql/README.md) - MySQL initialization guide
- [Docker Guide](../../docs/DOCKER.md) - Docker setup and usage
- [Environment Guide](../../docs/ENVIRONMENT.md) - Configuration guide
