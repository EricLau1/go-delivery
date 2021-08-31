package store

import (
	"go-delivery/pb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	Id        primitive.ObjectID `bson:"_id"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	Role      int32              `bson:"role"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func (u *User) ToProto() *pb.User {
	return &pb.User{
		Id:        u.Id.Hex(),
		Email:     u.Email,
		Password:  u.Password,
		Role:      pb.Role(u.Role),
		CreatedAt: u.CreatedAt.Unix(),
		UpdatedAt: u.UpdatedAt.Unix(),
	}
}

func FromProto(user *pb.User) (*User, error) {
	var dbUser User
	var err error

	dbUser.Id, err = primitive.ObjectIDFromHex(user.Id)
	if err != nil {
		return nil, err
	}
	dbUser.Email = user.Email
	dbUser.Password = user.Password
	dbUser.Role = int32(user.Role)
	dbUser.CreatedAt = time.Unix(user.CreatedAt, 0)
	dbUser.UpdatedAt = time.Unix(user.UpdatedAt, 0)

	return &dbUser, nil
}
