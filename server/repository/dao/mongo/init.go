package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sw5005-sus/ceramicraft-comment-mservice/server/config"
)

var (
	CommentCollection *mongo.Collection
)

func Init() {
	url := fmt.Sprintf("mongodb://%s:%d", config.Config.MongoConfig.Host, config.Config.MongoConfig.Port)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(url))
	if err != nil {
		panic(err)
	}
	database := client.Database(config.Config.MongoConfig.Database)
	CommentCollection = database.Collection("comments")
}
