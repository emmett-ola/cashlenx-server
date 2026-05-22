package operation_confirm_code_mapper

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

var INSTANCE OperationConfirmCodeMapper

type OperationConfirmCodeMapper interface {
	// CreateCode creates a new operation confirmation code
	CreateCode(code model.OperationConfirmCode) error

	// GetCodeByToken retrieves a code by its token string
	GetCodeByToken(token string) model.OperationConfirmCode

	// GetCodeByVerificationToken retrieves a code row by its one-time verification token
	GetCodeByVerificationToken(token string) model.OperationConfirmCode

	// GetActiveCodeByPurposeAndPayload retrieves the current active code for a purpose/payload pair
	GetActiveCodeByPurposeAndPayload(operationType string, payload string) model.OperationConfirmCode

	// InvalidateActiveCodes revokes all active codes for a user and operation type
	InvalidateActiveCodes(userId string, operationType string) error

	// InvalidateActiveCodesByPurposeAndPayload revokes active codes for a purpose/payload pair
	InvalidateActiveCodesByPurposeAndPayload(operationType string, payload string) error

	// SetVerificationToken stores a one-time verification token on an existing code row
	SetVerificationToken(id string, verificationToken string) error

	// MarkCodeAsUsed marks a specific code as used
	MarkCodeAsUsed(id string) error

	// DeleteCode physically deletes a code (optional, mainly for cleanup)
	DeleteCode(id string) error
}

func init() {
	switch util.GetConfigByKey("db.type") {
	case "mongodb":
		INSTANCE = &OperationConfirmCodeMongoMapper{}
	case "mysql":
		INSTANCE = &OperationConfirmCodeMySQLMapper{}
	default:
		panic("database type not supported")
	}
}
