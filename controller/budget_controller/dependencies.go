package budget_controller

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/budget_service"
)

var (
	createBudget = budget_service.CreateForUser
	listBudgets  = budget_service.ListForUser
	getBudget    = budget_service.GetForUser
	updateBudget = budget_service.UpdateForUser
	deleteBudget = budget_service.DeleteForUser
)

type createBudgetFunc func(model.UpsertBudgetRequest, string) (model.BudgetView, error)
