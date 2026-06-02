package admin_cmd

import (
	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/macar-x/cashlenx-server/service/user_service"
)

var (
	requireAdminSession  = cli_auth.RequireAdmin
	createAdminUser      = user_service.CreateService
	getAdminUser         = user_service.GetUserByObjectId
	listAdminUsers       = user_service.GetAllUsers
	countAdminUsers      = user_service.CountAllUsers
	updateAdminUser      = user_service.UpdateService
	deleteAdminUser      = user_service.DeleteService
	adminDumpDatabase    = manage_service.AdminDumpDatabase
	adminRestoreDatabase = manage_service.AdminRestoreDatabase
)
