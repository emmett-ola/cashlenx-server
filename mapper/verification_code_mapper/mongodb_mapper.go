package verification_code_mapper

import (
	"context"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VerificationCodeMongoMapper struct{}

func (mapper *VerificationCodeMongoMapper) CreateCode(code model.OperationConfirmCode) error {
	collection := database.GetMongoCollection(code.CollectionName())

	// Convert ID string to ObjectID if needed, or let Mongo generate it
	// But usually ID is generated before calling create
	var id primitive.ObjectID
	var err error
	if code.Id != "" {
		id, err = primitive.ObjectIDFromHex(code.Id)
		if err != nil {
			id = primitive.NewObjectID()
		}
	} else {
		id = primitive.NewObjectID()
	}

	// Prepare document
	doc := bson.M{
		"_id":            id,
		"user_id":        code.UserId,
		"code":           code.Code,
		"operation_type": code.OperationType,
		"payload":        code.Payload,
		"expires_time":   code.ExpiresTime,
		"used_time":      code.UsedTime,
		"create_user_id": code.CreateUserId,
		"create_time":    code.CreateTime,
		"update_user_id": code.UpdateUserId,
		"update_time":    code.UpdateTime,
		"delete_user_id": code.DeleteUserId,
		"delete_time":    code.DeleteTime,
		"is_delete":      code.IsDelete,
	}

	_, err = collection.InsertOne(context.TODO(), doc)
	return err
}

func (mapper *VerificationCodeMongoMapper) GetCodeByToken(token string) model.OperationConfirmCode {
	collection := database.GetMongoCollection(model.OperationConfirmCode{}.CollectionName())

	filter := bson.M{
		"code":      token,
		"is_delete": false,
	}

	var result bson.M
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return model.OperationConfirmCode{}
	}

	return mapper.mapBsonToEntity(result)
}

func (mapper *VerificationCodeMongoMapper) InvalidateActiveCodes(userId string, operationType string) error {
	collection := database.GetMongoCollection(model.OperationConfirmCode{}.CollectionName())

	filter := bson.M{
		"user_id":        userId,
		"operation_type": operationType,
		"is_delete":      false,
		"used_time":      nil,
	}

	update := bson.M{
		"$set": bson.M{
			"is_delete":   true,
			"delete_time": time.Now(),
		},
	}

	_, err := collection.UpdateMany(context.TODO(), filter, update)
	return err
}

func (mapper *VerificationCodeMongoMapper) MarkCodeAsUsed(id string) error {
	collection := database.GetMongoCollection(model.OperationConfirmCode{}.CollectionName())

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectId}
	update := bson.M{
		"$set": bson.M{
			"used_time":   time.Now(),
			"update_time": time.Now(),
		},
	}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (mapper *VerificationCodeMongoMapper) DeleteCode(id string) error {
	collection := database.GetMongoCollection(model.OperationConfirmCode{}.CollectionName())

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectId}
	update := bson.M{
		"$set": bson.M{
			"is_delete":   true,
			"delete_time": time.Now(),
		},
	}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	return err
}

func (mapper *VerificationCodeMongoMapper) mapBsonToEntity(data bson.M) model.OperationConfirmCode {
	var entity model.OperationConfirmCode

	if id, ok := data["_id"].(primitive.ObjectID); ok {
		entity.Id = id.Hex()
	}

	if val, ok := data["user_id"].(string); ok {
		entity.UserId = val
	}
	if val, ok := data["code"].(string); ok {
		entity.Code = val
	}
	if val, ok := data["operation_type"].(string); ok {
		entity.OperationType = val
	}
	if val, ok := data["payload"].(string); ok {
		entity.Payload = val
	}

	// Handle time fields
	if val, ok := data["expires_time"].(primitive.DateTime); ok {
		entity.ExpiresTime = val.Time()
	} else if val, ok := data["expires_time"].(time.Time); ok {
		entity.ExpiresTime = val
	}

	if val, ok := data["used_time"].(primitive.DateTime); ok {
		t := val.Time()
		entity.UsedTime = &t
	} else if val, ok := data["used_time"].(time.Time); ok {
		entity.UsedTime = &val
	}

	// Base entity fields
	if val, ok := data["create_user_id"].(primitive.ObjectID); ok {
		entity.CreateUserId = val
	}
	if val, ok := data["create_time"].(primitive.DateTime); ok {
		entity.CreateTime = val.Time()
	} else if val, ok := data["create_time"].(time.Time); ok {
		entity.CreateTime = val
	}

	if val, ok := data["update_user_id"].(primitive.ObjectID); ok {
		entity.UpdateUserId = val
	}
	if val, ok := data["update_time"].(primitive.DateTime); ok {
		entity.UpdateTime = val.Time()
	} else if val, ok := data["update_time"].(time.Time); ok {
		entity.UpdateTime = val
	}

	if val, ok := data["is_delete"].(bool); ok {
		entity.IsDelete = val
	}

	return entity
}
