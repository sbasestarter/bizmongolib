package userpass

type User struct {
	ID       uint64 `bson:"_id"`
	UserName string `bson:"user_name"`
	Password string `bson:"password"`
	CreateAt int64  `bson:"create_at"`
}
