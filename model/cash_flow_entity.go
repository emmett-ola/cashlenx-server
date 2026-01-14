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
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	UserId      primitive.ObjectID `json:"user_id" bson:"user_id"`
	CategoryId  primitive.ObjectID `json:"category_id" bson:"category_id"`
	BelongsDate time.Time          `json:"belongs_date" bson:"belongs_date"`
	FlowType    string             `json:"flow_type" bson:"flow_type"`
	Amount      float64            `json:"amount" bson:"amount"`
	Description string             `json:"description" bson:"description"`
	Remark      string             `json:"remark" bson:"remark"`
	CreateTime  time.Time          `json:"create_time" bson:"create_time"`
	ModifyTime  time.Time          `json:"modify_time" bson:"modify_time"`
}

func (entity CashFlowEntity) IsEmpty() bool {
	return reflect.DeepEqual(entity, CashFlowEntity{})
}

func (entity CashFlowEntity) ToString() string {
	return "[ " +
		"Id: " + entity.Id.Hex() +
		", Date: " + util.FormatDateToStringWithoutDash(entity.BelongsDate) +
		", Type: " + entity.FlowType +
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
		case "UserId":
			newEntity.UserId = util.Convert2ObjectId(value)
		case "CategoryId":
			newEntity.CategoryId = util.Convert2ObjectId(value)
		case "BelongsDate":
			newEntity.BelongsDate = util.FormatDateFromStringWithoutDash(value)
		case "FlowType":
			newEntity.FlowType = value
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
	localModifyTime := util.ToTimezone(entity.ModifyTime)

	// Create a temporary struct with local timezone timestamps
	type Alias CashFlowEntity
	return json.Marshal(&struct {
		BelongsDate time.Time `json:"belongs_date"`
		CreateTime  time.Time `json:"create_time"`
		ModifyTime  time.Time `json:"modify_time"`
		*Alias
	}{
		BelongsDate: localBelongsDate,
		CreateTime:  localCreateTime,
		ModifyTime:  localModifyTime,
		Alias:       (*Alias)(&entity),
	})
}
