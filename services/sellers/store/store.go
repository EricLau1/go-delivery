package store

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

const ProductsCollection = "products"

type ProductsStore interface {
	Create(ctx context.Context, product *Product) error
	Update(ctx context.Context, product *Product) error
	Get(ctx context.Context, id primitive.ObjectID) (*Product, error)
	GetByName(ctx context.Context, name string) (*Product, error)
	GetBySeller(ctx context.Context, id primitive.ObjectID) ([]*Product, error)
	GetAll(ctx context.Context) ([]*Product, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type store struct {
	conn *mongo.Collection
}

func NewProductsStore(dbConn *mongo.Database) ProductsStore {
	return &store{conn: dbConn.Collection(ProductsCollection)}
}

func (s *store) Create(ctx context.Context, product *Product) error {
	result, err := s.conn.InsertOne(ctx, product)
	if err != nil {
		return err
	}
	log.Printf("product added: id=%v\n", result.InsertedID)
	return nil
}

func (s *store) Update(ctx context.Context, product *Product) error {

	update := bson.M{
		"$set": bson.M{
			"name":          product.Name,
			"price":         product.Price,
			"delivery_cost": product.DeliveryCost,
			"quantity":      product.Quantity,
			"updated_at":    product.UpdatedAt,
		},
	}

	filter := bson.M{"_id": bson.M{"$eq": product.Id}}

	result, err := s.conn.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	log.Printf("product added: total=%v\n", result.ModifiedCount)

	return nil
}

func (s *store) Get(ctx context.Context, id primitive.ObjectID) (*Product, error) {
	var product Product
	err := s.conn.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err != nil {
		return nil, err
	}
	fmt.Printf("product found: id=%v\n", product.Id.Hex())
	return &product, nil
}

func (s *store) GetByName(ctx context.Context, name string) (*Product, error) {
	var product Product
	err := s.conn.FindOne(ctx, bson.M{"name": name}).Decode(&product)
	if err != nil {
		return nil, err
	}
	fmt.Printf("product found: name=%v\n", product.Name)
	return &product, nil
}

func (s *store) GetBySeller(ctx context.Context, id primitive.ObjectID) ([]*Product, error) {
	cursor, err := s.conn.Find(ctx, bson.M{"seller_id": id.Hex()})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*Product

	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, err
	}

	fmt.Printf("list products: total=%v\n", len(products))

	return products, nil
}

func (s *store) GetAll(ctx context.Context) ([]*Product, error) {
	cursor, err := s.conn.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*Product

	err = cursor.All(ctx, &products)
	if err != nil {
		return nil, err
	}

	fmt.Printf("list products: total=%v\n", len(products))

	return products, nil
}

func (s *store) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.conn.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	fmt.Printf("product deleted: total=%d\n", result.DeletedCount)
	return nil
}
