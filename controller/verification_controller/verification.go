package verification_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

func SendCode(w http.ResponseWriter, r *http.Request) {
	var req model.SendVerificationCodeRequest
	if err := util.ParseJSONRequest(r, &req); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	if err := sendVerificationCode(req.Purpose, req.Email, util.GetClientIP(r)); err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "verification code sent",
	})
}

func VerifyCode(w http.ResponseWriter, r *http.Request) {
	var req model.VerifyVerificationCodeRequest
	if err := util.ParseJSONRequest(r, &req); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	response, err := verifyCodeForPurpose(req.Purpose, req.Email, req.Code)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, response)
}
