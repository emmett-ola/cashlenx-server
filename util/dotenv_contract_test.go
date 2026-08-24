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
MONGO_PORT=27017
MONGO_DB_URI=mongodb://${MONGO_ROOT_USERNAME}:${MONGO_ROOT_PASSWORD}@localhost:${MONGO_PORT}/${DB_NAME}?authSource=admin
MYSQL_USER=cashlenx
MYSQL_PASSWORD=single-definition
MYSQL_PORT=3306
MYSQL_DB_URI=${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(localhost:${MYSQL_PORT})
`)
	if err != nil {
		t.Fatalf("parse dotenv contract: %v", err)
	}

	want := "mongodb://cashlenx:single-definition@localhost:27017/cashlenx?authSource=admin"
	if got := values["MONGO_DB_URI"]; got != want {
		t.Fatalf("expanded MONGO_DB_URI = %q, want %q", got, want)
	}

	mySQLWant := "cashlenx:single-definition@tcp(localhost:3306)"
	if got := values["MYSQL_DB_URI"]; got != mySQLWant {
		t.Fatalf("expanded MYSQL_DB_URI = %q, want %q", got, mySQLWant)
	}
}
