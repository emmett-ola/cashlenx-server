# Backend Testing Guide

Testing guide for CashLenX backend (CLI and API).

## Prerequisites

- Go 1.21+
- MongoDB or MySQL running
- Docker (for database)

## Quick Start

### 1. Start Database

```bash
# Start MongoDB with demo data
docker compose up -d mongodb

# Verify it's running
docker compose ps

# Check logs
docker compose logs -f mongodb
```

### 2. Verify Demo Data

```bash
# Connect to MongoDB
docker exec -it cashlenx-mongodb mongosh \
  -u cashlenx -p cashlenx123 \
  --authenticationDatabase admin cashlenx

# In MongoDB shell:
db.cash_flows.countDocuments()  # Should return 15
db.categories.countDocuments()  # Should return 8

# Check summary
db.cash_flows.aggregate([
  {
    $group: {
      _id: "$flow_type",
      total: { $sum: "$amount" },
      count: { $count: {} }
    }
  }
])

# Exit
exit
```

### 3. Set Environment Variables

```bash
# MongoDB
export MONGO_DB_URI="mongodb://cashlenx:cashlenx123@localhost:27017/cashlenx?authSource=admin"
export DB_TYPE=mongodb
export DB_NAME=cashlenx
export JWT_SECRET="your-secret-key"

# Or use .env file
export $(cat .env | xargs)
```

## CLI Testing

### Build CLI

```bash
# In project root (cashlenx-server)
go build -o cashlenx main.go
```

### Test Commands

```bash
# Version
./cashlenx open version

# Add expense
./cashlenx cash expense -c "Food & Dining" -a 45.50 -d "Lunch"

# Add income
./cashlenx cash income -c "Salary" -a 5000

# Query transactions
./cashlenx cash query -b 2024-12-04

# List categories
./cashlenx category list

# Export data
./cashlenx statistic export -o test_export.xlsx

# Show summary
./cashlenx statistic summary -p monthly -d 2024-12
```

See [cli.md](cli.md) for complete command reference.

## API Testing

### Start Server

```bash
./cashlenx open start -p 8080
```

### Test Endpoints

#### Public Endpoints

```bash
# Health check
curl http://localhost:8080/api/open/health

# Version info
curl http://localhost:8080/api/open/version
```

#### Authentication

```bash
# Register
curl -X POST http://localhost:8080/api/open/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"tester","password":"password123","email":"tester@example.com"}'

# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/open/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"tester","password":"password123"}' | jq -r .data.token)
```

#### User Endpoints (Require Token)

```bash
# Create category
curl -X POST http://localhost:8080/api/category \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"TestCat","type":"expense"}'

# Create expense
curl -X POST http://localhost:8080/api/cash/expense \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 45.50,
    "belongs_date": "2024-12-05",
    "category_name": "TestCat",
    "description": "Lunch"
  }'

# Get today's transactions
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/cash/date/$(date +%Y-%m-%d)

# Delete by ID
curl -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/cash/{id}
```

### Test CORS

```bash
curl -i -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     http://localhost:8080/api/open/health
```

Should return CORS headers:
```
Access-Control-Allow-Origin: http://localhost:3000
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

## Unit Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./service/cash_flow_service/

# Verbose output
go test -v ./...
```

## Integration Testing

### Test Database Connection

```bash
# This command doesn't exist anymore as a standalone, 
# but health check validates DB connection
curl http://localhost:8080/api/open/health
```

### Test Full Workflow

```bash
# 1. Add income
./cashlenx cash income -c "Salary" -a 5000

# 2. Add expenses
./cashlenx cash expense -c "Food" -a 45.50 -d "Lunch"
./cashlenx cash expense -c "Transportation" -a 20 -d "Bus fare"

# 3. Query today's transactions
./cashlenx cash query -b $(date +%Y-%m-%d)

# 4. Export data
./cashlenx statistic export -o test_data.xlsx

# 5. Check stats
./cashlenx statistic summary -p monthly -d $(date +%Y-%m)
```

## Performance Testing

### Load Testing with Apache Bench

```bash
# Install Apache Bench
sudo apt-get install apache2-utils

# Test health endpoint
ab -n 1000 -c 10 http://localhost:8080/api/open/health
```

### Benchmark Tests

```bash
# Run Go benchmarks
go test -bench=. ./...

# With memory profiling
go test -bench=. -benchmem ./...
```

## Troubleshooting

### Database Connection Issues

```bash
# Check if MongoDB is running
docker compose ps

# Check MongoDB logs
docker compose logs mongodb

# Restart MongoDB
docker compose restart mongodb

# Reset database
docker compose down -v
docker compose up -d mongodb
```

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different port
./cashlenx open start -p 8081
```

### Build Errors

```bash
# Clean build cache
go clean -cache

# Update dependencies
go mod tidy
go mod download

# Rebuild
go build -o cashlenx main.go
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Backend Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mongodb:
        image: mongo:7
        env:
          MONGO_INITDB_ROOT_USERNAME: cashlenx
          MONGO_INITDB_ROOT_PASSWORD: cashlenx123
        ports:
          - 27017:27017
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: |
          go test -v ./...
      
      - name: Build
        run: |
          go build -o cashlenx main.go
```

## Next Steps

1. Implement remaining API endpoints (see [api.md](api.md))
2. Add unit tests for new services
3. Add integration tests for API endpoints
4. Set up CI/CD pipeline
5. Add performance benchmarks

## See Also

- [CLI Reference](cli.md) - Complete CLI documentation
- [API Reference](api.md) - API implementation tasks
