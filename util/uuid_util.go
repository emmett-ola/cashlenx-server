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
		return primitive.NilObjectID
	}
	return objectId
}

// GenerateObjectId generates a new ObjectID in hex string format
// Uses Snowflake algorithm for better uniqueness guarantees
func GenerateObjectId() string {
	id, err := GenerateSnowflakeIDString()
	if err != nil {
		// Fallback to MongoDB ObjectID if Snowflake fails
		Logger.Warnw("Snowflake ID generation failed, using MongoDB ObjectID fallback", "error", err)
		return primitive.NewObjectID().Hex()
	}
	return id
}

// GenerateObjectIdWithFallback generates an ID with fallback to ensure it never fails
func GenerateObjectIdWithFallback() string {
	id := GenerateObjectId()
	if id == "" {
		// Ultimate fallback
		return primitive.NewObjectID().Hex()
	}
	return id
}
