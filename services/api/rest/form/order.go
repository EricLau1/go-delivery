package form

import (
	"go-delivery/pb"
	"time"
)

type OrderInput struct {
	SellerId  string `validate:"required" json:"seller_id"`
	ProductId string `validate:"required" json:"product_id"`
	Quantity  int32  `validate:"required" json:"quantity"`
}

type Order struct {
	Id           string    `json:"id"`
	CustomerId   string    `json:"customer_id"`
	SellerId     string    `json:"seller_id"`
	ProductId    string    `json:"product_id"`
	DelivererId  string    `json:"delivery_id"`
	Status       string    `json:"status"`
	Quantity     int32     `json:"quantity"`
	UnitPrice    float32   `json:"unit_price"`
	DeliveryCost float32   `json:"delivery_cost"`
	Amount       float32   `json:"amount"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func FromOrder(order *pb.Order) *Order {
	return &Order{
		Id:           order.Id,
		CustomerId:   order.CustomerId,
		SellerId:     order.SellerId,
		ProductId:    order.ProductId,
		DelivererId:  order.DelivererId,
		Status:       order.Status.String(),
		Quantity:     order.Quantity,
		UnitPrice:    order.UnitPrice,
		DeliveryCost: order.DeliveryCost,
		Amount:       order.Amount,
		CreatedAt:    time.Unix(order.CreatedAt, 0),
		UpdatedAt:    time.Unix(order.UpdatedAt, 0),
	}
}
