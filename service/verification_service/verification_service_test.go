package verification_service

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
)

func TestGenerateVerificationTokenStoresCodeAndInvalidatesPriorActiveCodes(t *testing.T) {
	repo := installVerificationTestDeps(t)
	userID := "507f1f77bcf86cd799439011"
	repo.codesByToken["old-token"] = model.OperationConfirmCode{
		Id:            "old-id",
		UserId:        userID,
		Code:          "old-token",
		OperationType: string(OperationPasswordReset),
		ExpiresTime:   time.Now().Add(time.Hour),
	}

	token, err := GenerateVerificationToken(userID, OperationPasswordReset, "payload")
	if err != nil {
		t.Fatalf("GenerateVerificationToken returned error: %v", err)
	}

	code := repo.codesByToken[token]
	if code.Id == "" {
		t.Fatal("expected generated token to be stored")
	}
	if code.UserId != userID {
		t.Fatalf("UserId = %q, want %q", code.UserId, userID)
	}
	if code.OperationType != string(OperationPasswordReset) {
		t.Fatalf("OperationType = %q", code.OperationType)
	}
	if code.Payload != "payload" {
		t.Fatalf("Payload = %q, want payload", code.Payload)
	}
	if !repo.usedIDs["old-id"] {
		t.Fatal("expected prior active code to be invalidated")
	}
}

func TestGenerateVerificationTokenReturnsRandomReadError(t *testing.T) {
	installVerificationTestDeps(t)
	randomRead = func([]byte) (int, error) {
		return 0, io.ErrUnexpectedEOF
	}

	_, err := GenerateVerificationToken("507f1f77bcf86cd799439011", OperationPasswordReset, "")
	if err == nil {
		t.Fatal("expected random read error")
	}
}

func TestVerifyTokenChecksOperationExpiryAndUsedState(t *testing.T) {
	repo := installVerificationTestDeps(t)
	activeToken := "active-token"
	repo.codesByToken[activeToken] = model.OperationConfirmCode{
		Id:            "active-id",
		UserId:        "user-1",
		Code:          activeToken,
		OperationType: string(OperationEmailChange),
		Payload:       "new@example.test",
		ExpiresTime:   time.Now().Add(time.Hour),
	}

	payload, userID, valid := VerifyToken(activeToken, OperationEmailChange)
	if !valid {
		t.Fatal("expected active token to be valid")
	}
	if payload != "new@example.test" || userID != "user-1" {
		t.Fatalf("payload/userID = %q/%q", payload, userID)
	}

	_, _, valid = VerifyToken(activeToken, OperationPasswordReset)
	if valid {
		t.Fatal("expected token with wrong operation to be invalid")
	}

	usedTime := time.Now()
	repo.codesByToken["used-token"] = model.OperationConfirmCode{
		Id:            "used-id",
		Code:          "used-token",
		OperationType: string(OperationEmailChange),
		ExpiresTime:   time.Now().Add(time.Hour),
		UsedTime:      &usedTime,
	}
	_, _, valid = VerifyToken("used-token", OperationEmailChange)
	if valid {
		t.Fatal("expected used token to be invalid")
	}

	repo.codesByToken["expired-token"] = model.OperationConfirmCode{
		Id:            "expired-id",
		Code:          "expired-token",
		OperationType: string(OperationEmailChange),
		ExpiresTime:   time.Now().Add(-time.Minute),
	}
	_, _, valid = VerifyToken("expired-token", OperationEmailChange)
	if valid {
		t.Fatal("expected expired token to be invalid")
	}
	if !repo.deletedIDs["expired-id"] {
		t.Fatal("expected expired token to be deleted")
	}
}

func TestInvalidateTokenMarksCodeUsed(t *testing.T) {
	repo := installVerificationTestDeps(t)
	repo.codesByToken["token"] = model.OperationConfirmCode{
		Id:            "token-id",
		Code:          "token",
		OperationType: string(OperationPasswordReset),
		ExpiresTime:   time.Now().Add(time.Hour),
	}

	InvalidateToken("token")

	if !repo.usedIDs["token-id"] {
		t.Fatal("expected token to be marked used")
	}
}

func TestInvalidateTokensByUserAndOperationMarksMatchingActiveCodesUsed(t *testing.T) {
	repo := installVerificationTestDeps(t)
	repo.codesByToken["matching-token"] = model.OperationConfirmCode{
		Id:            "matching-id",
		UserId:        "user-1",
		Code:          "matching-token",
		OperationType: string(OperationPasswordReset),
		ExpiresTime:   time.Now().Add(time.Hour),
	}
	repo.codesByToken["other-token"] = model.OperationConfirmCode{
		Id:            "other-id",
		UserId:        "user-2",
		Code:          "other-token",
		OperationType: string(OperationPasswordReset),
		ExpiresTime:   time.Now().Add(time.Hour),
	}

	InvalidateTokensByUserAndOperation("user-1", OperationPasswordReset)

	if !repo.usedIDs["matching-id"] {
		t.Fatal("expected matching code to be used")
	}
	if repo.usedIDs["other-id"] {
		t.Fatal("expected other user's code not to be used")
	}
}

func installVerificationTestDeps(t *testing.T) *codeRepoStub {
	t.Helper()

	originalRepo := codeRepo
	originalRandomRead := randomRead
	stub := &codeRepoStub{
		codesByToken: map[string]model.OperationConfirmCode{},
		usedIDs:      map[string]bool{},
		deletedIDs:   map[string]bool{},
	}
	codeRepo = stub
	randomRead = func(p []byte) (int, error) {
		for i := range p {
			p[i] = byte(i + 1)
		}
		return len(p), nil
	}

	t.Cleanup(func() {
		codeRepo = originalRepo
		randomRead = originalRandomRead
	})
	return stub
}

type codeRepoStub struct {
	codesByToken map[string]model.OperationConfirmCode
	usedIDs      map[string]bool
	deletedIDs   map[string]bool
	createErr    error
}

func (stub *codeRepoStub) CreateCode(code model.OperationConfirmCode) error {
	if stub.createErr != nil {
		return stub.createErr
	}
	stub.codesByToken[code.Code] = code
	return nil
}

func (stub *codeRepoStub) GetCodeByToken(token string) model.OperationConfirmCode {
	return stub.codesByToken[token]
}

func (stub *codeRepoStub) GetCodeByVerificationToken(token string) model.OperationConfirmCode {
	for _, code := range stub.codesByToken {
		if code.VerificationToken == token {
			return code
		}
	}
	return model.OperationConfirmCode{}
}

func (stub *codeRepoStub) GetActiveCodeByPurposeAndPayload(operationType string, payload string) model.OperationConfirmCode {
	for _, code := range stub.codesByToken {
		if code.OperationType == operationType && code.Payload == payload && code.UsedTime == nil && !code.IsDelete {
			return code
		}
	}
	return model.OperationConfirmCode{}
}

func (stub *codeRepoStub) InvalidateActiveCodes(userId string, operationType string) error {
	for token, code := range stub.codesByToken {
		if code.UserId == userId && code.OperationType == operationType && code.UsedTime == nil {
			usedTime := time.Now()
			code.UsedTime = &usedTime
			stub.codesByToken[token] = code
			stub.usedIDs[code.Id] = true
		}
	}
	return nil
}

func (stub *codeRepoStub) InvalidateActiveCodesByPurposeAndPayload(operationType string, payload string) error {
	for token, code := range stub.codesByToken {
		if code.OperationType == operationType && code.Payload == payload && code.UsedTime == nil {
			usedTime := time.Now()
			code.UsedTime = &usedTime
			code.IsDelete = true
			stub.codesByToken[token] = code
			stub.usedIDs[code.Id] = true
		}
	}
	return nil
}

func (stub *codeRepoStub) SetVerificationToken(id string, verificationToken string) error {
	for token, code := range stub.codesByToken {
		if code.Id == id {
			code.VerificationToken = verificationToken
			stub.codesByToken[token] = code
			return nil
		}
	}
	return errors.New("code not found")
}

func (stub *codeRepoStub) MarkCodeAsUsed(id string) error {
	for token, code := range stub.codesByToken {
		if code.Id == id {
			usedTime := time.Now()
			code.UsedTime = &usedTime
			stub.codesByToken[token] = code
			stub.usedIDs[id] = true
			return nil
		}
	}
	return errors.New("code not found")
}

func (stub *codeRepoStub) DeleteCode(id string) error {
	stub.deletedIDs[id] = true
	for token, code := range stub.codesByToken {
		if code.Id == id {
			delete(stub.codesByToken, token)
			return nil
		}
	}
	return nil
}
