package mongolib

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetDataID(ctx context.Context, db *mongo.Database, collectionName string) (n uint64, err error) {
	table := db.Collection("ids")

	var result struct {
		Name   string `json:"name" bson:"name"`
		NextID uint64 `json:"next_id" bson:"next_id"`
	}

	var retryCount int
RETRY:
	err = table.FindOneAndUpdate(
		ctx,
		bson.M{"name": collectionName},
		bson.M{"$inc": bson.M{"next_id": 1}}).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			_, err = table.InsertOne(ctx, bson.M{
				"name":    collectionName,
				"next_id": 1,
			})
			if err == nil {
				_, _ = table.Indexes().CreateOne(ctx, mongo.IndexModel{
					Keys: bson.D{
						{Key: "name", Value: 1},
					},
					Options: options.Index().SetUnique(true),
				})

				if retryCount == 0 {
					retryCount++

					goto RETRY
				}
			}
		}

		return
	}

	n = result.NextID

	return
}
