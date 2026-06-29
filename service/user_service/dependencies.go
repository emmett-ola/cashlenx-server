package user_service

import (
	"github.com/macar-x/cashlenx-server/mapper/user_config_mapper"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/macar-x/cashlenx-server/service/refresh_token_service"
	"github.com/macar-x/cashlenx-server/service/verification_service"
)

type userRepository interface {
	GetUserByObjectId(plainId string) model.UserEntity
	GetUserByUsername(username string) model.UserEntity
	GetUserByUsernameIncludeDeleted(username string) model.UserEntity
	GetUserByEmail(email string) model.UserEntity
	InsertUserByEntity(newEntity model.UserEntity) string
	UpdateUserByEntity(plainId string, updatedEntity model.UserEntity) model.UserEntity
	GetAllUsers(limit, offset int) []model.UserEntity
	GetUsersByRole(role string) []model.UserEntity
	CountAllUsers() int64
	DeleteUserByObjectId(plainId string) model.UserEntity
}

type mapperUserRepository struct{}

func (mapperUserRepository) GetUserByObjectId(plainId string) model.UserEntity {
	return user_mapper.INSTANCE.GetUserByObjectId(plainId)
}

func (mapperUserRepository) GetUserByUsername(username string) model.UserEntity {
	return user_mapper.INSTANCE.GetUserByUsername(username)
}

func (mapperUserRepository) GetUserByUsernameIncludeDeleted(username string) model.UserEntity {
	return user_mapper.INSTANCE.GetUserByUsernameIncludeDeleted(username)
}

func (mapperUserRepository) GetUserByEmail(email string) model.UserEntity {
	return user_mapper.INSTANCE.GetUserByEmail(email)
}

func (mapperUserRepository) InsertUserByEntity(newEntity model.UserEntity) string {
	return user_mapper.INSTANCE.InsertUserByEntity(newEntity)
}

func (mapperUserRepository) UpdateUserByEntity(plainId string, updatedEntity model.UserEntity) model.UserEntity {
	return user_mapper.INSTANCE.UpdateUserByEntity(plainId, updatedEntity)
}

func (mapperUserRepository) GetAllUsers(limit, offset int) []model.UserEntity {
	return user_mapper.INSTANCE.GetAllUsers(limit, offset)
}

func (mapperUserRepository) GetUsersByRole(role string) []model.UserEntity {
	return user_mapper.INSTANCE.GetUsersByRole(role)
}

func (mapperUserRepository) CountAllUsers() int64 {
	return user_mapper.INSTANCE.CountAllUsers()
}

func (mapperUserRepository) DeleteUserByObjectId(plainId string) model.UserEntity {
	return user_mapper.INSTANCE.DeleteUserByObjectId(plainId)
}

type userConfigurationRepository interface {
	GetByUserId(plainUserId string) model.UserConfigurationEntity
	InsertByEntity(newEntity model.UserConfigurationEntity) string
	UpdateByUserId(plainUserId string, updatedEntity model.UserConfigurationEntity) model.UserConfigurationEntity
}

type mapperUserConfigurationRepository struct{}

func (mapperUserConfigurationRepository) GetByUserId(plainUserId string) model.UserConfigurationEntity {
	return user_config_mapper.INSTANCE.GetByUserId(plainUserId)
}

func (mapperUserConfigurationRepository) InsertByEntity(newEntity model.UserConfigurationEntity) string {
	return user_config_mapper.INSTANCE.InsertByEntity(newEntity)
}

func (mapperUserConfigurationRepository) UpdateByUserId(plainUserId string, updatedEntity model.UserConfigurationEntity) model.UserConfigurationEntity {
	return user_config_mapper.INSTANCE.UpdateByUserId(plainUserId, updatedEntity)
}

var (
	userRepo                           userRepository              = mapperUserRepository{}
	userConfigurationRepo              userConfigurationRepository = mapperUserConfigurationRepository{}
	initializeDefaultCategoriesForUser                             = category_service.InitializeDefaultCategoriesForUser
	revokeAllRefreshTokens                                         = refresh_token_service.RevokeAllRefreshTokens
	consumeVerifiedToken                                           = verification_service.ConsumeVerifiedToken
	sendVerificationCode                                           = verification_service.SendVerificationCode
)
