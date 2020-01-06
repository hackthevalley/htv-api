package database

import (
	"context"
	"github.com/hackthevalley/htv-api/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

var DbClient *mongo.Database

func RetrieveDatabaseConn(fallbackDBUrl string, fallbackDBName string) *mongo.Database {
	dbURL := utils.GetEnv("DB_URL", fallbackDBUrl)
	dbName := utils.GetEnv("DB_NAME", fallbackDBName)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURL))
	if err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("Database could not be pinged: %s", err)
	}
	return client.Database(dbName)
}
