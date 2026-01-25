package model

// FlowType constants for cash flow types
const (
	FlowTypeIncome  = "income"
	FlowTypeExpense = "expense"
)

// DateFormat constants
const (
	DateFormatYYYYMMDD     = "20060102"
	DateFormatYYYYMMDDDash = "2006-01-02"
	DateFormatYYYYMM       = "2006-01"
	DateFormatYYYY         = "2006"
)

// Database table names
const (
	TableCashFlow = "cash_flow"
	TableCategory = "categories"
	TableUser     = "users"
)

// UserRole constants
const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)

// Gender constants
const (
	GenderMale   = "male"
	GenderFemale = "female"
	GenderOthers = "others"
)

// AuthConstants for JWT and authentication
const (
	JWTExpirationHours = 24
	JWTAlgorithm       = "HS256"
)
