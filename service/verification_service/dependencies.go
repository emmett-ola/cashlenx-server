package verification_service

import (
	"crypto/rand"

	"github.com/macar-x/cashlenx-server/mapper/operation_confirm_code_mapper"
	"github.com/macar-x/cashlenx-server/model"
)

type codeRepository interface {
	CreateCode(code model.OperationConfirmCode) error
	GetCodeByToken(token string) model.OperationConfirmCode
	GetCodeByVerificationToken(token string) model.OperationConfirmCode
	GetActiveCodeByPurposeAndPayload(operationType string, payload string) model.OperationConfirmCode
	InvalidateActiveCodes(userId string, operationType string) error
	InvalidateActiveCodesByPurposeAndPayload(operationType string, payload string) error
	SetVerificationToken(id string, verificationToken string) error
	MarkCodeAsUsed(id string) error
	DeleteCode(id string) error
}

type mapperCodeRepository struct{}

func (mapperCodeRepository) CreateCode(code model.OperationConfirmCode) error {
	return operation_confirm_code_mapper.INSTANCE.CreateCode(code)
}

func (mapperCodeRepository) GetCodeByToken(token string) model.OperationConfirmCode {
	return operation_confirm_code_mapper.INSTANCE.GetCodeByToken(token)
}

func (mapperCodeRepository) GetCodeByVerificationToken(token string) model.OperationConfirmCode {
	return operation_confirm_code_mapper.INSTANCE.GetCodeByVerificationToken(token)
}

func (mapperCodeRepository) GetActiveCodeByPurposeAndPayload(operationType string, payload string) model.OperationConfirmCode {
	return operation_confirm_code_mapper.INSTANCE.GetActiveCodeByPurposeAndPayload(operationType, payload)
}

func (mapperCodeRepository) InvalidateActiveCodes(userId string, operationType string) error {
	return operation_confirm_code_mapper.INSTANCE.InvalidateActiveCodes(userId, operationType)
}

func (mapperCodeRepository) InvalidateActiveCodesByPurposeAndPayload(operationType string, payload string) error {
	return operation_confirm_code_mapper.INSTANCE.InvalidateActiveCodesByPurposeAndPayload(operationType, payload)
}

func (mapperCodeRepository) SetVerificationToken(id string, verificationToken string) error {
	return operation_confirm_code_mapper.INSTANCE.SetVerificationToken(id, verificationToken)
}

func (mapperCodeRepository) MarkCodeAsUsed(id string) error {
	return operation_confirm_code_mapper.INSTANCE.MarkCodeAsUsed(id)
}

func (mapperCodeRepository) DeleteCode(id string) error {
	return operation_confirm_code_mapper.INSTANCE.DeleteCode(id)
}

var (
	codeRepo   codeRepository = mapperCodeRepository{}
	randomRead                = rand.Read
)
