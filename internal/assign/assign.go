package assign

import (
	"context"
	"errors"
	"fmt"
	"time"

	"certificatemgmt/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrNoAccountAvailable 没有仍可分配的账号。
var ErrNoAccountAvailable = errors.New("no unassigned account available")

// ErrUIDRequired 请求未提供 uid。
var ErrUIDRequired = errors.New("uid is required")

// AccountByUID 若 uid 已分配则返回对应账号；否则从未出现在 accountAssigned 的账号中选一条（按 account 升序）、写入分配记录并返回。
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

		assignedRaw, err := as.Distinct(ctx, "account", bson.M{})
		if err != nil {
			return nil, err
		}
		assigned := make([]string, 0, len(assignedRaw))
		for _, v := range assignedRaw {
			s, ok := v.(string)
			if !ok {
				continue
			}
			assigned = append(assigned, s)
		}

		filter := bson.M{}
		if len(assigned) > 0 {
			filter["account"] = bson.M{"$nin": assigned}
		}

		var cand model.Account
		err = ac.FindOne(ctx, filter, options.FindOne().SetSort(bson.D{{Key: "account", Value: 1}})).Decode(&cand)
		if err == mongo.ErrNoDocuments {
			return nil, ErrNoAccountAvailable
		}
		if err != nil {
			return nil, err
		}

		doc := model.AccountAssigned{
			UUID:       uid,
			Account:    cand.Account,
			CreateTime: time.Now().UTC(),
		}
		_, err = as.InsertOne(ctx, doc)
		if err == nil {
			return &cand, nil
		}
		if mongo.IsDuplicateKeyError(err) {
			if err := as.FindOne(ctx, bson.M{"uuid": uid}).Decode(&existing); err == nil {
				return loadAccount(ctx, ac, existing.Account)
			} else if err != mongo.ErrNoDocuments {
				return nil, err
			}
			continue
		}
		return nil, err
	}

	return nil, fmt.Errorf("assign account: exceeded retries")
}

func loadAccount(ctx context.Context, ac *mongo.Collection, name string) (*model.Account, error) {
	var acc model.Account
	if err := ac.FindOne(ctx, bson.M{"account": name}).Decode(&acc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("accounts: missing row for assigned account %q", name)
		}
		return nil, err
	}
	return &acc, nil
}
