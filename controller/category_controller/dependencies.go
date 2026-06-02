package category_controller

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/category_service"
)

var (
	createCategoryForUser       = category_service.CreateForUser
	queryCategoriesForUser      = category_service.QueryAllForUser
	queryCategoryByIDForUser    = category_service.QueryByIdForUser
	queryCategoryByNameForUser  = category_service.QueryByNameForUser
	queryChildCategoriesForUser = category_service.GetChildCategoriesForUser
	queryCategoryTreeForUser    = category_service.GetCategoryTreeByUser
	updateCategoryForUser       = category_service.UpdateByIdForUser
	deleteCategoryForUser       = category_service.DeleteByIdForUser
)

type queryCategoriesForUserFunc func(string, string, int, int) ([]model.CategoryEntity, int64, error)
