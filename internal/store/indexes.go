package store

import (
	"context"

	"certificatemgmt/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes creates indexes required by the assignment logic (idempotent).
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	assigned := db.Collection(model.CollectionAccountAssigned)

	// Older releases enforced unique account per row; drop so multiple uids can share an account.
	_, _ = assigned.Indexes().DropOne(ctx, "accountAssigned_account_unique")

	// uuid must stay unique; account is not unique — several uids may share one account when the pool is exhausted.
	_, err := assigned.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "uuid", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("accountAssigned_uuid_unique"),
		},
		{
			Keys: bson.D{{Key: "account", Value: 1}},
			// Non-unique index speeds up aggregation and queries by account.
			Options: options.Index().SetName("accountAssigned_account_idx"),
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
