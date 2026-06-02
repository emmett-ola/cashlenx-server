package user_controller

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
)

var (
	getProfileUser    = user_service.GetService
	updateProfileUser = user_service.UpdateProfileService
)

type updateProfileUserFunc func(string, model.UserProfileUpdateRequest) (model.UserEntity, error)
