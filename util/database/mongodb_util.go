package database

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient   *mongo.Client
	mongoDatabase *mongo.Database
	collection    *mongo.Collection
)

// InitMongoDbConnection initializes the MongoDB connection pool (called once at startup)
func InitMongoDbConnection() error {
	once.Do(initMongoDbConnection)
	if defaultDatabaseUri == "" {
		return errors.New("environment value 'MONGO_DB_URI' not set")
	}

	if mongoClient != nil {
		return nil // Already initialized
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := connectMongoClient(ctx, defaultDatabaseUri, func(ctx context.Context, client *mongo.Client) error {
		return client.Ping(ctx, nil)
	})
	if err != nil {
		return err
	}

	mongoClient = client
	mongoDatabase = mongoClient.Database(defaultDatabaseName)
	isConnected = true
	util.Logger.Info("MongoDB connection pool initialized")
	return nil
}

func connectMongoClient(ctx context.Context, uri string, ping func(context.Context, *mongo.Client) error) (*mongo.Client, error) {
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(50).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	if err := ping(ctx, client); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return client, nil
}

// GetMongoCollection returns a collection from the connection pool
func GetMongoCollection(collectionName string) *mongo.Collection {
	if mongoClient == nil || mongoDatabase == nil {
		if err := InitMongoDbConnection(); err != nil {
			log.Fatal("Failed to initialize MongoDB connection:", err)
		}
	}
	return mongoDatabase.Collection(collectionName)
}

// OpenMongoDbConnection sets the current collection (for backward compatibility)
// Deprecated: Use GetMongoCollection instead
func OpenMongoDbConnection(collectionName string) {
	collection = GetMongoCollection(collectionName)
	util.Logger.Debug("Using MongoDB collection: ", collectionName)
}

// CloseMongoDbConnection is now a no-op for backward compatibility
// The connection pool stays open for the lifetime of the application
// Use ShutdownMongoDbConnection() for actual shutdown
func CloseMongoDbConnection() {
	// No-op: connection pool remains open
}

// ShutdownMongoDbConnection closes the MongoDB connection pool (called only on application shutdown)
func ShutdownMongoDbConnection() {
	if mongoClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mongoClient.Disconnect(ctx); err != nil {
		util.Logger.Errorw("Failed to close MongoDB connection", "error", err)
		return
	}

	mongoClient = nil
	mongoDatabase = nil
	isConnected = false
	util.Logger.Info("MongoDB connection pool closed")
}

func GetOneInMongoDB(filter bson.D) bson.M {
	// Append is_delete: false
	filter = append(filter, primitive.E{Key: "is_delete", Value: false})
	return GetOneInMongoDBIncludeDeleted(filter)
}

func GetOneInMongoDBIncludeDeleted(filter bson.D) bson.M {
	checkDbConnection()

	var resultInBson bson.M
	err := collection.FindOne(context.TODO(), filter).Decode(&resultInBson)

	// Handle query failure
	if errors.Is(err, mongo.ErrNoDocuments) {
		// Logger.Warnln("record does not exist")
	} else if err != nil {
		log.Fatal(err)
	}

	return resultInBson
}

func GetManyInMongoDB(filter bson.D) []bson.M {
	// Append is_delete: false
	filter = append(filter, primitive.E{Key: "is_delete", Value: false})
	return GetManyInMongoDBIncludeDeleted(filter)
}

func GetManyInMongoDBIncludeDeleted(filter bson.D) []bson.M {
	checkDbConnection()

	var resultInBsonArray []bson.M
	cursor, err := collection.Find(context.TODO(), filter)

	// Handle query failure
	if errors.Is(err, mongo.ErrNoDocuments) {
		// Logger.Warnln("record does not exist")
	} else if err != nil {
		log.Fatal(err)
	}

	if err2 := cursor.All(context.TODO(), &resultInBsonArray); err2 != nil {
		log.Fatal(err2)
	}

	return resultInBsonArray
}

// GetManyInMongoDBWithPagination retrieves multiple documents with pagination support
func GetManyInMongoDBWithPagination(filter bson.D, limit int64, offset int64) []bson.M {
	// Append is_delete: false
	filter = append(filter, primitive.E{Key: "is_delete", Value: false})
	return GetManyInMongoDBWithPaginationIncludeDeleted(filter, limit, offset)
}

func GetManyInMongoDBWithPaginationIncludeDeleted(filter bson.D, limit int64, offset int64) []bson.M {
	checkDbConnection()

	var resultInBsonArray []bson.M
	options := options.Find().SetSkip(offset).SetLimit(limit)
	cursor, err := collection.Find(context.TODO(), filter, options)

	// Handle query failure
	if errors.Is(err, mongo.ErrNoDocuments) {
		// Logger.Warnln("record does not exist")
	} else if err != nil {
		log.Fatal(err)
	}

	if err2 := cursor.All(context.TODO(), &resultInBsonArray); err2 != nil {
		log.Fatal(err2)
	}

	return resultInBsonArray
}

func CountInMongoDB(filter bson.D) int64 {
	// Append is_delete: false
	filter = append(filter, primitive.E{Key: "is_delete", Value: false})
	return CountInMongoDBIncludeDeleted(filter)
}

func CountInMongoDBIncludeDeleted(filter bson.D) int64 {
	checkDbConnection()

	result, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		panic(err)
	}

	return result
}

func CountInMongoDBWithError(filter bson.D) (int64, error) {
	// Append is_delete: false
	filter = append(filter, primitive.E{Key: "is_delete", Value: false})
	return CountInMongoDBWithErrorIncludeDeleted(filter)
}

func CountInMongoDBWithErrorIncludeDeleted(filter bson.D) (int64, error) {
	checkDbConnection()

	result, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func InsertOneInMongoDB(data bson.D) primitive.ObjectID {
	checkDbConnection()

	/* result:
	 *	type InsertOneResult struct {
	 *		InsertedID primitive.ObjectID
	 *	}
	 */
	result, err := collection.InsertOne(context.TODO(), data)
	if err != nil {
		panic(err)
	}

	return result.InsertedID.(primitive.ObjectID)
}

func UpdateManyInMongoDB(filter, data bson.D) int64 {
	// Append is_delete: false
	filter = append(filter, primitive.E{Key: "is_delete", Value: false})
	return UpdateManyInMongoDBIncludeDeleted(filter, data)
}

func UpdateManyInMongoDBIncludeDeleted(filter, data bson.D) int64 {
	checkDbConnection()

	updateData := bson.D{
		primitive.E{Key: "$set", Value: data},
	}
	// Upsert disable by default.
	result, err := collection.UpdateMany(context.TODO(), filter, updateData)
	if err != nil {
		panic(err)
	}

	return result.ModifiedCount
}

func DeleteManyInMongoDB(filter bson.D) int64 {
	// Append is_delete: false
	filter = append(filter, primitive.E{Key: "is_delete", Value: false})
	return DeleteManyInMongoDBIncludeDeleted(filter)
}

func DeleteManyInMongoDBIncludeDeleted(filter bson.D) int64 {
	checkDbConnection()

	result, err := collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		panic(err)
	}

	return result.DeletedCount
}

// GetMongoDbCollection returns the current MongoDB collection for advanced operations
func GetMongoDbCollection() *mongo.Collection {
	checkDbConnection()
	return collection
}
