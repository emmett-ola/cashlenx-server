package model

type CashFlowDTO struct {
	BelongsDate  string  `json:"belongs_date"`
	CategoryName string  `json:"category_name"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
}
