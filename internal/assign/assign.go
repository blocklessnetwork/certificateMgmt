package assign

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"certificatemgmt/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ErrNoAccountAvailable is returned when the accounts collection has no documents.
var ErrNoAccountAvailable = errors.New("no documents in accounts collection")

// ErrNoAssignableAccounts is returned when documents exist but none have a usable account field.
var ErrNoAssignableAccounts = errors.New("accounts documents must include a non-empty account (bson keys: account, Account, or username)")

// ErrUIDRequired is returned when the request omits uid.
var ErrUIDRequired = errors.New("uid is required")

// ErrAssignedAccountMissing means accountAssigned references an account name with no matching document in accounts.
var ErrAssignedAccountMissing = errors.New("assigned account not found in accounts collection")

// AccountByUID returns credentials from accounts for uuid when a row exists in accountAssigned.
// Otherwise it samples a random document from accounts, inserts an accountAssigned row for uuid,
// and returns that account.
func AccountByUID(ctx context.Context, db *mongo.Database, uid string) (*model.Account, error) {
	if uid == "" {
		return nil, ErrUIDRequired
	}

	ac := db.Collection(model.CollectionAccounts)
	as := db.Collection(model.CollectionAccountAssigned)

	var existing model.AccountAssigned
	err := as.FindOne(ctx, bson.M{"uuid": uid}).Decode(&existing)
	if err == nil {
		return loadAccount(ctx, ac, existing.Account)
	}
	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	n, err := ac.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrNoAccountAvailable
	}

	const maxAttempts = 48
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if err := as.FindOne(ctx, bson.M{"uuid": uid}).Decode(&existing); err == nil {
			return loadAccount(ctx, ac, existing.Account)
		} else if err != mongo.ErrNoDocuments {
			return nil, err
		}

		chosen, err := sampleRandomAccount(ctx, ac)
		if err != nil {
			return nil, err
		}

		doc := model.AccountAssigned{
			UUID:       uid,
			Account:    chosen.Account,
			CreateTime: time.Now().UTC(),
		}
		_, err = as.InsertOne(ctx, doc)
		if err == nil {
			return chosen, nil
		}
		if mongo.IsDuplicateKeyError(err) {
			continue
		}
		return nil, err
	}

	return nil, fmt.Errorf("assign account: exceeded retries")
}

func sampleRandomAccount(ctx context.Context, ac *mongo.Collection) (*model.Account, error) {
	const maxTries = 32
	for range maxTries {
		pipeline := mongo.Pipeline{
			bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}},
		}
		cur, err := ac.Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}

		if !cur.Next(ctx) {
			_ = cur.Close(ctx)
			if err := cur.Err(); err != nil {
				return nil, err
			}
			return nil, ErrNoAccountAvailable
		}

		var raw bson.M
		if err := cur.Decode(&raw); err != nil {
			_ = cur.Close(ctx)
			return nil, err
		}
		_ = cur.Close(ctx)

		a, err := model.ParseAccountDocument(raw)
		if err != nil {
			continue
		}
		return &a, nil
	}
	return nil, ErrNoAssignableAccounts
}

func loadAccount(ctx context.Context, ac *mongo.Collection, name string) (*model.Account, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("accounts: empty account key in assignment")
	}
	filter := bson.M{"$or": []bson.M{
		{"account": name},
		{"Account": name},
		{"username": name},
	}}
	var raw bson.M
	if err := ac.FindOne(ctx, filter).Decode(&raw); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w: %q", ErrAssignedAccountMissing, name)
		}
		return nil, err
	}
	a, err := model.ParseAccountDocument(raw)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
