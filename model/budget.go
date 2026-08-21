package model

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BudgetEntity stores a monthly spending limit. Spent amounts are derived from
// cash flows so the persisted limit can never drift from the ledger.
type BudgetEntity struct {
	Id            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BelongsUserId primitive.ObjectID `json:"belongs_user_id" bson:"belongs_user_id"`
	CategoryId    primitive.ObjectID `json:"category_id" bson:"category_id"`
	Period        string             `json:"period" bson:"period"`
	LimitAmount   float64            `json:"limit_amount" bson:"limit_amount"`
	BaseEntity    `bson:",inline"`
}

func (entity BudgetEntity) IsEmpty() bool { return reflect.DeepEqual(entity, BudgetEntity{}) }

type UpsertBudgetRequest struct {
	CategoryId  string  `json:"category_id"`
	Period      string  `json:"period"`
	LimitAmount float64 `json:"limit_amount"`
}

type BudgetView struct {
	Id           string  `json:"id"`
	CategoryId   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Period       string  `json:"period"`
	LimitAmount  float64 `json:"limit_amount"`
	SpentAmount  float64 `json:"spent_amount"`
	Remaining    float64 `json:"remaining"`
	Progress     float64 `json:"progress"`
}
