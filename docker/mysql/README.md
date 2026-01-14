# MySQL Initialization Files

## Overview

This directory contains MySQL initialization scripts for CashLenX. The scripts are separated into schema initialization and optional demo data to provide flexibility for different deployment scenarios.

## Files

### 1. `init-mysql.sql` (Primary - Auto-Executed)
**Purpose**: Schema initialization with basic default categories

**Executed Automatically On**: `docker-compose up` (first run)

**What It Does**:
- Creates `categories` table with foreign key support
- Creates `cash_flows` table with proper indexes for performance
- Inserts 10 default categories (Salary, Freelance, Investment, Food & Dining, Transportation, Shopping, Entertainment, Healthcare, Utilities, Other Income)
- No transaction/demo data (clean slate)

**Data Model Alignment**:
- Matches Go `CategoryEntity` model with `parent_id` for hierarchical categories
- Matches Go `CashFlowEntity` model with proper field names (`category_id`, `belongs_date`, `flow_type`, `remark`)
- Uses `DECIMAL(19, 4)` for precision and safety
- Foreign key constraints for data integrity

### 2. `init-mysql-demo.sql` (Optional - Manual Import)
**Purpose**: Demo transaction data for testing and development

**When to Use**:
- Development environment testing
- Demo/POC deployments
- Testing queries and analytics

**How to Load**:

**Option A: Manual database import**
```bash
mysql -u cashlenx -p cashlenx123 cashlenx < docker/mysql/init-mysql-demo.sql
```

**Option B: Via CLI (Recommended)**
1. Convert demo data to Excel format
2. Run: `cashlenx manage import -i demo-data.xlsx`

**Data Included**:
- 15 sample transactions spread across the past month
- Mix of income and expense transactions
- Realistic amounts and categories
- Includes today, yesterday, this week, and earlier month data

### 3. `init-mysql-schema.sql` (Backup - Mirror of init-mysql.sql)
**Purpose**: Backup/reference of schema-only initialization

**Note**: This is identical to `init-mysql.sql` for redundancy. Use `init-mysql.sql` in docker-compose.

## Schema Design

### Categories Table
```sql
CREATE TABLE categories (
    id VARCHAR(36) PRIMARY KEY,              -- UUID
    parent_id VARCHAR(36),                   -- Hierarchical structure support
    name VARCHAR(100) NOT NULL,              -- Category name
    remark TEXT,                             -- Additional notes
    create_time TIMESTAMP,                   -- Creation time
    modify_time TIMESTAMP,                   -- Last modified
    FOREIGN KEY (parent_id) REFERENCES categories(id)
);
```

**Default Categories**:
- **Income**: Salary, Freelance, Investment, Other Income
- **Expense**: Food & Dining, Transportation, Shopping, Entertainment, Healthcare, Utilities

### Cash Flows Table
```sql
CREATE TABLE cash_flows (
    id VARCHAR(36) PRIMARY KEY,              -- UUID
    category_id VARCHAR(36) NOT NULL,        -- FK to categories
    belongs_date DATE NOT NULL,              -- Transaction date
    flow_type VARCHAR(20) NOT NULL,          -- 'income' or 'expense'
    amount DECIMAL(19, 4) NOT NULL,          -- Precise decimal (4 places)
    description TEXT,                        -- Transaction description
    remark TEXT,                             -- Additional notes
    create_time TIMESTAMP,                   -- Creation time
    modify_time TIMESTAMP,                   -- Last modified
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
```

### Indexes
- `idx_belongs_date` - Query by date
- `idx_flow_type` - Filter by income/expense
- `idx_category_id` - Filter by category
- `idx_date_type` - Common combined query
- `idx_date_category` - Date + category filtering

## Improvements Over Original

### 1. Entity Model Alignment
- **Before**: Used `date`, `type`, `category` (string)
- **After**: Uses `belongs_date`, `flow_type`, `category_id` (FK)

### 2. Better Foreign Keys
- **Before**: No foreign key relationships
- **After**: `category_id` references `categories.id` with proper constraints

### 3. Hierarchical Categories
- **Before**: Flat category structure
- **After**: `parent_id` field enables category hierarchies

### 4. Field Name Consistency
- **Before**: `created_at`, `updated_at`
- **After**: `create_time`, `modify_time` (matches Go models)

### 5. Improved Precision
- **Before**: `DECIMAL(15, 2)`
- **After**: `DECIMAL(19, 4)` (better range and precision)

### 6. Separated Demo Data
- **Before**: Schema and demo data mixed
- **After**: Clean separation for production readiness

### 7. Enhanced Indexes
- Composite indexes for common query patterns
- Index on hierarchical parent_id lookups

### 8. Better Documentation
- Column comments explaining purpose
- Table comments
- Clear initialization instructions

## Usage Examples

### Start with Schema Only (Default)
```bash
# Docker Compose automatically uses init-mysql.sql
docker-compose --profile mysql up -d mysql

# Verify schema
docker exec -it cashlenx-mysql mysql -u cashlenx -p cashlenx123 cashlenx -e "SHOW TABLES;"
```

### Load Demo Data Manually
```bash
# Option 1: Direct SQL
docker exec -i cashlenx-mysql mysql -u cashlenx -p cashlenx123 cashlenx < docker/mysql/init-mysql-demo.sql

# Option 2: Via CLI (after building cashlenx)
cd backend
go build -o cashlenx main.go

# Create Excel from demo SQL (manual step) or use CLI:
./cashlenx manage import -i path/to/demo-data.xlsx
```

### Query Examples
```bash
# View categories
docker exec -it cashlenx-mysql mysql -u cashlenx -p cashlenx123 cashlenx \
  -e "SELECT id, name, parent_id FROM categories;"

# View transactions for today
docker exec -it cashlenx-mysql mysql -u cashlenx -p cashlenx123 cashlenx \
  -e "SELECT * FROM cash_flows WHERE belongs_date = CURDATE();"

# Summary statistics
docker exec -it cashlenx-mysql mysql -u cashlenx -p cashlenx123 cashlenx \
  -e "SELECT 
    flow_type, 
    COUNT(*) as count, 
    SUM(amount) as total 
  FROM cash_flows 
  GROUP BY flow_type;"
```

## Migration Path

If migrating from old schema to new:

```sql
-- Add parent_id column to categories
ALTER TABLE categories ADD COLUMN parent_id VARCHAR(36);
ALTER TABLE categories ADD FOREIGN KEY (parent_id) REFERENCES categories(id);

-- Rename fields in cash_flows
ALTER TABLE cash_flows 
  CHANGE COLUMN date belongs_date DATE,
  CHANGE COLUMN type flow_type VARCHAR(20),
  ADD COLUMN category_id VARCHAR(36),
  CHANGE COLUMN created_at create_time TIMESTAMP,
  CHANGE COLUMN updated_at modify_time TIMESTAMP;

-- Populate category_id from category names
UPDATE cash_flows cf 
SET cf.category_id = (SELECT id FROM categories WHERE name = cf.category)
WHERE cf.category IS NOT NULL;

-- Add foreign key constraint
ALTER TABLE cash_flows 
  ADD FOREIGN KEY (category_id) REFERENCES categories(id);
```

## Production Considerations

1. **Backup volumes** before production:
   ```bash
docker run --rm -v cashlenx_mysql_data:/data -v $(pwd):/backup ubuntu \
     tar czf /backup/mysql-backup.tar.gz /data
   ```

2. **Use strong passwords** instead of defaults:
   ```yaml
   environment:
     MYSQL_PASSWORD: <strong-password>
   ```

3. **Configure backups** in cron or Docker backup service

4. **Monitor** indexes and query performance

## Related Files

- Docker Compose: `../../docker-compose.yml`
- MongoDB init: `../mongodb/init-mongo.js`
- Backend models: `../../backend/model/`
- CLI commands: `../../backend/cmd/manage_cmd/`
