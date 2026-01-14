# CashLenX API - Local Docker Deployment & Testing Guide

**Quick guide to deploy and test the CashLenX API locally**

---

## 📋 Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ installed
- Terminal/Command line access
- curl or Postman for API testing

---

## 🚀 Step 1: Start MongoDB with Docker

### Option A: Using Docker Compose (Recommended)

```bash
# Navigate to project root
cd cashlenx-server

# Start MongoDB container
docker compose up -d mongodb

# Verify MongoDB is running
docker compose ps

# Expected output:
# NAME                 STATUS              PORTS
# cashlenx-mongodb     Up X seconds        0.0.0.0:27017->27017/tcp
```

### Option B: Using Docker CLI Directly

```bash
docker run -d \
  --name cashlenx-mongodb \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=cashlenx \
  -e MONGO_INITDB_ROOT_PASSWORD=cashlenx123 \
  -e MONGO_INITDB_DATABASE=cashlenx \
  mongo:7.0
```

### Verify MongoDB Connection

```bash
# Test connection (should return "ping: 1")
docker exec cashlenx-mongodb mongosh \
  --username cashlenx \
  --password cashlenx123 \
  --authenticationDatabase admin \
  --eval "db.adminCommand('ping')"
```

---

## 🔧 Step 2: Build the Backend

```bash
# Navigate to project root (if not already there)
cd cashlenx-server

# Download dependencies
go mod download

# Build the binary
go build -o cashlenx main.go

# Verify build
ls -lh cashlenx
```

**Expected**: You should see a `cashlenx` binary (size ~20-30 MB)

---

## ⚙️ Step 3: Configure Environment

### Set Environment Variables

```bash
# Database type (mongodb or mysql)
export DB_TYPE=mongodb

# MongoDB connection string
export MONGO_DB_URI="mongodb://cashlenx:cashlenx123@localhost:27017/cashlenx?authSource=admin"

# Database name
export DB_NAME=cashlenx

# JWT Secret (Required for authentication)
export JWT_SECRET="your-secret-key"

# Optional: Log file location
export LOG_FOLDER="./logs"
```

### Create a .env file (Alternative)

```bash
cat > .env <<'EOF'
DB_TYPE=mongodb
MONGO_DB_URI=mongodb://cashlenx:cashlenx123@localhost:27017/cashlenx?authSource=admin
DB_NAME=cashlenx
JWT_SECRET=your-secret-key
LOG_FOLDER=./logs
EOF

# Load environment
export $(cat .env | xargs)
```

---

## 🎯 Step 4: Seed Test Data (Optional)

You can use the CLI to create some initial data (this bypasses the API and auth).

```bash
# Create demo category
./cashlenx category create -n "Food" -t "expense" -r "Demo data"

# Create demo expense
./cashlenx cash expense -c "Food" -a 100 -d "Demo lunch"
```

---

## 🚀 Step 5: Start API Server

### Start the server

```bash
# Start on port 8080
./cashlenx open start -p 8080
```

**Expected output**:
```
API server is running on http://localhost:8080
```

### Keep server running in background (Alternative)

```bash
# Start in background
nohup ./cashlenx open start -p 8080 > server.log 2>&1 &

# Get process ID
echo $! > server.pid

# Check it's running
cat server.log
```

---

## ✅ Step 6: Verify API is Running

### Test 1: Health Check

```bash
curl http://localhost:8080/api/open/health
```

**Expected response**:
```json
{
  "code": "OK",
  "message": "",
  "data": {
    "status": "healthy",
    "service": "cashlenx-api",
    "message": "API is running"
  },
  ...
}
```

### Test 2: Version Info

```bash
curl http://localhost:8080/api/open/version
```

---

## 🧪 Step 7: Run Basic API Tests (With Auth)

### 1. Register User

```bash
curl -X POST http://localhost:8080/api/open/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser", "password":"password123", "email":"test@example.com"}'
```

### 2. Login & Get Token

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/open/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser", "password":"password123"}' | jq -r .data.token)

echo "Token: $TOKEN"
```

### 3. Create a Category

```bash
curl -X POST http://localhost:8080/api/category \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Food",
    "type": "expense",
    "remark": "Food and dining expenses"
  }'
```

### 4. Create an Expense

```bash
curl -X POST http://localhost:8080/api/cash/expense \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "belongs_date": "2025-12-12",
    "category_name": "Food",
    "amount": 45.50,
    "description": "Lunch at restaurant"
  }'
```

### 5. List All Transactions

```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/cash?limit=20&offset=0"
```

---

## 🛠️ Troubleshooting

### Problem: Can't connect to MongoDB
**Solution**:
```bash
docker compose ps
docker compose restart mongodb
```

### Problem: API returns "401 Unauthorized"
**Solution**:
Ensure you are including the header `-H "Authorization: Bearer $TOKEN"` and that the token is valid.

### Problem: Port 8080 already in use
**Solution**:
```bash
./cashlenx open start -p 8081
```

### Problem: Permission denied when running ./cashlenx
**Solution**:
```bash
chmod +x cashlenx
```

---

## 🧹 Step 8: Cleanup (When Done Testing)

### Stop API Server

```bash
# If running in foreground: Ctrl+C

# If running in background:
kill $(cat server.pid)
rm server.pid
```

### Stop MongoDB

```bash
docker compose down
```

---

## 🎯 Next Steps

1. **Run all 21 endpoint tests** - See `docs/api.md`
2. **Try the CLI commands** - See `docs/cli.md`
3. **Import test data** - Use `cashlenx statistic import`
4. **Export data** - Use `cashlenx statistic export`
