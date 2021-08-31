package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
)

type Connection interface {
	Close(ctx context.Context)
	Ping(ctx context.Context) error
	DB() *mongo.Database
}

type dbConn struct {
	client *mongo.Client
}

func New(ctx context.Context, cfg Config) Connection {
	uri :=  options.Client().ApplyURI(cfg.URI())

	client, err := mongo.Connect(ctx, uri)
	if err != nil {
		log.Panicln(err)
	}

	return &dbConn{client: client}
}

func (c *dbConn) Ping(ctx context.Context) error {
	// Ping the primary
	if err := c.client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	return nil
}

func (c *dbConn) DB() *mongo.Database {
	return c.client.Database(os.Getenv("DB_NAME"))
}

func (c *dbConn) Close(ctx context.Context) {
	if err := c.client.Disconnect(ctx); err != nil {
		log.Panicln(err)
	}
}