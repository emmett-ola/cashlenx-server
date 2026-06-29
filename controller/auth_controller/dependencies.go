package auth_controller

import "github.com/macar-x/cashlenx-server/service/user_service"

var (
	registerPublicUser = user_service.RegisterPublicUser
	getRegisteredUser  = user_service.GetUserByObjectId
)
