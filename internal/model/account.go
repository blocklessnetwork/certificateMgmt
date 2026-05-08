package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// CollectionAccounts is the MongoDB collection name for accounts.
const CollectionAccounts = "accounts"

// Account is a document with optional _id plus account, password, and twoFASec.
type Account struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Account  string             `bson:"account" json:"account"`
	Password string             `bson:"password" json:"password"`
	TwoFASec string             `bson:"twoFASec" json:"twoFASec"`
}
