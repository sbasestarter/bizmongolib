package mongolib

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongo(dsn string) (client *mongo.Client, clientOps *options.ClientOptions, err error) {
	if !strings.HasPrefix(dsn, "mongodb://") {
		dsn = "mongodb://" + dsn
	}

	clientOps = options.Client().ApplyURI(dsn)

	err = clientOps.Validate()
	if err != nil {
		return
	}

	client, err = mongo.Connect(context.Background(), clientOps)

	return
}
