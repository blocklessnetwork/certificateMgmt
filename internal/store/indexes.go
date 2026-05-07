package store

import (
	"context"

	"certificatemgmt/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes 创建分配逻辑依赖的索引（幂等）。
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	assigned := db.Collection(model.CollectionAccountAssigned)
	_, err := assigned.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "uuid", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("accountAssigned_uuid_unique"),
		},
		{
			Keys:    bson.D{{Key: "account", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("accountAssigned_account_unique"),
		},
	})
	if err != nil {
		return err
	}

	accounts := db.Collection(model.CollectionAccounts)
	_, err = accounts.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "account", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("accounts_account_unique"),
	})
	return err
}
