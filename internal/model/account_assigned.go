package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionAccountAssigned 对应 MongoDB 集合 accountAssigned。
const CollectionAccountAssigned = "accountAssigned"

// AccountAssigned 对应文档: {"uuid": String, "account": String, "createTime": Date}。
type AccountAssigned struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UUID       string             `bson:"uuid" json:"uuid"`
	Account    string             `bson:"account" json:"account"`
	CreateTime time.Time          `bson:"createTime" json:"createTime"`
}
