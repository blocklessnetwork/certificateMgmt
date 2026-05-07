package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// CollectionAccounts 对应 MongoDB 集合 accounts。
const CollectionAccounts = "accounts"

// Account 对应文档: {"account": String, "password": String, "twoFASec": String}。
type Account struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Account  string             `bson:"account" json:"account"`
	Password string             `bson:"password" json:"password"`
	TwoFASec string             `bson:"twoFASec" json:"twoFASec"`
}
