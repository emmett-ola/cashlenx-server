package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/macar-x/cashlenx-server/util"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CashFlowEntity struct {
	Id            primitive.ObjectID `bson:"_id,omitempty"`
	BelongsUserId primitive.ObjectID `json:"belongs_user_id" bson:"belongs_user_id"`
	CategoryId    primitive.ObjectID `json:"category_id" bson:"category_id"`
	BelongsDate   time.Time          `json:"belongs_date" bson:"belongs_date"`
	// FlowType    string             `json:"flow_type" bson:"flow_type"` // Deprecated: Use CategoryType instead
	CategoryName string  `json:"category_name" bson:"-"`
	CategoryType string  `json:"category_type" bson:"-"`
	Amount       float64 `json:"amount" bson:"amount"`
	Description  string  `json:"description" bson:"description"`
	Remark       string  `json:"remark" bson:"remark"`
	BaseEntity   `bson:",inline"`
}

// CashFlowFilter defines filters for querying cash flows
type CashFlowFilter struct {
	UserId           primitive.ObjectID
	CashType         string    // "income" or "expense"
	CategoryId       string    // Hex string
	Description      string    // Fuzzy match
	ExactDescription string    // Exact match
	FromDate         time.Time // Inclusive
	ToDate           time.Time // Inclusive
	Limit            int
	Offset           int
}

func (entity CashFlowEntity) IsEmpty() bool {
	return reflect.DeepEqual(entity, CashFlowEntity{})
}

func (entity CashFlowEntity) ToString() string {
	return "[ " +
		"Id: " + entity.Id.Hex() +
		", Date: " + util.FormatDateToStringWithoutDash(entity.BelongsDate) +
		", Type: " + entity.CategoryType +
		", Amount: " + fmt.Sprintf("%.2f", entity.Amount) +
		", Description: " + entity.Description +
		" ]"
}

func (entity CashFlowEntity) Build(fieldMap map[string]string) CashFlowEntity {
	newEntity := entity
	for key, value := range fieldMap {
		switch key {
		case "Id":
			objectId, err := primitive.ObjectIDFromHex(value)
			if err != nil {
				util.Logger.Warnln("build cash failed with err: " + err.Error())
			}
			newEntity.Id = objectId
		case "BelongsUserId":
			newEntity.BelongsUserId = util.Convert2ObjectId(value)
		case "CategoryId":
			newEntity.CategoryId = util.Convert2ObjectId(value)
		case "BelongsDate":
			newEntity.BelongsDate = util.FormatDateFromStringWithoutDash(value)
		// case "FlowType":
		// 	newEntity.FlowType = value
		case "Amount":
			amount, err := strconv.ParseFloat(value, 64)
			if err != nil {
				util.Logger.Warnln("build cash failed with err: " + err.Error())
			}
			newEntity.Amount = amount
		case "Description":
			newEntity.Description = value
		case "Remark":
			newEntity.Remark = value
		}
	}
	return newEntity
}

// MarshalJSON customizes JSON marshaling for CashFlowEntity
// Converts timestamps to the configured timezone for display
func (entity CashFlowEntity) MarshalJSON() ([]byte, error) {
	// Convert timestamps to configured timezone for display
	localBelongsDate := util.ToTimezone(entity.BelongsDate)
	localCreateTime := util.ToTimezone(entity.CreateTime)
	localUpdateTime := util.ToTimezone(entity.UpdateTime)

	// Create a temporary struct with local timezone timestamps
	type Alias CashFlowEntity
	return json.Marshal(&struct {
		BelongsDate time.Time `json:"belongs_date"`
		CreateTime  time.Time `json:"create_time"`
		UpdateTime  time.Time `json:"update_time"`
		*Alias
	}{
		BelongsDate: localBelongsDate,
		CreateTime:  localCreateTime,
		UpdateTime:  localUpdateTime,
		Alias:       (*Alias)(&entity),
	})
}
