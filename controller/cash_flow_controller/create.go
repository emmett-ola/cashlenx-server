package cash_flow_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/macar-x/cashlenx-server/util"
)

func CreateExpense(w http.ResponseWriter, r *http.Request) {
	requestBody, err := validCashFlowDTO(r)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, err)
		return
	}

	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	cashFlowEntity, err := cash_flow_service.SaveExpense(requestBody.BelongsDate, requestBody.CategoryName, requestBody.Amount, requestBody.Description, userId)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.ComposeJSONResponse(w, http.StatusCreated, cashFlowEntity)
}

func CreateIncome(w http.ResponseWriter, r *http.Request) {
	requestBody, err := validCashFlowDTO(r)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, err)
		return
	}

	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	cashFlowEntity, err := cash_flow_service.SaveIncome(requestBody.BelongsDate, requestBody.CategoryName, requestBody.Amount, requestBody.Description, userId)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}
	util.ComposeJSONResponse(w, http.StatusCreated, cashFlowEntity)
}

func validCashFlowDTO(r *http.Request) (model.CashFlowDTO, error) {
	var requestBody model.CashFlowDTO
	err := util.ParseJSONRequest(r, &requestBody)
	if err != nil {
		return model.CashFlowDTO{}, errors.NewInvalidInputError("invalid request body")
	}

	if !cash_flow_service.IsExpenseRequiredFiledSatisfied(requestBody.CategoryName, requestBody.Amount) {
		return model.CashFlowDTO{}, errors.NewValidationError("some required fields are empty")
	}
	return requestBody, nil
}