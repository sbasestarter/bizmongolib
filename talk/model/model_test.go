package model

import (
	"context"
	"testing"
	"time"

	"github.com/sbasestarter/bizinters/talkinters"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func Test1(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m, err := NewMongoModel("mongodb://mongo_default_user:mongo_default_pass@env:8309/my_db", nil)
	assert.Nil(t, err)

	m1, ok := m.(*mongoModelImpl)
	assert.True(t, ok)

	collectionNames, err := m1.mongoCli.Database(m1.clientOps.Auth.AuthSource).ListCollectionNames(
		context.Background(), bson.D{})
	assert.Nil(t, err)

	for _, name := range collectionNames {
		_ = m1.mongoCli.Database(m1.clientOps.Auth.AuthSource).Collection(name).Drop(context.TODO())
	}

	talkID, err := m.CreateTalk(ctx, &talkinters.TalkInfoW{
		Status:    talkinters.TalkStatusOpened,
		Title:     "testTalk1",
		StartAt:   time.Now().Unix(),
		CreatorID: 1,
	})
	assert.Nil(t, err)
	t.Log(talkID)

	err = m.AddTalkMessage(ctx, talkID, &talkinters.TalkMessageW{
		Type: talkinters.TalkMessageTypeText,
		Text: "talk_message_1",
	})
	assert.Nil(t, err)

	err = m.AddTalkMessage(ctx, talkID, &talkinters.TalkMessageW{
		Type: talkinters.TalkMessageTypeImage,
		Data: []byte("talk_image_1"),
	})
	assert.Nil(t, err)

	err = m.AddTalkMessage(ctx, talkID, &talkinters.TalkMessageW{
		Type: talkinters.TalkMessageTypeText,
		Text: "talk_message_3",
	})
	assert.Nil(t, err)

	err = m.AddTalkMessage(ctx, talkID, &talkinters.TalkMessageW{
		Type: talkinters.TalkMessageTypeText,
		Text: "talk_message_4",
	})
	assert.Nil(t, err)

	messages, err := m.GetTalkMessages(ctx, talkID, 0, 0)
	assert.Nil(t, err)
	assert.EqualValues(t, 4, len(messages))
	assert.EqualValues(t, "talk_message_1", messages[0].Text)
	assert.EqualValues(t, "talk_message_4", messages[3].Text)

	messages, err = m.GetTalkMessages(ctx, talkID, 0, 1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(messages))
	assert.EqualValues(t, "talk_message_1", messages[0].Text)

	messages, err = m.GetTalkMessages(ctx, talkID, 1, 2)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, len(messages))
	assert.EqualValues(t, "talk_message_3", messages[1].Text)

	err = m.CloseTalk(ctx, talkID)
	assert.Nil(t, err)

	err = m.OpenTalk(ctx, talkID)
	assert.Nil(t, err)

	err = m.OpenTalk(ctx, talkID[0:len(talkID)-1]+"9")
	assert.NotNil(t, err)

	talks, err := m.QueryTalks(ctx, 1, 0, "", nil)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(talks))
	assert.EqualValues(t, talkinters.TalkStatusOpened, talks[0].Status)

	talks, err = m.QueryTalks(ctx, 1, 0, "", []talkinters.TalkStatus{talkinters.TalkStatusOpened})
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(talks))

	talks, err = m.QueryTalks(ctx, 1, 0, "", []talkinters.TalkStatus{talkinters.TalkStatusClosed})
	assert.Nil(t, err)
	assert.EqualValues(t, 0, len(talks))
}
