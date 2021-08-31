package form

import (
	"go-delivery/pb"
	"strings"
	"time"
)

type ProductInput struct {
	Name         string  `validate:"required" json:"name"`
	Price        float32 `validate:"required" json:"price"`
	DeliveryCost float32 `validate:"required" json:"delivery_cost"`
	Quantity     int32   `validate:"required" json:"quantity"`
}

func (i *ProductInput) Clear() {
	i.Name = strings.TrimSpace(i.Name)
}

type Product struct {
	Id           string    `json:"id"`
	SellerId     string    `json:"seller_id"`
	Name         string    `json:"name"`
	Price        float32   `json:"price"`
	DeliveryCost float32   `json:"delivery_cost"`
	Quantity     int32     `json:"quantity"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func FromProduct(p *pb.Product) *Product {
	return &Product{
		Id:           p.Id,
		SellerId:     p.SellerId,
		Name:         p.Name,
		Price:        p.Price,
		DeliveryCost: p.DeliveryCost,
		Quantity:     p.Quantity,
		CreatedAt:    time.Unix(p.CreatedAt, 0),
		UpdatedAt:    time.Unix(p.UpdatedAt, 0),
	}
}
