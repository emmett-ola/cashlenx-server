# API vs CLI Feature Parity

**Last Updated**: 2025-12-28
**Purpose**: Track feature parity between API and CLI interfaces

## Current Implementation Status

### ✅ Public Features (No Authentication)

| Feature | API Endpoint | CLI Command | Status |
|---------|-------------|-------------|---------|
| Health Check | `GET /api/open/health` | `cashlenx open health` | ✅ Both |
| Version Info | `GET /api/open/version` | `cashlenx open version` | ✅ Both |
| Server Start | N/A | `cashlenx open start` | ✅ CLI only |
| User Login | `POST /api/open/auth/login` | N/A | ✅ API only |
| User Registration | `POST /api/open/auth/register` | N/A | ✅ API only |

### ✅ User Features (Authentication Required)

#### Cash Flow Operations
| Feature | API Endpoint | CLI Command | Status | User Isolation |
|---------|-------------|-------------|---------|----------------|
| Create Income | `POST /api/cash/income` | `cashlenx cash income` | ✅ Both | ✅ Yes |
| Create Expense | `POST /api/cash/expense` | `cashlenx cash expense` | ✅ Both | ✅ Yes |
| Query by ID | `GET /api/cash/{id}` | `cashlenx cash query -i {id}` | ✅ Both | ✅ Yes |
| Query by Date | `GET /api/cash/date/{date}` | `cashlenx cash query -b {date}` | ✅ Both | ✅ Yes |
| Update by ID | `PUT /api/cash/{id}` | `cashlenx cash update` | ✅ Both | ✅ Yes |
| Delete by ID | `DELETE /api/cash/{id}` | `cashlenx cash delete -i {id}` | ✅ Both | ✅ Yes |
| Delete by Date | `DELETE /api/cash/date/{date}` | `cashlenx cash delete -b {date}` | ✅ Both | ✅ Yes |
| List All | `GET /api/cash?limit=N&offset=M` | `cashlenx cash list` | ✅ Both | ✅ Yes |
| Query Range | `GET /api/cash/range?from=X&to=Y` | `cashlenx cash range` | ✅ Both | ✅ Yes |
| Monthly Summary | `GET /api/cash/summary/monthly/{yyyymm}` | `cashlenx cash summary` | ✅ Both | ✅ Yes |

#### Category Operations
| Feature | API Endpoint | CLI Command | Status | User Isolation |
|---------|-------------|-------------|---------|----------------|
| Create Category | `POST /api/category` | `cashlenx category create` | ✅ Both | ✅ Yes |
| List Categories | `GET /api/category?limit=N&offset=M` | `cashlenx category list` | ✅ Both | ✅ Yes |
| Query by ID | `GET /api/category/{id}` | `cashlenx category query -i {id}` | ✅ Both | ✅ Yes |
| Query by Name | `GET /api/category/name/{name}` | `cashlenx category query -n {name}` | ✅ Both | ✅ Yes |
| Get Child Categories | `GET /api/category/{id}/children` | `cashlenx category query -p {id}` | ✅ Both | ✅ Yes |
| Update Category | `PUT /api/category/{id}` | `cashlenx category update` | ✅ Both | ✅ Yes |
| Delete Category | `DELETE /api/category/{id}` | `cashlenx category delete` | ✅ Both | ✅ Yes |
| Get Category Tree | `GET /api/category/tree` | `cashlenx category tree` | ✅ Both | ✅ Yes |

#### ✅ Statistic & Analytics Operations (NEW in v2.0.0)

**Export/Import**
| Feature | API Endpoint | CLI Command | Status | User Isolation |
|---------|-------------|-------------|---------|----------------|
| Export to Excel | `GET /api/statistic/export?format=xlsx` | `cashlenx statistic export -o file.xlsx` | ✅ Both | ✅ Yes |
| Export to CSV | `GET /api/statistic/export?format=csv` | `cashlenx statistic export -o file.csv` | ✅ Both | ✅ Yes |
| Export to PDF | `GET /api/statistic/export?format=pdf` | `cashlenx statistic export -o file.pdf` | ✅ Both | ✅ Yes |
| Import from Excel | `POST /api/statistic/import?file_path=X` | `cashlenx statistic import -i file.xlsx` | ✅ Both | ✅ Yes |

**Summaries**
| Feature | API Endpoint | CLI Command | Status | User Isolation |
|---------|-------------|-------------|---------|----------------|
| Daily Summary | `GET /api/statistic/summary/daily/{date}` | `cashlenx statistic summary -p daily -d {date}` | ✅ Both | ✅ Yes |
| Monthly Summary | `GET /api/statistic/summary/monthly/{month}` | `cashlenx statistic summary -p monthly -d {month}` | ✅ Both | ✅ Yes |
| Yearly Summary | `GET /api/statistic/summary/yearly/{year}` | `cashlenx statistic summary -p yearly -d {year}` | ✅ Both | ✅ Yes |

**Analytics**
| Feature | API Endpoint | CLI Command | Status | User Isolation |
|---------|-------------|-------------|---------|----------------|
| Daily Breakdown | `GET /api/statistic/breakdown/daily/{date}` | `cashlenx statistic breakdown -p daily -d {date}` | ✅ Both | ✅ Yes |
| Monthly Breakdown | `GET /api/statistic/breakdown/monthly/{month}` | `cashlenx statistic breakdown -p monthly -d {month}` | ✅ Both | ✅ Yes |
| Yearly Breakdown | `GET /api/statistic/breakdown/yearly/{year}` | `cashlenx statistic breakdown -p yearly -d {year}` | ✅ Both | ✅ Yes |
| Daily Trends | `GET /api/statistic/trends/daily/{date}` | `cashlenx statistic trends -p daily -d {date}` | ✅ Both | ✅ Yes |
| Monthly Trends | `GET /api/statistic/trends/monthly/{month}` | `cashlenx statistic trends -p monthly -d {month}` | ✅ Both | ✅ Yes |
| Yearly Trends | `GET /api/statistic/trends/yearly/{year}` | `cashlenx statistic trends -p yearly -d {year}` | ✅ Both | ✅ Yes |
| Top Daily Expenses | `GET /api/statistic/top/daily/{date}?limit=N` | `cashlenx statistic top -p daily -d {date} -n N` | ✅ Both | ✅ Yes |
| Top Monthly Expenses | `GET /api/statistic/top/monthly/{month}?limit=N` | `cashlenx statistic top -p monthly -d {month} -n N` | ✅ Both | ✅ Yes |
| Top Yearly Expenses | `GET /api/statistic/top/yearly/{year}?limit=N` | `cashlenx statistic top -p yearly -d {year} -n N` | ✅ Both | ✅ Yes |

**Dashboard Visualizations (API Only)**
| Feature | API Endpoint | Status | User Isolation | Notes |
|---------|-------------|---------|----------------|-------|
| Dashboard Overview | `GET /api/statistic/dashboard/{period}/{date}` | ✅ API only | ✅ Yes | Chart-ready data |
| Income vs Expense Chart | `GET /api/statistic/chart/income-expense/{period}/{date}` | ✅ API only | ✅ Yes | Time-series data |
| Category Distribution | `GET /api/statistic/chart/category-distribution/{period}/{date}` | ✅ API only | ✅ Yes | Pie/donut chart data |
| Monthly Comparison | `GET /api/statistic/chart/monthly-comparison/{year}` | ✅ API only | ✅ Yes | Bar chart data |
| Spending Heatmap | `GET /api/statistic/chart/spending-heatmap/{year}` | ✅ API only | ✅ Yes | Calendar heatmap |

### ✅ Admin Features (Admin Only)

| Feature | API Endpoint | CLI Command | Status | Notes |
|---------|-------------|-------------|---------|-------|
| Create User | `POST /api/admin/user` | N/A | ✅ API only | Admin management |
| List Users | `GET /api/admin/user` | N/A | ✅ API only | Admin management |
| Query User | `GET /api/admin/user/{id}` | N/A | ✅ API only | Admin management |
| Update User | `PUT /api/admin/user/{id}` | N/A | ✅ API only | Admin management |
| Delete User | `DELETE /api/admin/user/{id}` | N/A | ✅ API only | Admin management |
| Database Backup | `GET /api/admin/manage/dump` | `cashlenx admin backup` | ✅ Both | Creates backup file |
| Database Restore | `POST /api/admin/manage/restore` | `cashlenx admin restore` | ✅ Both | Restores from backup |
| Export All Users | `GET /api/admin/manage/export` | `cashlenx admin export` | ⚠️ Deprecated | Use `/api/statistic/export` |
| Import All Users | `POST /api/admin/manage/import` | `cashlenx admin import` | ⚠️ Deprecated | Use `/api/statistic/import` |

### ❌ Removed Features (Rarely Used)

| Feature | Reason | Alternative |
|---------|--------|-------------|
| DB Connect | Rarely needed | Use application logs |
| DB Seed/Init | Development only | Manual data creation |
| DB Stats | Rarely used | Use monitoring tools |
| DB Reset/Truncate | Dangerous, rarely used | Manual database operations |
| DB Indexes | Rarely needed | Migrations handle this |

## Architecture Notes

### User Data Isolation (Implemented)
- **Three-layer architecture**: Mapper → Service → Controller
- **Mapper layer**: `*AndUser()` methods enforce database-level isolation
- **Service layer**: `*ForUser()` methods provide business logic with user context
- **Controller layer**: Extract `userId` from JWT and pass to services
- **Applied to**: Cash flows, Categories, Statistics, Backup/Restore

### Route Organization (Implemented)
- **Public routes**: `/api/open/*` - No authentication required
- **Admin routes**: `/api/admin/*` - Admin role required
- **User routes**: `/api/cash/*`, `/api/category/*`, `/api/statistic/*` - Authentication required, user-specific data

### CLI Organization (Implemented)
- **Public commands**: `cashlenx open` - No server needed (health/version need server, start doesn't)
- **Admin commands**: `cashlenx admin` - Admin privileges required
- **User commands**: `cashlenx cash`, `cashlenx category`, `cashlenx statistic` - Authentication required

## Export Format Support

The statistic export feature supports multiple formats with automatic file download:

| Format | Content-Type | Extension | Use Case |
|--------|-------------|-----------|----------|
| Excel | `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` | .xlsx | Full-featured spreadsheet with multiple sheets |
| CSV | `text/csv` | .csv | Simple format, import to any spreadsheet app |
| PDF | `application/pdf` | .pdf | Print-friendly professional report |

**API Behavior**: Returns binary file for direct download (no file_path needed)
**CLI Behavior**: Saves to specified file path

## Testing Checklist

### User Isolation Testing
- [x] User A cannot access User B's cash flows
- [x] User A cannot access User B's categories
- [x] User A can only export their own data
- [x] User A can only import to their own account
- [x] Backup includes all users' data (admin only)
- [x] Restore properly isolates data by user
- [x] Statistics show only user's own data
- [x] Dashboard shows only user's own analytics

### Feature Parity Testing
- [x] All API endpoints have CLI equivalents (or documented reason why not)
- [x] All CLI commands have API equivalents (or documented reason why not)
- [x] Response formats are consistent between API and CLI
- [x] Error handling is consistent between API and CLI
- [x] Export formats work in both API and CLI
- [x] All analytics functions work in both API and CLI

## Known Gaps

### API-only Features
- User management (login, register, user CRUD) - Makes sense, CLI users are already authenticated
- Dashboard visualization endpoints - Designed for web/mobile frontends with charting libraries

### CLI-only Features
- Server start command - Makes sense, can't start server via API

## Version History

### v2.1.0 (Current)
- ✅ Implemented complete statistic module with user data isolation
- ✅ Added multi-format export (Excel, CSV, PDF)
- ✅ Binary file download for API exports
- ✅ Dashboard visualization endpoints for charts
- ✅ Comprehensive analytics: summary, breakdown, trends, top expenses
- ✅ Feature parity between API and CLI for all user-facing features

### v2.0.0
- ✅ Implemented user data isolation for cash flows and categories
- ✅ Reorganized API routes into /open and /admin
- ✅ Reorganized CLI into open and admin structure
- ✅ Removed rarely-used commands
- ✅ Added backup/restore with user support

### v1.0.0 (Previous)
- Basic cash flow and category CRUD
- No user isolation
- No authentication
