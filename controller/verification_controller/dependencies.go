package verification_controller

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/verification_service"
)

var (
	sendVerificationCode = verification_service.SendVerificationCode
	verifyCodeForPurpose = verification_service.VerifyCode
)

type verifyCodeForPurposeFunc func(string, string, string) (model.VerifyVerificationCodeResponse, error)
