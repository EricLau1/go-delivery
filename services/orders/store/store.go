package store

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

const OrdersCollection = "orders"

type OrdersStore interface {
	Create(ctx context.Context, order *Order) error
	Update(ctx context.Context, order *Order) error
	Get(ctx context.Context, id primitive.ObjectID) (*Order, error)
	GetAll(ctx context.Context) ([]*Order, error)
	GetBySeller(ctx context.Context, sellerId primitive.ObjectID) ([]*Order, error)
	GetByStatus(ctx context.Context, status int32) ([]*Order, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type store struct {
	conn *mongo.Collection
}

func NewOrdersStore(dbConn *mongo.Database) OrdersStore {
	return &store{conn: dbConn.Collection(OrdersCollection)}
}

func (s *store) Create(ctx context.Context, order *Order) error {
	result, err := s.conn.InsertOne(ctx, order)
	if err != nil {
		return err
	}
	log.Printf("order created: id=%v\n", result.InsertedID)
	return nil
}

func (s *store) Update(ctx context.Context, order *Order) error {
	update := bson.M{
		"$set": bson.M{
			"delivery_id": order.DeliveryId,
			"status":      order.Status,
			"updated_at":  order.UpdatedAt,
		},
	}

	filter := bson.M{"_id": bson.M{"$eq": order.Id}}

	result, err := s.conn.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	log.Printf("order updated: total=%v\n", result.ModifiedCount)
	return nil
}

func (s *store) Get(ctx context.Context, id primitive.ObjectID) (*Order, error) {
	var order Order
	err := s.conn.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		return nil, err
	}
	log.Printf("order found: id=%v\n", order.Id.Hex())
	return &order, nil
}

func (s *store) GetAll(ctx context.Context) ([]*Order, error) {
	cursor, err := s.conn.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*Order

	err = cursor.All(ctx, &orders)
	if err != nil {
		return nil, err
	}

	log.Printf("list orders: total=%v\n", len(orders))
	return orders, nil
}

func (s *store) GetBySeller(ctx context.Context, sellerId primitive.ObjectID) ([]*Order, error) {
	cursor, err := s.conn.Find(ctx, bson.M{"seller_id": sellerId.Hex()})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*Order

	err = cursor.All(ctx, &orders)
	if err != nil {
		return nil, err
	}
	log.Printf("list orders: total=%v\n", len(orders))
	return orders, nil
}

func (s *store) GetByStatus(ctx context.Context, status int32) ([]*Order, error) {
	cursor, err := s.conn.Find(ctx, bson.M{"status": status})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []*Order

	err = cursor.All(ctx, &orders)
	if err != nil {
		return nil, err
	}

	log.Printf("list orders: total=%v\n", len(orders))
	return orders, nil
}

func (s *store) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.conn.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	log.Printf("order deleted: total=%v\n", result.DeletedCount)
	return nil
}
