package store

import (
	"go-delivery/pb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Wallet struct {
	Id        primitive.ObjectID `bson:"_id"`
	UserId    string             `bson:"user_id"`
	Cash      float32            `bson:"cash"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func (w *Wallet) ToProto() *pb.Wallet {
	return &pb.Wallet{
		Id:        w.Id.Hex(),
		UserId:    w.UserId,
		Cash:      w.Cash,
		CreatedAt: w.CreatedAt.Unix(),
		UpdatedAt: w.UpdatedAt.Unix(),
	}
}

func FromProto(w *pb.Wallet) (*Wallet, error) {
	var wallet Wallet
	var err error

	wallet.Id, err = primitive.ObjectIDFromHex(w.Id)
	if err != nil {
		return nil, err
	}

	wallet.UserId = w.UserId
	wallet.Cash = w.Cash
	wallet.CreatedAt = time.Unix(w.CreatedAt, 0)
	wallet.UpdatedAt = time.Unix(w.UpdatedAt, 0)

	return &wallet, nil
}