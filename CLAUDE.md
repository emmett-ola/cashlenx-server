# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CashLenX Server is a Go-based personal finance tracking backend providing both CLI and REST API interfaces. It supports multi-user deployments with strict data isolation, pluggable storage (MongoDB/MySQL), and comprehensive financial analytics.

**Key Technologies:**
- Go 1.23+
- Cobra (CLI framework)
- Gorilla Mux (HTTP routing)
- Zap (structured logging)
- MongoDB and MySQL drivers
- JWT authentication
- Excelize (Excel), gofpdf (PDF generation)

## Common Commands

### Development Workflow

```bash
# Set up database (choose one)
docker compose up -d mongodb
docker compose --profile mysql up -d mysql

# Configure environment
cp .env.sample .env
export $(cat .env | xargs)

# Run the server
go run main.go server start -p 8080

# Run tests
go test ./...                    # All tests
go test ./service/...            # Specific package tests
go test -run TestName ./package  # Single test
go test -v ./...                 # Verbose output

# Build
go build -o cashlenx main.go

# CLI usage examples
go run main.go cash expense -c "Food" -a 45.50 -d "Lunch"
go run main.go cash summary -p monthly -d 2024-01
go run main.go admin backup -o backup.json
```

### Database Operations

```bash
# MongoDB
docker exec -it cashlenx-mongodb mongosh -u cashlenx -p cashlenx123 --authenticationDatabase admin

# MySQL
docker exec -it cashlenx-mysql mysql -u cashlenx -p cashlenx
```

## Architecture

### Layered Architecture Pattern

The codebase follows a strict 4-layer architecture with clear separation of concerns:

```
HTTP Request → Controller → Service → Mapper → Database
              (routing)    (logic)   (data)
```

**Critical Rule:** Data isolation is enforced at ALL layers for multi-tenant security.

### 1. Mapper Layer (Database Abstraction)

**Location:** `mapper/`

**Pattern:** Strategy Pattern with factory initialization

Each entity has:
- An interface defining data operations (`CashFlowMapper`, `UserMapper`, etc.)
- Two implementations: `*MongoDbMapper` and `*MySqlMapper`
- A singleton `INSTANCE` variable selected at runtime based on `DB_TYPE` config

**Dual-mode Methods:**
- **Admin methods:** No user filtering (e.g., `GetAllCashFlows()`)
- **User-isolated methods:** Suffixed with `AndUser` (e.g., `GetCashFlowByIdAndUser(id, userId)`)

**Important Files:**
- `mapper/cash_flow_mapper/mongodb.go` - MongoDB implementation
- `mapper/cash_flow_mapper/mysql.go` - MySQL implementation
- `mapper/cash_flow_mapper/interface.go` - Shared interface

**Soft-Delete Pattern:** All deletes set `is_delete=true` and populate `delete_time`, `delete_user_id`. Queries automatically exclude deleted records.

### 2. Service Layer (Business Logic)

**Location:** `service/`

**Key Services:**
- `cash_flow_service/` - CRUD and analytics for income/expenses
- `category_service/` - Category hierarchy management
- `user_service/` - User lifecycle, authentication
- `statistic_service/` - Financial reporting and analytics
- `manage_service/` - Backup/restore/import/export
- `refresh_token_service/` - Token lifecycle
- `db_service/` - Database connection management

**Validation Pattern:** All services validate inputs using the `validation` package before calling mappers:

```go
if err := validation.ValidateAmount(amount); err != nil {
    return model.CashFlowEntity{}, err
}
```

**User Context:** Services receive `userId` parameter and pass it to `*AndUser` mapper methods to enforce data isolation.

### 3. Controller Layer (HTTP Handlers)

**Location:** `controller/`

**Route Organization:**
- `/api/open/*` - Public (health, version, login, register)
- `/api/admin/*` - Admin-only (backup, restore, user management)
- `/api/cash/*` - User cash flows (requires auth)
- `/api/category/*` - User categories (requires auth)
- `/api/statistic/*` - User analytics (requires auth)

**User Context Extraction:**
```go
userId, ok := r.Context().Value("user_id").(string)
if !ok || userId == "" {
    return errors.NewUnauthorizedError("user not authenticated")
}
```

### 4. Middleware Stack

**Location:** `middleware/`

**Execution Order (LIFO - innermost first):**
1. **CORS** (`cors.go`) - Handles preflight and headers
2. **SchemaValidation** (`schema_validation.go`) - OpenAPI spec validation (optional)
3. **Auth** (`auth.go`) - JWT validation, populates context with `user_id`, `username`, `role`
4. **Logging** (`logging.go`) - Request/response logging

**Registration in `controller/server.go`:**
```go
handler := middleware.Logging(
    middleware.Auth(
        middleware.SchemaValidation(
            middleware.CORS(router)
        )
    )
)
```

### Authentication System

**Location:** `auth/`

**Pattern:** Provider pattern with service wrapper

- `auth/service/auth_service.go` - Service wrapper
- `auth/provider/local_auth.go` - JWT implementation
- `auth/auth.go` - Global singleton

**JWT Flow:**
1. User logs in with username/password
2. Service validates credentials via bcrypt
3. Returns access token (JWT) + refresh token (stored in DB)
4. Client sends `Authorization: Bearer <token>` on subsequent requests
5. Auth middleware validates JWT and populates request context

**Token Types:**
- **Access Token:** Short-lived JWT (configurable hours via `JWT_EXPIRATION_HOURS`)
- **Refresh Token:** Long-lived token stored in `refresh_token_mapper`
- **Password Reset Token:** Time-limited token in `password_reset_mapper`

### Data Models

**Location:** `model/`

**Base Entity Pattern:**
All entities embed `BaseEntity`:
```go
type BaseEntity struct {
    CreateUserId  primitive.ObjectID
    CreateTime    time.Time
    UpdateUserId  primitive.ObjectID
    UpdateTime    time.Time
    DeleteUserId  *primitive.ObjectID  // Null if not deleted
    DeleteTime    *time.Time           // Null if not deleted
    IsDelete      bool                 // Soft-delete flag
}
```

**Key Entities:**
- `CashFlowEntity` - Income/expense transactions
- `CategoryEntity` - Category with hierarchical support (ParentId)
- `UserEntity` - User account (password never exposed in JSON)

**JSON Marshaling:** Models override `MarshalJSON()` to convert UTC timestamps to configured timezone for API responses.

**DTOs (Data Transfer Objects):**
- `CashFlowDTO` - Input for cash flow creation
- `UserDTO` - API representation (excludes password hash)
- `UserLoginRequest/Response` - Login endpoints

### Validation

**Location:** `validation/validators.go`

Centralized validation functions:
- `ValidateDate(dateStr)` - Supports YYYYMMDD, YYYY-MM-DD, YYYY/MM/DD
- `ValidateDateRange(from, to)` - Ensures from <= to
- `ValidateAmount(amount)` - Positive, max 999999999.99
- `ValidateID(id)` - 24-char hex string
- `ValidateCategoryName(name)` - Max 100 chars
- `ValidatePassword(password)` - 6-100 chars

### Configuration

**Location:** `util/config/`

**Pattern:** Key-value store with environment variable override

**Critical Keys:**
- `db.type` - "mongodb" or "mysql"
- `db.mongodb.url` / `db.mysql.url` - Connection strings
- `db.name` - Database name
- `auth.jwt.secret` - JWT signing key
- `auth.jwt.expiration_hours` - Token lifetime
- `api.schema.validation` - Enable OpenAPI validation
- `cors.origins` - Comma-separated allowed origins
- `timezone` - Application timezone (default: UTC)

**Usage:**
```go
dbType := util.GetConfigByKey("db.type")
util.SetConfigByKey("db.type", "mongodb")
```

### Database Connection Management

**Location:** `util/database/`

**Connection Pooling:**
- **MongoDB:** Min 10, max 50 connections, 5-min idle timeout
- **MySQL:** Max 10 open connections, 3-min max lifetime

**Singleton Pattern:** `sync.Once` ensures single pool initialization

**Files:**
- `mongodb_util.go` - MongoDB connection pool and collection access
- `mysql_util.go` - MySQL connection lifecycle
- `database_util.go` - Shared table/collection name constants

## Critical Patterns and Conventions

### User Data Isolation (Security Critical)

**Three-layer enforcement:**

1. **Mapper layer:** Use `*AndUser()` methods
   ```go
   GetCashFlowByObjectIdAndUser(id string, userId primitive.ObjectID)
   ```

2. **Service layer:** Accept and pass `userId` parameter
   ```go
   func GetCashFlowForUser(id string, userId string) (model.CashFlowEntity, error)
   ```

3. **Controller layer:** Extract from JWT context
   ```go
   userId := r.Context().Value("user_id").(string)
   ```

**Never skip this pattern** - it's the security boundary between users.

### Soft-Delete Pattern

All entity deletions:
1. Set `is_delete = true`
2. Set `delete_time = now()`
3. Set `delete_user_id = current_user`
4. Keep record in database for audit trail

Queries automatically filter `is_delete = false` unless using `IncludeDeleted` methods.

### Error Handling

**Location:** `errors/`

Custom error types:
- `UnauthorizedError` - 401 (missing/invalid auth)
- `ForbiddenError` - 403 (insufficient permissions)
- `ValidationError` - 400 (input validation)
- `FieldValidationError` - 400 (per-field errors)
- `InternalError` - 500 (server errors)

### Timezone Handling

- **Storage:** Always UTC (`time.UTC`)
- **API responses:** Converted to configured timezone via `util.ToTimezone()`
- **Configuration:** `timezone` config key (default: "UTC")

### Decimal Precision

Uses `github.com/shopspring/decimal` for monetary amounts to avoid floating-point precision issues.

## Project Structure Details

```
cashlenx-server/
├── cmd/                    # Cobra commands
│   ├── server_cmd/         # Server start command
│   ├── cash_cmd/           # Cash flow CLI commands
│   ├── category_cmd/       # Category CLI commands
│   ├── admin_cmd/          # Admin CLI commands
│   └── statistic_cmd/      # Statistics CLI commands
├── controller/             # HTTP controllers
│   ├── server.go           # Server initialization & routing
│   ├── cash_controller/    # Cash flow endpoints
│   ├── category_controller/# Category endpoints
│   ├── user_controller/    # User management endpoints
│   ├── statistic_controller/# Analytics endpoints
│   └── manage_controller/  # Backup/restore endpoints
├── service/                # Business logic (see above)
├── mapper/                 # Data access (see above)
├── model/                  # Data models and DTOs
├── middleware/             # HTTP middleware
├── auth/                   # Authentication system
├── validation/             # Input validation
├── util/                   # Utilities
│   ├── config/             # Configuration management
│   ├── database/           # Database connections
│   └── timezone/           # Timezone conversion
├── errors/                 # Custom error types
├── cache/                  # Caching utilities
├── migrations/             # Database migration scripts
├── docker/                 # Docker init scripts
├── docs/                   # Documentation
│   ├── api.md              # REST API reference
│   ├── cli.md              # CLI reference
│   └── roadmap.md          # Feature roadmap
└── main.go                 # Entry point
```

## Important Implementation Notes

### When Adding New Features

1. **Create mapper interface and implementations** - Both MongoDB and MySQL
2. **Implement service layer** - Add validation, call mapper methods
3. **Create controller** - Extract user context, call service
4. **Register routes** - Add to appropriate route group in `controller/server.go`
5. **Add validation** - Use or extend `validation` package
6. **Consider user isolation** - Always enforce unless explicitly admin-only

### When Modifying Database Queries

- **Always** use parameterized queries (never string concatenation)
- **Always** include `is_delete = false` filter (unless explicitly querying deleted records)
- **Always** include `belongs_user_id = userId` filter for user-specific queries
- **Consider** adding both admin and user-isolated versions of methods

### When Adding Authentication

- User context is in `r.Context().Value("user_id")`, `r.Context().Value("username")`, `r.Context().Value("role")`
- Admin routes should check `role == "admin"`
- Use `middleware.AdminAuth()` for admin-only routes

### Testing Guidelines

- Tests should cover both MongoDB and MySQL implementations
- Use `db_service.ConnectDB()` for test database setup
- Clean up test data after each test
- Test both success and error paths
- Test user isolation (user A cannot access user B's data)

## Environment Variables Reference

See `.env.sample` for complete list. Key variables:

```bash
# Database
DB_TYPE=mongodb              # or mysql
MONGO_DB_URI=mongodb://...   # MongoDB connection
MYSQL_DB_URI=user:pass@...   # MySQL connection
DB_NAME=cashlenx

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
CORS_ORIGINS=http://localhost:3000

# Authentication
JWT_SECRET=your-secret-key   # CHANGE IN PRODUCTION
JWT_EXPIRATION_HOURS=24
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin

# Logging
LOG_LEVEL=info
LOG_FOLDER=./logs

# Timezone
TIMEZONE=UTC
```

## API Documentation

Full API documentation in `docs/api.md`. Key points:

- **All endpoints return JSON** in standardized format:
  ```json
  {
    "code": "OK",
    "message": "Success",
    "data": {...},
    "errors": []
  }
  ```

- **Authentication:** Most endpoints require `Authorization: Bearer <token>` header

- **User isolation:** Attempting to access another user's data returns 404 (not 403) to prevent information disclosure

## CLI Documentation

Full CLI documentation in `docs/cli.md`. The CLI mirrors the API functionality with commands organized by access level (`open`, `admin`, `cash`, `category`, `statistic`).

## Migration Notes

When switching between MongoDB and MySQL:
1. Change `DB_TYPE` in `.env`
2. Update connection string (`MONGO_DB_URI` or `MYSQL_DB_URI`)
3. Restart server
4. No code changes needed - mapper layer handles differences

## Production Considerations

1. **Change JWT_SECRET** - Use strong random key
2. **Change default admin password** - Update `ADMIN_PASSWORD`
3. **Set LOG_LEVEL** - Use "warn" or "error" in production
4. **Configure CORS** - Restrict `CORS_ORIGINS` to known domains
5. **Use connection pooling** - Already configured, but monitor limits
6. **Enable HTTPS** - Use reverse proxy (nginx, caddy)
7. **Backup regularly** - Use `cashlenx admin backup` command
