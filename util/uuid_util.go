package util

import "go.mongodb.org/mongo-driver/bson/primitive"

func Convert2ObjectId(plainId string) primitive.ObjectID {
	objectId, err := primitive.ObjectIDFromHex(plainId)
	if err != nil {
		Logger.Warnln(err.Error())
	}
	return objectId
}
