package budget_mapper

import (
	"context"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoMapper struct{}

func budgetCollection() *mongo.Collection {
	return database.GetMongoCollection(database.BudgetTableName)
}

func (MongoMapper) Insert(entity model.BudgetEntity) (model.BudgetEntity, error) {
	if entity.Id.IsZero() {
		entity.Id = primitive.NewObjectID()
	}
	_, err := budgetCollection().InsertOne(context.Background(), entity)
	return entity, err
}

func (MongoMapper) ListByUserAndPeriod(userID primitive.ObjectID, period string) ([]model.BudgetEntity, error) {
	filter := bson.M{"belongs_user_id": userID, "period": period, "is_delete": false}
	cursor, err := budgetCollection().Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	items := []model.BudgetEntity{}
	if err := cursor.All(context.Background(), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (MongoMapper) GetByIDAndUser(id, userID primitive.ObjectID) (model.BudgetEntity, error) {
	var entity model.BudgetEntity
	err := budgetCollection().FindOne(context.Background(), bson.M{"_id": id, "belongs_user_id": userID, "is_delete": false}).Decode(&entity)
	if err == mongo.ErrNoDocuments {
		return model.BudgetEntity{}, nil
	}
	return entity, err
}

func (MongoMapper) GetByScope(userID, categoryID primitive.ObjectID, period string) (model.BudgetEntity, error) {
	var entity model.BudgetEntity
	err := budgetCollection().FindOne(context.Background(), bson.M{"belongs_user_id": userID, "category_id": categoryID, "period": period, "is_delete": false}).Decode(&entity)
	if err == mongo.ErrNoDocuments {
		return model.BudgetEntity{}, nil
	}
	return entity, err
}

func (MongoMapper) Update(entity model.BudgetEntity) (model.BudgetEntity, error) {
	result, err := budgetCollection().UpdateOne(context.Background(), bson.M{"_id": entity.Id, "belongs_user_id": entity.BelongsUserId, "is_delete": false}, bson.M{"$set": bson.M{"category_id": entity.CategoryId, "period": entity.Period, "limit_amount": entity.LimitAmount, "update_user_id": entity.UpdateUserId, "update_time": entity.UpdateTime}})
	if err != nil || result.MatchedCount != 1 {
		return model.BudgetEntity{}, err
	}
	return entity, nil
}

func (MongoMapper) Delete(id, userID, actorID primitive.ObjectID) (bool, error) {
	now := time.Now().UTC()
	result, err := budgetCollection().UpdateOne(context.Background(), bson.M{"_id": id, "belongs_user_id": userID, "is_delete": false}, bson.M{"$set": bson.M{"is_delete": true, "delete_user_id": actorID, "delete_time": now}})
	return err == nil && result.ModifiedCount == 1, err
}
