package operation_confirm_code_mapper

import (
	"testing"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestOperationConfirmCodeDocumentOmitsEmptyVerificationToken(t *testing.T) {
	doc := operationConfirmCodeDocument(model.OperationConfirmCode{}, primitive.NewObjectID())

	if _, exists := doc["verification_token"]; exists {
		t.Fatal("verification_token must be absent while it is empty")
	}
}

func TestOperationConfirmCodeDocumentIncludesVerificationToken(t *testing.T) {
	const token = "verification-token"
	doc := operationConfirmCodeDocument(model.OperationConfirmCode{VerificationToken: token}, primitive.NewObjectID())

	if got := doc["verification_token"]; got != token {
		t.Fatalf("verification_token = %v, want %q", got, token)
	}
}
