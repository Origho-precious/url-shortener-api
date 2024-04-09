package configs

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func ConnectDB() (*mongo.Database, error) {
	cfg, err := LoadEnvs()
	if err != nil {
		return nil, fmt.Errorf("error loading env: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MONGO_URI))
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("internal server error: %w", err)
	}

	fmt.Println("Database connection established!")

	database := client.Database("likr")

	return database, nil
}
