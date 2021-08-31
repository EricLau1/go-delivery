package store

import (
	"go-delivery/pb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Product struct {
	Id           primitive.ObjectID `bson:"_id"`
	SellerId     string             `bson:"seller_id"`
	Name         string             `bson:"name"`
	Price        float32            `bson:"price"`
	DeliveryCost float32            `bson:"delivery_cost"`
	Quantity     int32              `bson:"quantity"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
}

func (p *Product) ToProto() *pb.Product {
	return &pb.Product{
		Id:           p.Id.Hex(),
		SellerId:     p.SellerId,
		Name:         p.Name,
		Price:        p.Price,
		DeliveryCost: p.DeliveryCost,
		Quantity:     p.Quantity,
		CreatedAt:    p.CreatedAt.Unix(),
		UpdatedAt:    p.UpdatedAt.Unix(),
	}
}

func FromProto(p *pb.Product) (*Product, error) {
	id, err := primitive.ObjectIDFromHex(p.Id)
	if err != nil {
		return nil, err
	}
	sellerId, err := primitive.ObjectIDFromHex(p.SellerId)
	if err != nil {
		return nil, err
	}

	return &Product{
		Id:           id,
		SellerId:     sellerId.Hex(),
		Name:         p.Name,
		Price:        p.Price,
		DeliveryCost: p.DeliveryCost,
		Quantity:     p.Quantity,
		CreatedAt:    time.Unix(p.CreatedAt, 0),
		UpdatedAt:    time.Unix(p.UpdatedAt, 0),
	}, nil
}
