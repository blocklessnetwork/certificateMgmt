package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(ctx context.Context, uri string, timeout time.Duration) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(uri).SetServerSelectionTimeout(timeout)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return mongo.Connect(ctx, opts)
}

func Ping(ctx context.Context, client *mongo.Client, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return client.Ping(ctx, nil)
}
