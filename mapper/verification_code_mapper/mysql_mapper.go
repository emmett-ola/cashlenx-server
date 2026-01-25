package verification_code_mapper

import (
	"github.com/macar-x/cashlenx-server/model"
)

// VerificationCodeMySQLMapper implements VerificationCodeMapper for MySQL
// Currently just a placeholder as the project seems to primarily use MongoDB
type VerificationCodeMySQLMapper struct{}

func (mapper *VerificationCodeMySQLMapper) CreateCode(code model.OperationConfirmCode) error {
	// TODO: Implement MySQL logic
	return nil
}

func (mapper *VerificationCodeMySQLMapper) GetCodeByToken(token string) model.OperationConfirmCode {
	// TODO: Implement MySQL logic
	return model.OperationConfirmCode{}
}

func (mapper *VerificationCodeMySQLMapper) InvalidateActiveCodes(userId string, operationType string) error {
	// TODO: Implement MySQL logic
	return nil
}

func (mapper *VerificationCodeMySQLMapper) MarkCodeAsUsed(id string) error {
	// TODO: Implement MySQL logic
	return nil
}

func (mapper *VerificationCodeMySQLMapper) DeleteCode(id string) error {
	// TODO: Implement MySQL logic
	return nil
}
