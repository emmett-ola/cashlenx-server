package util

import (
	"testing"

	"github.com/joho/godotenv"
)

func TestDotenvExpandsPreviouslyDefinedDatabaseValues(t *testing.T) {
	values, err := godotenv.Unmarshal(`
DB_NAME=cashlenx
MONGO_ROOT_USERNAME=cashlenx
MONGO_ROOT_PASSWORD=single-definition
MONGO_DB_URI=mongodb://${MONGO_ROOT_USERNAME}:${MONGO_ROOT_PASSWORD}@localhost:27017/${DB_NAME}?authSource=admin
`)
	if err != nil {
		t.Fatalf("parse dotenv contract: %v", err)
	}

	want := "mongodb://cashlenx:single-definition@localhost:27017/cashlenx?authSource=admin"
	if got := values["MONGO_DB_URI"]; got != want {
		t.Fatalf("expanded MONGO_DB_URI = %q, want %q", got, want)
	}
}
