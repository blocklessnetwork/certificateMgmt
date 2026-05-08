package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionAccountAssigned is the MongoDB collection name for assignments.
const CollectionAccountAssigned = "accountAssigned"

// AccountAssigned binds a uuid to an account with a createTime (BSON date).
type AccountAssigned struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UUID       string             `bson:"uuid" json:"uuid"`
	Account    string             `bson:"account" json:"account"`
	CreateTime time.Time          `bson:"createTime" json:"createTime"`
}
