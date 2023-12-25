package proxy

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuthPlugin struct {
	mongoClient *mongo.Client
	dbName      string
	collName    string
	apiTokenKey string
}

func NewAuthPlugin(mongoURI string, dbName string, collName string, apiTokenKey string) *AuthPlugin {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	return &AuthPlugin{
		mongoClient: client,
		dbName:      dbName,
		collName:    collName,
		apiTokenKey: apiTokenKey,
	}
}

func (ap *AuthPlugin) Auth(apiToken string) bool {
	result := ap.mongoClient.Database(ap.dbName).Collection(ap.collName).FindOne(context.TODO(), bson.M{
		"apiToken": apiToken,
	})
	result.Decode(&struct{}{}) // ignore the result
	if result.Err() != nil {
		log.Printf("Auth failed: %s", result.Err())
		return false
	}

	return true
}
