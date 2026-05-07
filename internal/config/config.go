package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr        string
	MongoURI        string
	MongoDBName     string
	MongoTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":" + port
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = os.Getenv("MONGO_URI")
	}
	if mongoURI == "" {
		return Config{}, fmt.Errorf("MONGODB_URI (or MONGO_URI) is required")
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "app"
	}

	mongoTimeout := 10 * time.Second
	if v := os.Getenv("MONGODB_TIMEOUT_SEC"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("MONGODB_TIMEOUT_SEC: %w", err)
		}
		mongoTimeout = time.Duration(n) * time.Second
	}

	shutdown := 15 * time.Second
	if v := os.Getenv("SHUTDOWN_TIMEOUT_SEC"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT_SEC: %w", err)
		}
		shutdown = time.Duration(n) * time.Second
	}

	return Config{
		HTTPAddr:        addr,
		MongoURI:        mongoURI,
		MongoDBName:     dbName,
		MongoTimeout:    mongoTimeout,
		ShutdownTimeout: shutdown,
	}, nil
}
