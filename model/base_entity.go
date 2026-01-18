package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BaseEntity struct {
	CreateUserId primitive.ObjectID `json:"create_user_id" bson:"create_user_id"`
	CreateTime   time.Time          `json:"create_time" bson:"create_time"`
	UpdateUserId primitive.ObjectID `json:"update_user_id" bson:"update_user_id"`
	UpdateTime   time.Time          `json:"update_time" bson:"update_time"`
	DeleteUserId *primitive.ObjectID `json:"-" bson:"delete_user_id"`
	DeleteTime   *time.Time          `json:"-" bson:"delete_time"`
	IsDelete     bool                `json:"-" bson:"is_delete"`
}
