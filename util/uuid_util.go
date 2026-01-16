package util

import (
	"crypto/rand"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateUUID generates a random UUID
func GenerateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to ObjectID if UUID generation fails
		return primitive.NewObjectID().Hex()
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func Convert2ObjectId(plainId string) primitive.ObjectID {
	plainId = strings.TrimSpace(plainId)
	objectId, err := primitive.ObjectIDFromHex(plainId)
	if err != nil {
		// Logger.Warnln(err.Error())
		return primitive.NilObjectID
	}
	return objectId
}
