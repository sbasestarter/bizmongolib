package mongolib

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongo(dsn string) (client *mongo.Client, err error) {
	if !strings.HasPrefix(dsn, "mongodb://") {
		dsn = "mongodb://" + dsn
	}

	clientOptions := options.Client().ApplyURI(dsn)

	err = clientOptions.Validate()
	if err != nil {
		return
	}

	client, err = mongo.Connect(context.Background(), clientOptions)

	return
}
