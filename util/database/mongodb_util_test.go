package database

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func TestConnectMongoClientDoesNotReturnFailedClient(t *testing.T) {
	wantErr := errors.New("ping failed")
	client, err := connectMongoClient(context.Background(), "mongodb://localhost", func(context.Context, *mongo.Client) error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("connectMongoClient() error = %v, want %v", err, wantErr)
	}
	if client != nil {
		t.Fatal("connectMongoClient() returned a client after ping failure")
	}
}
