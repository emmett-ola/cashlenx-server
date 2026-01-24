# CashLenX API Documentation

**Version**: 2.1.0
**Last Updated**: 2025-12-28

## Overview

CashLenX provides a RESTful API for personal finance management with multi-user support and data isolation.

### Base URL
```
http://localhost:8080/api/v0
```

**Note**: The API version is configurable via `API_VERSION` environment variable. Default is `v0`.
Paths in this documentation are relative to the versioned base URL (e.g., `/open/health` maps to `/api/v0/open/health`).

### Authentication
Most endpoints require JWT authentication. Include the token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

### Route Organization

Routes are organized by access level:

- **`/open/*`** - Public endpoints (no authentication required)
- **`/admin/*`** - Admin-only endpoints (requires admin role)
- **`/cash/*`** - User-specific cash flow operations (requires authentication)
- **`/category/*`** - User-specific category operations (requires authentication)
- **`/statistic/*`** - User-specific analytics and exports (requires authentication)

## API Reference

### Public Endpoints (`/open/*`)

#### Health Check
```http
GET /open/health
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Version Info
```http
GET /open/version
```

**Response**:
```json
{
  "version": "2.0.0",
  "buildTime": "2024-01-15T10:00:00Z",
  "gitCommit": "abc1234"
}
```

#### User Login
```http
POST /open/auth/login
```

**Request**:
```json
{
  "username": "john_doe",
  "password": "securepassword"
}
```

**Response**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "username": "john_doe",
    "role": "user"
  }
}
```

#### User Registration
```http
POST /open/auth/register
```

**Request**:
```json
{
  "username": "john_doe",
  "password": "securepassword",
  "email": "john@example.com"
}
```

**Response**:
```json
{
  "id": "507f1f77bcf86cd799439011",
  "username": "john_doe",
  "role": "user"
}
```

---

### Cash Flow Endpoints (`/cash/*`)

All cash flow endpoints enforce user data isolation - users can only access their own transactions.

#### Create Income
```http
POST /cash/income
Authorization: Bearer <token>
```

**Request**:
```json
{
  "amount": 5000.00,
  "belongs_date": "2024-01-15",
  "category_name": "Salary",
  "description": "Monthly salary"
}
```

#### Create Expense
```http
POST /cash/expense
Authorization: Bearer <token>
```

**Request**:
```json
{
  "amount": 45.50,
  "belongs_date": "2024-01-15",
  "category_name": "Food & Dining",
  "description": "Lunch"
}
```

#### Query by ID
```http
GET /cash/{id}
Authorization: Bearer <token>
```

**Note**: Only returns the transaction if it belongs to the authenticated user.

#### Query by Date
```http
GET /cash/date/{date}
Authorization: Bearer <token>
```

**Example**: `GET /cash/date/2024-01-15`

**Note**: Only returns transactions belonging to the authenticated user.

#### Update Transaction
```http
PUT /cash/{id}
Authorization: Bearer <token>
```

**Request**:
```json
{
  "amount": 50.00,
  "category": "Groceries",
  "description": "Updated description"
}
```

**Note**: Can only update transactions belonging to the authenticated user.

#### Delete by ID
```http
DELETE /cash/{id}
Authorization: Bearer <token>
```

**Note**: Can only delete transactions belonging to the authenticated user.

#### Delete by Date
```http
DELETE /cash/date/{date}
Authorization: Bearer <token>
```

**Note**: Only deletes transactions belonging to the authenticated user.

#### List Transactions
```http
GET /cash?limit=50&offset=0&type=expense
Authorization: Bearer <token>
```

**Query Parameters**:
- `limit` (optional): Max records to return (default: 50)
- `offset` (optional): Records to skip (default: 0)
- `type` (optional): Filter by type (`income` or `expense`)

**Note**: Only returns transactions belonging to the authenticated user.

#### Query Date Range
```http
GET /cash/range?from=2024-01-01&to=2024-01-31
Authorization: Bearer <token>
```

**Query Parameters**:
- `from` (required): Start date (YYYY-MM-DD)
- `to` (required): End date (YYYY-MM-DD)

**Response**:
```json
{
  "from": "2024-01-01",
  "to": "2024-01-31",
  "total_income": 5000.00,
  "total_expense": 2500.00,
  "balance": 2500.00,
  "count": 15,
  "transactions": [...]
}
```

**Note**: Only returns transactions belonging to the authenticated user.

#### Monthly Summary
```http
GET /cash/summary/monthly/{yyyymm}
Authorization: Bearer <token>
```

**Example**: `GET /cash/summary/monthly/202401`

**Response**:
```json
{
  "period": "2024-01",
  "income": 5000.00,
  "expense": 2500.00,
  "balance": 2500.00,
  "transaction_count": 15,
  "income_count": 2,
  "expense_count": 13
}
```

**Note**: Only includes transactions belonging to the authenticated user.

---

### Category Endpoints (`/category/*`)

All category endpoints enforce user data isolation - users can only access their own categories.

#### Create Category
```http
POST /category
Authorization: Bearer <token>
```

**Request**:
```json
{
  "name": "Food & Dining",
  "type": "expense",
  "parent_id": "507f1f77bcf86cd799439011",
  "remark": "All food-related expenses"
}
```

**Note**: Category is created for the authenticated user. Name must be unique within the same parent category (or root) for the user and type.

#### List Categories
```http
GET /category?limit=50&offset=0&type=expense
Authorization: Bearer <token>
```

**Query Parameters**:
- `limit` (optional): Max records to return (default: 50)
- `offset` (optional): Records to skip (default: 0)
- `type` (optional): Filter by type (`income` or `expense`)

**Note**: Only returns categories belonging to the authenticated user.

#### Query by ID
```http
GET /category/{id}
Authorization: Bearer <token>
```

**Note**: Only returns the category if it belongs to the authenticated user.

#### Query by Name
```http
GET /category/name/{name}
Authorization: Bearer <token>
```

**Example**: `GET /category/name/Food%20%26%20Dining`

**Note**: Only searches categories belonging to the authenticated user.

#### Get Child Categories
```http
GET /category/{id}/children?type=expense
Authorization: Bearer <token>
```

**Query Parameters**:
- `type` (optional): Filter by type (`income` or `expense`)

**Note**: Only returns child categories if parent belongs to the authenticated user.

#### Update Category
```http
PUT /category/{id}
Authorization: Bearer <token>
```

**Request**:
```json
{
  "name": "Dining Out",
  "type": "expense",
  "parent_id": "507f1f77bcf86cd799439011",
  "remark": "Updated description"
}
```

**Note**: Can only update categories belonging to the authenticated user. Name must be unique within the same parent category (or root) for the user and type.

#### Delete Category
```http
DELETE /category/{id}
Authorization: Bearer <token>
```

**Note**: Can only delete categories belonging to the authenticated user.

#### Get Category Tree
```http
GET /category/tree?type=expense
Authorization: Bearer <token>
```

**Query Parameters**:
- `type` (optional): Filter by type (`income` or `expense`)

**Response**: Hierarchical tree structure of categories

**Note**: Only returns categories belonging to the authenticated user.

---

### Admin Endpoints (`/admin/*`)

All admin endpoints require the `admin` role.

#### User Management

##### Create User
```http
POST /admin/user
Authorization: Bearer <admin-token>
```

**Request**:
```json
{
  "username": "new_user",
  "password": "securepassword",
  "email": "user@example.com",
  "role": "user"
}
```

##### List Users
```http
GET /admin/user?limit=50&offset=0
Authorization: Bearer <admin-token>
```

##### Get User by ID
```http
GET /admin/user/{id}
Authorization: Bearer <admin-token>
```

##### Update User
```http
PUT /admin/user/{id}
Authorization: Bearer <admin-token>
```

##### Delete User
```http
DELETE /admin/user/{id}
Authorization: Bearer <admin-token>
```

#### Database Management

##### Database Backup
```http
GET /admin/manage/dump
Authorization: Bearer <admin-token>
Header: ADMIN_TOKEN=<your-admin-token>
```

**Response**: JSON file containing all database data (all users)

**Statistics Returned**:
- Users: success/failed counts
- Categories: success/failed counts
- Cash Flows: success/failed counts

##### Database Restore
```http
POST /admin/manage/restore
Authorization: Bearer <admin-token>
Header: ADMIN_TOKEN=<your-admin-token>
Content-Type: application/json
```

**Request**: Upload backup JSON file

**Statistics Returned**:
- Users: success/failed counts
- Categories: success/failed counts
- Cash Flows: success/failed counts

##### Export to Excel
```http
GET /admin/manage/export?from=2024-01-01&to=2024-01-31
Authorization: Bearer <admin-token>
```

**Query Parameters**:
- `from` (optional): Start date (YYYY-MM-DD)
- `to` (optional): End date (YYYY-MM-DD)

**Response**: Excel file (.xlsx)

**TODO**: This endpoint will be moved to `/statistic/export` with user data isolation.

##### Import from Excel
```http
POST /admin/manage/import
Authorization: Bearer <admin-token>
Content-Type: multipart/form-data
```

**Request**: Upload Excel file (.xlsx)

**TODO**: This endpoint will be moved to `/statistic/import` with user data isolation.

---

## ✅ User Statistic Endpoints (`/statistic/*`)

All statistic endpoints are user-specific with complete data isolation. Only authenticated users can access their own data.

### Export/Import Operations

#### Export User Data (Multi-Format)
```http
GET /statistic/export?format=xlsx&from_date=20240101&to_date=20241231
Authorization: Bearer <token>
```

**Query Parameters**:
- `format` (optional): Export format - `xlsx` (default), `csv`, or `pdf`
- `from_date` (optional): Start date (YYYYMMDD or YYYY-MM-DD)
- `to_date` (optional): End date (YYYYMMDD or YYYY-MM-DD)

**Response**: Binary file download with appropriate Content-Type
- Excel: `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
- CSV: `text/csv`
- PDF: `application/pdf`

**Headers**:
```
Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
Content-Disposition: attachment; filename=cashlenx-export-12345678-20250128-103045.xlsx
Content-Length: 15234
```

**Note**: Only exports data belonging to the authenticated user. File is automatically downloaded.

#### Import User Data
```http
POST /statistic/import?file_path=/path/to/file.xlsx
Authorization: Bearer <token>
```

**Query Parameters**:
- `file_path` (required): Path to Excel file to import

**Response**:
```json
{
  "message": "Data imported successfully",
  "file_path": "/path/to/file.xlsx",
  "user_id": "507f1f77bcf86cd799439011"
}
```

**Note**: Only imports data to the authenticated user's account. Categories are auto-created if needed.

### Summary Endpoints

#### Daily Summary
```http
GET /statistic/summary/daily/20240115
Authorization: Bearer <token>
```

**Response**:
```json
{
  "period": "20240115",
  "period_type": "daily",
  "income": 500.00,
  "expense": 125.50,
  "balance": 374.50,
  "transaction_count": 8,
  "income_count": 1,
  "expense_count": 7,
  "average_transaction": 78.19,
  "categories": {
    "Food": -45.00,
    "Transport": -30.00,
    "Salary": 500.00
  }
}
```

#### Monthly Summary
```http
GET /statistic/summary/monthly/202401
Authorization: Bearer <token>
```

#### Yearly Summary
```http
GET /statistic/summary/yearly/2024
Authorization: Bearer <token>
```

### Breakdown Endpoints

#### Daily Breakdown
```http
GET /statistic/breakdown/daily/20240115
Authorization: Bearer <token>
```

**Response**:
```json
{
  "period": "20240115",
  "total_expense": 125.50,
  "total_income": 500.00,
  "expense_categories": [
    {
      "category": "Food",
      "amount": 45.00,
      "percentage": 35.86,
      "count": 3
    },
    {
      "category": "Transport",
      "amount": 30.00,
      "percentage": 23.90,
      "count": 2
    }
  ],
  "income_categories": [
    {
      "category": "Salary",
      "amount": 500.00,
      "percentage": 100.00,
      "count": 1
    }
  ]
}
```

#### Monthly Breakdown
```http
GET /statistic/breakdown/monthly/202401
Authorization: Bearer <token>
```

#### Yearly Breakdown
```http
GET /statistic/breakdown/yearly/2024
Authorization: Bearer <token>
```

### Trends Endpoints

#### Daily Trends
```http
GET /statistic/trends/daily/20240115
Authorization: Bearer <token>
```

**Response**:
```json
{
  "period": "20240115",
  "period_type": "daily",
  "data_points": [
    {
      "date": "2024-01-15",
      "income": 500.00,
      "expense": 125.50,
      "balance": 374.50
    }
  ],
  "trends": {
    "income_trend": "stable",
    "expense_trend": "increasing",
    "average_monthly_expense": 125.50
  }
}
```

#### Monthly Trends
```http
GET /statistic/trends/monthly/202401
Authorization: Bearer <token>
```

**Response**: Daily data points for the entire month with trend analysis.

#### Yearly Trends
```http
GET /statistic/trends/yearly/2024
Authorization: Bearer <token>
```

**Response**: Monthly data points for the entire year with trend analysis.

### Top Expenses Endpoints

#### Top Daily Expenses
```http
GET /statistic/top/daily/20240115?limit=10
Authorization: Bearer <token>
```

**Query Parameters**:
- `limit` (optional): Number of top expenses to return (default: 10)

**Response**:
```json
{
  "period": "20240115",
  "limit": 10,
  "total_expense": 125.50,
  "expenses": [
    {
      "id": "507f1f77bcf86cd799439011",
      "date": "20240115",
      "category": "Food",
      "amount": 45.00,
      "description": "Dinner at restaurant",
      "percentage": 35.86
    }
  ]
}
```

#### Top Monthly Expenses
```http
GET /statistic/top/monthly/202401?limit=10
Authorization: Bearer <token>
```

#### Top Yearly Expenses
```http
GET /statistic/top/yearly/2024?limit=10
Authorization: Bearer <token>
```

### Dashboard Visualization Endpoints

These endpoints return data optimized for frontend visualization libraries (Chart.js, D3.js, etc.).

#### Dashboard Overview
```http
GET /statistic/dashboard/monthly/202401
Authorization: Bearer <token>
```

**Response**:
```json
{
  "period": "202401",
  "period_type": "monthly",
  "summary": { /* Summary object */ },
  "top_categories": [
    {
      "category": "Food",
      "amount": 450.00,
      "percentage": 35.00,
      "count": 15
    }
  ],
  "recent_trend": "increasing",
  "quick_stats": {
    "total_transactions": 48,
    "average_daily": 42.58,
    "highest_expense": 250.00,
    "lowest_expense": 5.00
  }
}
```

#### Income vs Expense Chart Data
```http
GET /statistic/chart/income-expense/monthly/202401
Authorization: Bearer <token>
```

**Response**:
```json
{
  "labels": ["2024-01-01", "2024-01-02", "2024-01-03"],
  "income": [500.00, 0.00, 0.00],
  "expense": [125.50, 85.00, 65.00],
  "balance": [374.50, -85.00, -65.00],
  "period": "monthly",
  "from_date": "2024-01-01",
  "to_date": "2024-01-31"
}
```

#### Category Distribution Chart Data
```http
GET /statistic/chart/category-distribution/monthly/202401?type=expense
Authorization: Bearer <token>
```

**Query Parameters**:
- `type` (optional): `income` or `expense` (default: expense)

**Response**:
```json
{
  "labels": ["Food", "Transport", "Entertainment"],
  "values": [450.00, 300.00, 200.00],
  "percentages": [47.37, 31.58, 21.05],
  "colors": ["#FF6384", "#36A2EB", "#FFCE56"],
  "total": 950.00,
  "type": "expense"
}
```

#### Monthly Comparison Chart Data
```http
GET /statistic/chart/monthly-comparison/2024
Authorization: Bearer <token>
```

**Response**:
```json
{
  "year": "2024",
  "months": ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"],
  "income": [5000, 5200, 4800, 5100, 5300, 5000, 5200, 5400, 5100, 5000, 5300, 5500],
  "expense": [3500, 3200, 3800, 3600, 3400, 3700, 3500, 3600, 3800, 3900, 3700, 4000],
  "balance": [1500, 2000, 1000, 1500, 1900, 1300, 1700, 1800, 1300, 1100, 1600, 1500]
}
```

#### Spending Heatmap Data
```http
GET /statistic/chart/spending-heatmap/2024
Authorization: Bearer <token>
```

**Response**:
```json
{
  "year": "2024",
  "data": [
    {
      "date": "2024-01-01",
      "amount": 125.50,
      "count": 3
    },
    {
      "date": "2024-01-02",
      "amount": 85.00,
      "count": 2
    }
  ],
  "max": 500.00,
  "min": 5.00
}
```

---

## Data Isolation

### User Data Isolation (Implemented)

All user-specific endpoints enforce strict data isolation:

- **Cash flows**: Users can only access their own transactions
- **Categories**: Users can only access their own categories
- **Three-layer enforcement**:
  1. **Mapper layer**: `*AndUser()` methods enforce database-level filtering
  2. **Service layer**: `*ForUser()` methods provide user-specific business logic
  3. **Controller layer**: Extracts `userId` from JWT and passes to services

### Admin Data Access

Admin endpoints can access data across all users:

- **Backup/Restore**: Includes data from all users with proper user_id references
- **User Management**: Admins can create, update, and delete users
- **Database Operations**: Full database access for backup and restore

---

## Error Handling

### Standard Response Format

**Success Response**:
```json
{
  "code": "OK",
  "message": "Success",
  "data": { ... },
  "meta": {},
  "extra": {},
  "errors": []
}
```

**Error Response**:
```json
{
  "code": "ERROR",
  "message": "Error message",
  "data": null,
  "meta": {},
  "extra": {},
  "errors": [
    {
      "field": "amount",
      "message": "Amount must be positive"
    }
  ]
}
```

### HTTP Status Codes

- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions (e.g., non-admin trying to access admin endpoint)
- `404 Not Found` - Resource not found or not owned by user
- `500 Internal Server Error` - Server error

---

## Testing

### Authentication Testing

```bash
# Register new user
curl -X POST http://localhost:8080/api/v0/open/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123","email":"test@example.com"}'

# Login
curl -X POST http://localhost:8080/api/v0/open/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123"}'

# Use token for authenticated requests
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v0/cash
```

### Data Isolation Testing

```bash
# User A creates transaction
curl -X POST http://localhost:8080/api/v0/cash/expense \
  -H "Authorization: Bearer $TOKEN_USER_A" \
  -H "Content-Type: application/json" \
  -d '{"amount":50,"category":"Food","description":"Lunch"}'

# User B cannot access User A's transaction
curl -H "Authorization: Bearer $TOKEN_USER_B" \
  http://localhost:8080/api/v0/cash/$TRANSACTION_ID_FROM_USER_A
# Should return 404 Not Found
```

---

## Version History

### v2.2.0 (Current)
- ✅ Global API versioning (default `/api/v0`)
- ✅ Simplified authentication flow (merged login/refresh)
- ✅ Enhanced database backup/restore with proper content types
- ✅ Removed redundant admin export/import endpoints
- ✅ Configurable API version prefix

### v2.1.0
- ✅ Complete statistic module with user data isolation
- ✅ Multi-format export (Excel, CSV, PDF) with binary file download
- ✅ Summary, breakdown, trends, and top expenses analytics
- ✅ Dashboard visualization endpoints for charts
- ✅ Import functionality with category auto-creation

### v2.0.0
- ✅ User authentication and authorization
- ✅ User data isolation for cash flows and categories
- ✅ Reorganized routes into /open and /admin
- ✅ Admin user management endpoints
- ✅ Backup/restore with user data support

### v1.0.0
- Basic cash flow and category CRUD
- No user isolation
- No authentication
