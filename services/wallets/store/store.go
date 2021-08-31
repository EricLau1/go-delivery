package store

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

const WalletsCollection = "wallets"

type WalletsStore interface {
	Create(ctx context.Context, wallet *Wallet) error
	Update(ctx context.Context, wallet *Wallet) error
	Get(ctx context.Context, id primitive.ObjectID) (*Wallet, error)
	GetByUser(ctx context.Context, id primitive.ObjectID) (*Wallet, error)
	GetAll(ctx context.Context) ([]*Wallet, error)
}

type store struct {
	conn *mongo.Collection
}

func NewWalletsStore(dbConn *mongo.Database) WalletsStore {
	return &store{conn: dbConn.Collection(WalletsCollection)}
}

func (s *store) Create(ctx context.Context, wallet *Wallet) error {
	result, err := s.conn.InsertOne(ctx, wallet)
	if err != nil {
		return err
	}

	log.Printf("wallet created: id=%v", result.InsertedID)

	return nil
}

func (s *store) Update(ctx context.Context, wallet *Wallet) error {
	update := bson.M{
		"$set": bson.M{
			"cash":       wallet.Cash,
			"updated_at": wallet.UpdatedAt,
		},
	}

	filter := bson.M{"_id": bson.M{"$eq": wallet.Id}}

	result, err := s.conn.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	log.Printf("wallet updated: total=%v", result.ModifiedCount)

	return nil
}

func (s *store) Get(ctx context.Context, id primitive.ObjectID) (*Wallet, error) {
	wallet := new(Wallet)

	err := s.conn.FindOne(ctx, bson.M{"_id": id}).Decode(wallet)
	if err != nil {
		return nil, err
	}

	log.Printf("wallet found: id=%v", wallet.Id.Hex())

	return wallet, nil
}

func (s *store) GetByUser(ctx context.Context, id primitive.ObjectID) (*Wallet, error) {
	wallet := new(Wallet)

	err := s.conn.FindOne(ctx, bson.M{"user_id": id.Hex()}).Decode(wallet)
	if err != nil {
		return nil, err
	}

	log.Printf("wallet found: id=%s", wallet.Id.Hex())

	return wallet, nil
}

func (s *store) GetAll(ctx context.Context) ([]*Wallet, error) {
	cursor, err := s.conn.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var wallets []*Wallet

	err = cursor.All(ctx, &wallets)
	if err != nil {
		return nil, err
	}

	log.Printf("list wallets: total=%d", len(wallets))

	return wallets, nil
}
