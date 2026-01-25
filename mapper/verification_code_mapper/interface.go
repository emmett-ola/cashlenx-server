package verification_code_mapper

import (
	"github.com/macar-x/cashlenx-server/model"
)

var INSTANCE VerificationCodeMapper

type VerificationCodeMapper interface {
	// CreateCode creates a new verification code
	CreateCode(code model.OperationConfirmCode) error

	// GetCodeByToken retrieves a code by its token string
	GetCodeByToken(token string) model.OperationConfirmCode

	// InvalidateActiveCodes revokes all active codes for a user and operation type
	InvalidateActiveCodes(userId string, operationType string) error

	// MarkCodeAsUsed marks a specific code as used
	MarkCodeAsUsed(id string) error
	
	// DeleteCode physically deletes a code (optional, mainly for cleanup)
	DeleteCode(id string) error
}
