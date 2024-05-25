package auth

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoAuthPlugin struct {
	mongoClient *mongo.Client
	dbName      string
	collName    string
	apiTokenKey string
	whitelist   []string
}

func NewMongoAuthPlugin(mongoURI, dbName, collName, apiTokenKey string, whitelist []string) *MongoAuthPlugin {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	return &MongoAuthPlugin{
		mongoClient: client,
		dbName:      dbName,
		collName:    collName,
		apiTokenKey: apiTokenKey,
		whitelist:   whitelist,
	}
}

func (ap *MongoAuthPlugin) Auth(apiToken string) bool {
	if ap.whitelist != nil {
		for _, token := range ap.whitelist {
			if token == apiToken {
				return true
			}
		}
	}

	result := ap.mongoClient.Database(ap.dbName).Collection(ap.collName).FindOne(context.TODO(), bson.M{
		ap.apiTokenKey: apiToken,
	})
	result.Decode(&struct{}{}) // ignore the result
	if result.Err() != nil {
		log.Printf("Auth failed: %s", result.Err())
		return false
	}

	return true
}
