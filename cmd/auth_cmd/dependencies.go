package auth_cmd

import (
	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
	"github.com/macar-x/cashlenx-server/service/refresh_token_service"
)

var (
	requireAuthUser      = cli_auth.RequireUser
	getAuthRefreshTokens = refresh_token_service.GetUserRefreshTokens
)
