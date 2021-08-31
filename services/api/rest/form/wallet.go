package form

import (
	"go-delivery/pb"
	"time"
)

type CreditInput struct {
	Amount float32 `validate:"required,gte=0" json:"amount"`
}

type Wallet struct {
	Id        string    `json:"id"`
	Cash      float32   `json:"cash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func FromWallet(w *pb.Wallet) *Wallet {
	return &Wallet{
		Id:        w.Id,
		Cash:      w.Cash,
		CreatedAt: time.Unix(w.CreatedAt, 0),
		UpdatedAt: time.Unix(w.UpdatedAt, 0),
	}
}
