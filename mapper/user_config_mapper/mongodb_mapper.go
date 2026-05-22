package user_config_mapper

import (
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserConfigurationMongoDbMapper struct{}

func convertBsonM2UserConfigurationEntity(bsonData bson.M) model.UserConfigurationEntity {
	if bsonData == nil {
		return model.UserConfigurationEntity{}
	}

	var config model.UserConfigurationEntity
	bsonBytes, err := bson.Marshal(bsonData)
	if err != nil {
		util.Logger.Errorln(err)
		return model.UserConfigurationEntity{}
	}
	if err = bson.Unmarshal(bsonBytes, &config); err != nil {
		util.Logger.Errorln(err)
		return model.UserConfigurationEntity{}
	}

	return config
}

func convertUserConfigurationEntity2BsonD(config model.UserConfigurationEntity) bson.D {
	return bson.D{
		primitive.E{Key: "_id", Value: config.Id},
		primitive.E{Key: "belongs_user_id", Value: config.BelongsUserId},
		primitive.E{Key: "display_language", Value: config.DisplayLanguage},
		primitive.E{Key: "currency_code", Value: config.CurrencyCode},
		primitive.E{Key: "active_theme_color", Value: config.ActiveThemeColor},
		primitive.E{Key: "create_user_id", Value: config.CreateUserId},
		primitive.E{Key: "create_time", Value: config.CreateTime},
		primitive.E{Key: "update_user_id", Value: config.UpdateUserId},
		primitive.E{Key: "update_time", Value: config.UpdateTime},
		primitive.E{Key: "delete_user_id", Value: config.DeleteUserId},
		primitive.E{Key: "delete_time", Value: config.DeleteTime},
		primitive.E{Key: "is_delete", Value: config.IsDelete},
	}
}

func convertUserConfigurationEntity2UpdateBsonD(config model.UserConfigurationEntity) bson.D {
	return bson.D{
		primitive.E{Key: "belongs_user_id", Value: config.BelongsUserId},
		primitive.E{Key: "display_language", Value: config.DisplayLanguage},
		primitive.E{Key: "currency_code", Value: config.CurrencyCode},
		primitive.E{Key: "active_theme_color", Value: config.ActiveThemeColor},
		primitive.E{Key: "create_user_id", Value: config.CreateUserId},
		primitive.E{Key: "create_time", Value: config.CreateTime},
		primitive.E{Key: "update_user_id", Value: config.UpdateUserId},
		primitive.E{Key: "update_time", Value: config.UpdateTime},
		primitive.E{Key: "delete_user_id", Value: config.DeleteUserId},
		primitive.E{Key: "delete_time", Value: config.DeleteTime},
		primitive.E{Key: "is_delete", Value: config.IsDelete},
	}
}

func (m UserConfigurationMongoDbMapper) GetByUserId(plainUserId string) model.UserConfigurationEntity {
	userId := util.Convert2ObjectId(plainUserId)
	if plainUserId == "" || userId == primitive.NilObjectID {
		util.Logger.Warnln("user configuration's user id is not acceptable")
		return model.UserConfigurationEntity{}
	}

	filter := bson.D{
		primitive.E{Key: "belongs_user_id", Value: userId},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.UserConfigurationTableName)
	defer database.CloseMongoDbConnection()
	return convertBsonM2UserConfigurationEntity(database.GetOneInMongoDB(filter))
}

func (m UserConfigurationMongoDbMapper) InsertByEntity(newEntity model.UserConfigurationEntity) string {
	now := time.Now()
	if newEntity.Id.IsZero() {
		newEntity.Id = primitive.NewObjectID()
	}
	if newEntity.CreateTime.IsZero() {
		newEntity.CreateTime = now
	}
	if newEntity.UpdateTime.IsZero() {
		newEntity.UpdateTime = now
	}

	database.OpenMongoDbConnection(database.UserConfigurationTableName)
	defer database.CloseMongoDbConnection()
	newId := database.InsertOneInMongoDB(convertUserConfigurationEntity2BsonD(newEntity))
	return newId.Hex()
}

func (m UserConfigurationMongoDbMapper) UpdateByUserId(plainUserId string, updatedEntity model.UserConfigurationEntity) model.UserConfigurationEntity {
	userId := util.Convert2ObjectId(plainUserId)
	if plainUserId == "" || userId == primitive.NilObjectID {
		util.Logger.Warnln("user configuration's user id is not acceptable")
		return model.UserConfigurationEntity{}
	}

	updatedEntity.UpdateTime = time.Now()
	filter := bson.D{
		primitive.E{Key: "belongs_user_id", Value: userId},
	}

	database.OpenMongoDbConnection(database.UserConfigurationTableName)
	defer database.CloseMongoDbConnection()
	database.UpdateManyInMongoDB(filter, convertUserConfigurationEntity2UpdateBsonD(updatedEntity))

	return m.GetByUserId(plainUserId)
}

func (m UserConfigurationMongoDbMapper) GetAllIncludeDeleted(limit, offset int) []model.UserConfigurationEntity {
	filter := bson.D{}

	database.OpenMongoDbConnection(database.UserConfigurationTableName)
	defer database.CloseMongoDbConnection()

	var configs []model.UserConfigurationEntity
	results := database.GetManyInMongoDBWithPaginationIncludeDeleted(filter, int64(limit), int64(offset))
	for _, result := range results {
		configs = append(configs, convertBsonM2UserConfigurationEntity(result))
	}

	return configs
}

func (m UserConfigurationMongoDbMapper) CountAll() int64 {
	filter := bson.D{}

	database.OpenMongoDbConnection(database.UserConfigurationTableName)
	defer database.CloseMongoDbConnection()

	return database.CountInMongoDB(filter)
}

func (m UserConfigurationMongoDbMapper) Truncate() error {
	database.OpenMongoDbConnection(database.UserConfigurationTableName)
	defer database.CloseMongoDbConnection()

	deletedCount := database.DeleteManyInMongoDBIncludeDeleted(bson.D{})
	util.Logger.Infow("User configurations truncated successfully", "deleted_count", deletedCount)
	return nil
}
