package open_cmd

import (
	"github.com/macar-x/cashlenx-server/auth"
	"github.com/macar-x/cashlenx-server/auth/provider"
	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
	"github.com/macar-x/cashlenx-server/controller"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/refresh_token_service"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/service/verification_service"
)

var (
	authenticateOpenUser       = auth.Service.Authenticate
	refreshOpenToken           = auth.Service.RefreshToken
	validateOpenAccessToken    = auth.Service.ValidateToken
	saveOpenSession            = cli_auth.Save
	currentOpenSession         = cli_auth.CurrentSession
	clearOpenSession           = cli_auth.Clear
	registerOpenUser           = user_service.RegisterPublicUser
	requestOpenPasswordReset   = user_service.RequestPasswordReset
	confirmOpenPasswordReset   = user_service.ConfirmPasswordReset
	getOpenRefreshToken        = refresh_token_service.GetRefreshTokenByToken
	revokeOpenRefreshToken     = refresh_token_service.RevokeRefreshToken
	revokeAllOpenRefreshTokens = refresh_token_service.RevokeAllRefreshTokens
	sendOpenVerificationCode   = verification_service.SendVerificationCode
	verifyOpenCode             = verification_service.VerifyCode
	startOpenServer            = controller.StartServer
)

type authenticateOpenUserFunc func(string, string, string, string, string, string) (string, string, model.UserEntity, error)
type refreshOpenTokenFunc func(string, string, string, string, string) (string, string, model.UserEntity, error)
type validateOpenAccessTokenFunc func(string) (*provider.Claims, error)
