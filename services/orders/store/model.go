package store

import (
	"go-delivery/pb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Order struct {
	Id           primitive.ObjectID `bson:"_id"`
	CustomerId   string             `bson:"customer_id"`
	SellerId     string             `bson:"seller_id"`
	ProductId    string             `bson:"product_id"`
	DeliveryId   string             `bson:"delivery_id"`
	Status       int32              `bson:"status"`
	Quantity     int32              `bson:"quantity"`
	UnitPrice    float32            `bson:"unit_price"`
	DeliveryCost float32            `bson:"delivery_cost"`
	Amount       float32            `bson:"amount"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
}

func (o *Order) ToProto() *pb.Order {
	return &pb.Order{
		Id:           o.Id.Hex(),
		CustomerId:   o.CustomerId,
		SellerId:     o.SellerId,
		ProductId:    o.ProductId,
		DelivererId:  o.DeliveryId,
		Status:       pb.OrderStatus(o.Status),
		Quantity:     o.Quantity,
		UnitPrice:    o.UnitPrice,
		DeliveryCost: o.DeliveryCost,
		Amount:       o.Amount,
		CreatedAt:    o.CreatedAt.Unix(),
		UpdatedAt:    o.UpdatedAt.Unix(),
	}
}

func (o *Order) CanCancel() bool {
	return pb.OrderStatus(o.Status) != pb.OrderStatus_Delivered
}

func FromProto(o *pb.Order) (*Order, error) {
	id, err := primitive.ObjectIDFromHex(o.Id)
	if err != nil {
		return nil, err
	}
	customerId, err := primitive.ObjectIDFromHex(o.CustomerId)
	if err != nil {
		return nil, err
	}
	sellerId, err := primitive.ObjectIDFromHex(o.SellerId)
	if err != nil {
		return nil, err
	}
	productId, err := primitive.ObjectIDFromHex(o.ProductId)
	if err != nil {
		return nil, err
	}

	var deliveryId string
	if o.Status == pb.OrderStatus_Delivering {
		objectId, err := primitive.ObjectIDFromHex(o.DelivererId)
		if err != nil {
			return nil, err
		}
		deliveryId = objectId.Hex()
	}

	return &Order{
		Id:           id,
		CustomerId:   customerId.Hex(),
		SellerId:     sellerId.Hex(),
		ProductId:    productId.Hex(),
		DeliveryId:   deliveryId,
		Status:       int32(o.Status),
		Quantity:     o.Quantity,
		UnitPrice:    o.UnitPrice,
		DeliveryCost: o.DeliveryCost,
		Amount:       o.Amount,
		CreatedAt:    time.Unix(o.CreatedAt, 0),
		UpdatedAt:    time.Unix(o.UpdatedAt, 0),
	}, nil
}
