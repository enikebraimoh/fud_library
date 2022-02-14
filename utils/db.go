package utils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDBHandle struct {
	client *mongo.Client
}

var defaultMongoHandle = &MongoDBHandle{}

var once sync.Once

type errChecker struct {
	err error
}

func (e *errChecker) Check(err error) {
	if e.err != nil {
		return
	}

	e.err = err
}

func ConnectToDB(clusterURL string) error {
	var ec errChecker

	once.Do(func() {
		ec.Check(defaultMongoHandle.Connect(clusterURL))
		ec.Check(CreateUniqueIndex("users", "email", 1))
	})

	return ec.err
}

func (mh *MongoDBHandle) Connect(clusterURL string) error {
	clientOptions := options.Client().ApplyURI(clusterURL)

	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return err
	}

	timeOutFactor := 3
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeOutFactor)*time.Second)

	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return err
	}

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		return err
	}

	mh.client = client

	return nil
}

func (mh *MongoDBHandle) GetCollection(collectionName string) *mongo.Collection {
	DBName := Env("DB_NAME")
	return mh.client.Database(DBName).Collection(collectionName)
}

// GetCollection return collection for the db in DB_NAME env variable.
func GetCollection(collectionName string) *mongo.Collection {
	return defaultMongoHandle.GetCollection(collectionName)
}

func (mh *MongoDBHandle) Client() *mongo.Client {
	return mh.client
}

func GetMongoDBCollection(dbname, collectionName string) (*mongo.Collection, error) {
	client := defaultMongoHandle.Client()

	collection := client.Database(dbname).Collection(collectionName)

	return collection, nil
}

// get MongoDb documents for a collection.
func GetMongoDBDocs(collectionName string, filter map[string]interface{}, opts ...*options.FindOptions) ([]bson.M, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	var data []bson.M

	filterCursor, err := collection.Find(ctx, MapToBson(filter), opts...)
	if err != nil {
		return nil, err
	}

	if err := filterCursor.All(ctx, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// get single MongoDb document for a collection.
func GetMongoDBDoc(collectionName string, filter map[string]interface{}, opts ...*options.FindOneOptions) (bson.M, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)

	var data bson.M
	if err := collection.FindOne(ctx, MapToBson(filter), opts...).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func CreateMongoDBDoc(collectionName string, data map[string]interface{}) (*mongo.InsertOneResult, error) {
	ctx := context.Background()
	collection := defaultMongoHandle.GetCollection(collectionName)
	res, err := collection.InsertOne(ctx, MapToBson(data))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func CreateUniqueIndex(collName, field string, order int) error {
	collection := defaultMongoHandle.GetCollection(collName)

	indexModel := mongo.IndexModel{
		Keys:    bson.M{field: order},
		Options: options.Index().SetUnique(true),
	}

	indexName, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		fmt.Printf("error creating unique index on %s field in %s: %v", field, collName, err)
		return nil
	}

	fmt.Printf("%s index on %s collection created successfully\n", indexName, collName)

	return nil
}

// Update single MongoDb document for a collection.
func UpdateOneMongoDBDoc(collectionName, id string, data map[string]interface{}) (*mongo.UpdateResult, error) {
	ctx := context.Background()
	opts := options.Update().SetUpsert(true)
	collection := defaultMongoHandle.GetCollection(collectionName)

	_id, _ := primitive.ObjectIDFromHex(id)

	filter := bson.M{"_id": _id}

	// updateOne sets the fields, without using $set the entire document will be overwritten
	updateData := bson.M{"$set": MapToBson(data)}
	res, err := collection.UpdateOne(ctx, filter, updateData, opts)

	if err != nil {
		return nil, err
	}

	return res, nil
}
