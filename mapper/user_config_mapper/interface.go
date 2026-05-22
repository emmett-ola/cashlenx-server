package user_config_mapper

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

var INSTANCE UserConfigurationMapper

type UserConfigurationMapper interface {
	GetByUserId(plainUserId string) model.UserConfigurationEntity
	InsertByEntity(newEntity model.UserConfigurationEntity) string
	UpdateByUserId(plainUserId string, updatedEntity model.UserConfigurationEntity) model.UserConfigurationEntity
	GetAllIncludeDeleted(limit, offset int) []model.UserConfigurationEntity
	CountAll() int64
	Truncate() error
}

func init() {
	switch util.GetConfigByKey("db.type") {
	case "mongodb":
		INSTANCE = UserConfigurationMongoDbMapper{}
	case "mysql":
		INSTANCE = UserConfigurationMySqlMapper{}
	default:
		panic("database type not supported")
	}
}
