package userpass

import (
	"context"
	"time"

	"github.com/sbasestarter/bizinters/userinters/userpass"
	"github.com/sbasestarter/bizmongolib/mongolib"
	"github.com/sgostarter/i/l"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoUserPasswordModel(mongoCli *mongo.Client, dbName, collectionName string, logger l.Wrapper) userpass.UserPasswordModel {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	if mongoCli == nil {
		logger.Error("nilMongoCli")

		return nil
	}

	impl := &mongoUserPasswordModelImpl{
		logger:         logger.WithFields(l.StringField(l.ClsKey, "mongoUserPasswordModelImpl")),
		mongoCli:       mongoCli,
		dbName:         dbName,
		collectionName: collectionName,
	}

	impl.init()

	return impl
}

type mongoUserPasswordModelImpl struct {
	logger l.Wrapper

	mongoCli       *mongo.Client
	dbName         string
	collectionName string
	db             *mongo.Database
	collection     *mongo.Collection
}

func (impl *mongoUserPasswordModelImpl) init() {
	impl.db = impl.mongoCli.Database(impl.dbName)
	impl.collection = impl.db.Collection(impl.collectionName)

	if _, err := impl.collection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_name", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		impl.logger.WithFields(l.ErrorField(err)).Error("CreateUniqueIndexFailed")
	}
}

func (impl *mongoUserPasswordModelImpl) AddUser(ctx context.Context, userName, password string) (user *userpass.User, err error) {
	id, err := mongolib.GetDataID(ctx, impl.db, impl.collectionName)
	if err != nil {
		return
	}

	u := &User{
		ID:       id,
		UserName: userName,
		Password: password,
		CreateAt: time.Now().Unix(),
	}
	_, err = impl.collection.InsertOne(ctx, u)

	if err != nil {
		return
	}

	user = &userpass.User{
		ID:       id,
		UserName: u.UserName,
		Password: u.Password,
		CreateAt: u.CreateAt,
	}

	return
}

func (impl *mongoUserPasswordModelImpl) DeleteUser(ctx context.Context, userID uint64) error {
	_, err := impl.collection.DeleteOne(ctx, bson.M{
		"_id": userID,
	})

	return err
}

func (impl *mongoUserPasswordModelImpl) GetUser(ctx context.Context, userID uint64) (user *userpass.User, err error) {
	return impl.findUserOne(ctx, bson.M{
		"_id": userID,
	})
}

func (impl *mongoUserPasswordModelImpl) GetUserByUserName(ctx context.Context, userName string) (user *userpass.User, err error) {
	return impl.findUserOne(ctx, bson.M{
		"user_name": userName,
	})
}

func (impl *mongoUserPasswordModelImpl) ListUsers(ctx context.Context) (users []*userpass.User, err error) {
	cursor, err := impl.collection.Find(ctx, bson.D{})
	if err != nil {
		return
	}

	var us []*User

	err = cursor.All(ctx, &us)
	if err != nil {
		return
	}

	for _, u := range us {
		users = append(users, &userpass.User{
			ID:       u.ID,
			UserName: u.UserName,
			Password: u.Password,
			CreateAt: u.CreateAt,
			ExData:   u.ExData,
		})
	}

	return
}

func (impl *mongoUserPasswordModelImpl) findUserOne(ctx context.Context, m bson.M) (user *userpass.User, err error) {
	var u User

	err = impl.collection.FindOne(ctx, m).Decode(&u)
	if err != nil {
		return
	}

	user = &userpass.User{
		ID:       u.ID,
		UserName: u.UserName,
		Password: u.Password,
		CreateAt: u.CreateAt,
		ExData:   u.ExData,
	}

	return
}

func (impl *mongoUserPasswordModelImpl) UpdateUserExData(ctx context.Context, userID uint64, key string, val interface{}) (err error) {
	var u User

	err = impl.collection.FindOne(ctx, bson.M{
		"_id": userID,
	}).Decode(&u)
	if err != nil {
		return
	}

	if len(u.ExData) == 0 {
		u.ExData = make(map[string]interface{})
	}

	u.ExData[key] = val

	update := bson.M{
		"ex_data": u.ExData,
	}

	err = impl.collection.FindOneAndUpdate(ctx, bson.M{
		"_id": userID,
	}, bson.M{
		"$set": update,
	}).Err()

	return
}

func (impl *mongoUserPasswordModelImpl) UpdateUserAllExData(ctx context.Context, userID uint64, exData map[string]interface{}) (err error) {
	update := bson.M{
		"ex_data": exData,
	}

	err = impl.collection.FindOneAndUpdate(ctx, bson.M{
		"_id": userID,
	}, bson.M{
		"$set": update,
	}).Err()

	return
}
