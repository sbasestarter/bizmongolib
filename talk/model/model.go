package model

import (
	"context"
	"fmt"
	"time"

	"github.com/sbasestarter/bizinters/talkinters"
	"github.com/sbasestarter/bizmongolib/mongolib"
	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/i/l"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collectionTalkInfo     = "talk_info"
	collectionTalkTemplate = "talk:%s"
)

func NewMongoModel(dsn string, logger l.Wrapper) (m talkinters.Model, err error) {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	client, clientOptions, err := mongolib.InitMongo(dsn)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.WithFields(l.ErrorField(err)).Fatal("MongoPing")
	}

	return &mongoModelImpl{
		clientOps: clientOptions,
		mongoCli:  client,
	}, nil
}

type mongoModelImpl struct {
	clientOps *options.ClientOptions
	mongoCli  *mongo.Client
}

func (m *mongoModelImpl) CreateTalk(ctx context.Context, talkInfo *talkinters.TalkInfoW) (talkID string, err error) {
	r, err := m.mongoCli.Database(m.clientOps.Auth.AuthSource).Collection(collectionTalkInfo).InsertOne(ctx, talkInfo)
	if err != nil {
		return
	}

	if oid, ok := r.InsertedID.(primitive.ObjectID); ok {
		talkID = oid.Hex()
	}

	return
}

func (m *mongoModelImpl) OpenTalk(ctx context.Context, talkID string) (err error) {
	return m.updateTalkInfo(ctx, talkID, bson.M{
		"Status": talkinters.TalkStatusOpened,
	})
}

func (m *mongoModelImpl) CloseTalk(ctx context.Context, talkID string) (err error) {
	return m.updateTalkInfo(ctx, talkID, bson.M{
		"Status": talkinters.TalkStatusClosed,
	})
}

func (m *mongoModelImpl) AddTalkMessage(ctx context.Context, talkID string, message *talkinters.TalkMessageW) (err error) {
	_, err = m.mongoCli.Database(m.clientOps.Auth.AuthSource).Collection(m.talkCollectionKey(talkID)).InsertOne(ctx, message)

	return
}

func (m *mongoModelImpl) GetTalkMessages(ctx context.Context, talkID string, offset, count int64) (messages []*talkinters.TalkMessageR, err error) {
	findOptions := options.Find()
	if count > 0 {
		findOptions.SetSkip(offset)
		findOptions.SetLimit(count)
	}

	cursor, err := m.mongoCli.Database(m.clientOps.Auth.AuthSource).Collection(m.talkCollectionKey(talkID)).Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return
	}

	err = cursor.All(ctx, &messages)

	return
}

func (m *mongoModelImpl) QueryTalks(ctx context.Context, creatorID, serviceID uint64, talkID string,
	statuses []talkinters.TalkStatus) (talks []*talkinters.TalkInfoR, err error) {
	return m.queryTalksEx(ctx, creatorID, serviceID, talkID, statuses, nil)
}

func (m *mongoModelImpl) GetPendingTalkInfos(ctx context.Context) ([]*talkinters.TalkInfoR, error) {
	bsonM := bson.M{}
	bsonM["ServiceID"] = 0

	talkInfos, err := m.queryTalksEx(ctx, 0, 0, "", []talkinters.TalkStatus{talkinters.TalkStatusOpened}, bsonM)
	if err != nil {
		return nil, err
	}

	return talkInfos, nil
}

func (m *mongoModelImpl) UpdateTalkServiceID(ctx context.Context, talkID string, serviceID uint64) (err error) {
	return m.updateTalkInfo(ctx, talkID, bson.M{
		"ServiceID": serviceID,
	})
}

//
//
//

func (m *mongoModelImpl) queryTalkFilter(creatorID, serviceID uint64, talkID string, statuses []talkinters.TalkStatus) (filter bson.M, err error) {
	filter = bson.M{}
	if creatorID > 0 {
		filter["CreatorID"] = creatorID
	}

	if serviceID > 0 {
		filter["ServiceID"] = serviceID
	}

	if len(statuses) > 0 {
		filter["Status"] = bson.M{"$in": statuses}
	}

	if talkID != "" {
		var objectID primitive.ObjectID

		objectID, err = primitive.ObjectIDFromHex(talkID)
		if err != nil {
			err = commerr.ErrInvalidArgument

			return
		}

		filter["_id"] = objectID
	}

	return
}

func (m *mongoModelImpl) queryTalksEx(ctx context.Context, creatorID, serviceID uint64, talkID string,
	statuses []talkinters.TalkStatus, bsonM bson.M) (talks []*talkinters.TalkInfoR, err error) {
	collection := m.mongoCli.Database(m.clientOps.Auth.AuthSource).Collection(collectionTalkInfo)

	filter, err := m.queryTalkFilter(creatorID, serviceID, talkID, statuses)
	if err != nil {
		err = commerr.ErrInvalidArgument

		return
	}

	for k, v := range bsonM {
		filter[k] = v
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return
	}

	err = cursor.All(ctx, &talks)

	return
}

func (m *mongoModelImpl) updateTalkInfo(ctx context.Context, talkID string, updateMap bson.M) (err error) {
	objectID, err := primitive.ObjectIDFromHex(talkID)
	if err != nil {
		return
	}

	r := m.mongoCli.Database(m.clientOps.Auth.AuthSource).Collection(collectionTalkInfo).FindOneAndUpdate(ctx,
		bson.M{
			"_id": objectID,
		}, bson.M{
			"$set": updateMap,
		})

	err = r.Err()

	return
}

func (m *mongoModelImpl) talkCollectionKey(talkID string) string {
	return fmt.Sprintf(collectionTalkTemplate, talkID)
}
