# CashLenX API - Quick Start

**Get the API running in 5 minutes**

---

## 🚀 Fast Track (Copy & Paste)

### 1. Start MongoDB
```bash
cd cashlenx-server
docker compose up -d mongodb
```

### 2. Build & Configure
```bash
# Build the binary
go build -o cashlenx main.go

# Set environment variables
export DB_TYPE=mongodb
export MONGO_DB_URI="mongodb://cashlenx:cashlenx123@localhost:27017/cashlenx?authSource=admin"
export DB_NAME=cashlenx
export JWT_SECRET="your-secret-key"
```

### 3. Start API Server
```bash
./cashlenx open start -p 8080
```

### 4. Test API (in new terminal)
```bash
# Check health
curl http://localhost:8080/api/open/health

# Check version
curl http://localhost:8080/api/open/version
```

---

## 📋 Common Commands Reference

### Server Operations
```bash
# Start server
./cashlenx open start -p 8080

# Start in background
nohup ./cashlenx open start -p 8080 > server.log 2>&1 &

# Check server status
curl http://localhost:8080/api/open/health
```

### Database Operations
```bash
# Create backup (requires admin token)
export ADMIN_TOKEN="your-admin-token"
./cashlenx admin backup -o backup.json
```

### Authentication (Required for most operations)
```bash
# Register a new user
curl -X POST http://localhost:8080/api/open/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user1", "password":"password123", "email":"user1@example.com"}'

# Login to get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/open/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user1", "password":"password123"}' | jq -r .data.token)

echo "Token: $TOKEN"
```

### Category Management (Requires Token)
```bash
# Create category
curl -X POST http://localhost:8080/api/category \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Food", "type": "expense", "remark": "Food expenses"}'

# List all categories
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/category

# Get category by name
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/category/name/Food
```

### Transaction Management (Requires Token)
```bash
# Create expense
curl -X POST http://localhost:8080/api/cash/expense \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "belongs_date": "2025-12-12",
    "category_name": "Food",
    "amount": 45.50,
    "description": "Lunch"
  }'

# Create income
curl -X POST http://localhost:8080/api/cash/income \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "belongs_date": "2025-12-12",
    "category_name": "Salary",
    "amount": 5000.00,
    "description": "Monthly salary"
  }'

# List all transactions
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/cash?limit=20&offset=0"

# Get transaction by ID
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/cash/{id}
```

### Analytics & Reports (Requires Token)
```bash
# Daily summary
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/cash/summary/monthly/202512

# Statistic Export
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/statistic/export?format=xlsx" > export.xlsx
```

### CLI Alternatives
```bash
# NOTE: CLI commands also need environment variables set (DB_TYPE, MONGO_DB_URI, etc.)
# and currently operate directly on the database.

# Category operations
./cashlenx category create -n "Food" -r "Food expenses"
./cashlenx category query -n "Food"
./cashlenx category list

# Cash flow operations
./cashlenx cash income -c "Salary" -a 5000 -d "2025-12-12" -d "Salary"
./cashlenx cash expense -c "Food" -a 45.50 -d "2025-12-12" -d "Lunch"
./cashlenx cash query -b "2025-12-12"
./cashlenx cash summary -p daily -d "2025-12-12"

# Statistic operations
./cashlenx statistic export -o "backup.xlsx"
./cashlenx statistic import -i "backup.xlsx"
```

---

## 🧪 Quick Test Script

Save as `test_api.sh`:

```bash
#!/bin/bash

API="http://localhost:8080/api"
USERNAME="testuser_$(date +%s)"

echo "Testing CashLenX API..."
echo ""

# Health check
echo "1. Health Check"
curl -s $API/open/health | jq .
echo ""

# Register
echo "2. Register User"
REGISTER_RES=$(curl -s -X POST $API/open/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\", \"password\":\"password123\", \"email\":\"$USERNAME@example.com\"}")
echo $REGISTER_RES | jq .
echo ""

# Login
echo "3. Login"
LOGIN_RES=$(curl -s -X POST $API/open/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\", \"password\":\"password123\"}")
TOKEN=$(echo $LOGIN_RES | jq -r .data.token)
echo "Token obtained"
echo ""

# Create category
echo "4. Create Category"
CATEGORY=$(curl -s -X POST $API/category \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"TestFood","remark":"Test"}')
CATEGORY_ID=$(echo $CATEGORY | jq -r .data.id)
echo "Category ID: $CATEGORY_ID"
echo ""

# Create expense
echo "5. Create Expense"
EXPENSE=$(curl -s -X POST $API/cash/expense \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "belongs_date":"2025-12-12",
    "category_name":"TestFood",
    "amount":25.00,
    "description":"Test expense"
  }')
EXPENSE_ID=$(echo $EXPENSE | jq -r .data.id)
echo "Expense ID: $EXPENSE_ID"
echo ""

# List transactions
echo "6. List Transactions"
curl -s -H "Authorization: Bearer $TOKEN" "$API/cash?limit=5" | jq '.data | length'
echo " transactions found"
echo ""

# Cleanup
echo "7. Cleanup"
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" $API/cash/$EXPENSE_ID > /dev/null
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" $API/category/$CATEGORY_ID > /dev/null
echo "Test data cleaned up"
echo ""

echo "✅ All tests passed!"
```

Run it:
```bash
chmod +x test_api.sh
./test_api.sh
```

---

## 📚 Full Documentation

- **Deployment Guide**: `docs/deployment_guide.md` (detailed step-by-step)
- **API Reference**: `docs/api.md`
- **CLI Reference**: `docs/cli.md`
- **Roadmap**: `docs/roadmap.md`
- **Testing Guide**: `docs/testing.md`
