package category_cmd

import (
	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
	"github.com/macar-x/cashlenx-server/service/category_service"
)

var (
	requireCategoryUser          = cli_auth.RequireUserID
	createCategoryForUser        = category_service.CreateForUser
	queryCategoriesForUser       = category_service.QueryAllForUser
	queryCategoryByIDForUser     = category_service.QueryByIdForUser
	queryCategoryByNameForUser   = category_service.QueryByNameForUser
	queryCategoryChildrenForUser = category_service.GetChildCategoriesForUser
	queryCategoryTreeForUser     = category_service.GetCategoryTreeByUser
	updateCategoryForUser        = category_service.UpdateByIdForUser
	deleteCategoryForUser        = category_service.DeleteByIdForUser
)
