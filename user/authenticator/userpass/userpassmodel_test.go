package userpass

import (
	"context"
	"testing"

	"github.com/sbasestarter/bizmongolib/mongolib"
	"github.com/stretchr/testify/assert"
)

func TestMongoUserPasswordManagerImpl_AddUser(t *testing.T) {
	cli, _, err := mongolib.InitMongo("mongodb://mongo_default_user:mongo_default_pass@127.0.0.1:8309/my_db")
	assert.Nil(t, err)
	assert.NotNil(t, cli)

	m := NewMongoUserPasswordModel(cli, "my_db", "users1", nil)
	t.Log(m)

	mi, ok := m.(*mongoUserPasswordModelImpl)
	assert.True(t, ok)

	_ = mi.collection.Drop(context.TODO())
	_ = mi.db.Collection("ids").Drop(context.TODO())

	m2 := NewMongoUserPasswordModel(cli, "my_db", "users1", nil)
	t.Log(m2)

	_, err = m2.AddUser(context.TODO(), "user1", "pass1")
	assert.Nil(t, err)

	_, err = m2.AddUser(context.TODO(), "user1", "pass1")
	assert.NotNil(t, err)

	_, err = m2.AddUser(context.TODO(), "user2", "pass1")
	assert.Nil(t, err)

	_, err = m2.AddUser(context.TODO(), "user3", "pass1")
	assert.Nil(t, err)

	err = m2.DeleteUser(context.TODO(), 4)
	assert.Nil(t, err)

	u, err := m2.GetUser(context.TODO(), 3)
	assert.Nil(t, err)
	assert.NotNil(t, u)
	assert.EqualValues(t, 3, u.ID)
	assert.EqualValues(t, "user2", u.UserName)
	assert.EqualValues(t, "pass1", u.Password)

	us, err := m2.ListUsers(context.TODO())
	assert.Nil(t, err)
	assert.EqualValues(t, 2, len(us))

	err = m2.UpdateUserExData(context.TODO(), u.ID, "k1", []byte("abcd"))
	assert.Nil(t, err)

	err = m2.UpdateUserExData(context.TODO(), u.ID, "k2", []byte("efgh"))
	assert.Nil(t, err)

	err = m2.UpdateUserExData(context.TODO(), u.ID, "k1", []byte("x"))
	assert.Nil(t, err)
}
