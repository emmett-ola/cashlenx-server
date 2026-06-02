package manage_controller

import (
	"github.com/macar-x/cashlenx-server/service/manage_service"
)

var (
	adminDumpDatabase    = manage_service.AdminDumpDatabase
	userExportData       = manage_service.UserExportData
	adminRestoreDatabase = manage_service.AdminRestoreDatabase
	userImportData       = manage_service.UserImportData
)
