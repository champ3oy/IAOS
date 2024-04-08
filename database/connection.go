package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client *mongo.Client
	ctx    = context.Background()
)

var dbName = "IssueReporting"

// Connect initializes a connection to MongoDB
func Connect() error {
	clientOptions := options.Client().ApplyURI("mongodb+srv://deadalus:Samarithan2293%40@cluster0.nwde7lt.mongodb.net/IssueReporting?retryWrites=true&w=majority")

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	log.Println("Connected to MongoDB!")

	Client = client
	return nil
}

// InsertOne inserts a single document into the specified collection
func InsertOne(collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	collection := Client.Database(dbName).Collection(collectionName)
	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindOne finds a single document in the specified collection that matches the filter
func FindOne(collectionName string, filter interface{}) *mongo.SingleResult {
	collection := Client.Database(dbName).Collection(collectionName)
	return collection.FindOne(ctx, filter)
}

// FindOne finds a single document in the specified collection that matches the filter
func Find(collectionName string, filter interface{}) (*mongo.Cursor, error) {
	collection := Client.Database(dbName).Collection(collectionName)
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	return cursor, nil
}

// UpdateOne updates a single document in the specified collection that matches the filter
func UpdateOne(collectionName string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	collection := Client.Database(dbName).Collection(collectionName)
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteOne deletes a single document from the specified collection that matches the filter
func DeleteOne(collectionName string, filter interface{}) (*mongo.DeleteResult, error) {
	collection := Client.Database(dbName).Collection(collectionName)
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindOneAndUpdate finds a single document in the specified collection that matches the filter and updates it
func FindOneAndUpdate(collectionName string, filter interface{}, update interface{}) *mongo.SingleResult {
	collection := Client.Database(dbName).Collection(collectionName)
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	return collection.FindOneAndUpdate(ctx, filter, update, opts)
}
