package form

import (
	"go-delivery/pb"
	"strings"
	"time"
)

type UserForm struct {
	Email    string `validate:"email,required" json:"email"`
	Password string `validate:"required,gte=3,lte=100" json:"password"`
}

func (f *UserForm) Clear() {
	f.Email = strings.ToLower(strings.TrimSpace(f.Email))
}

type SignUpInput struct {
	UserForm
	Role int32 `validate:"gte=0,lte=3" json:"role"`
}

type SignInInput struct {
	UserForm
}

func (i *SignInInput) ToProto() *pb.SignInRequest {
	return &pb.SignInRequest{
		Email:    i.Email,
		Password: i.Password,
	}
}

type User struct {
	Id        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func FromUser(u *pb.User) *User {
	return &User{
		Id:        u.Id,
		Email:     u.Email,
		Role:      u.Role.String(),
		CreatedAt: time.Unix(u.CreatedAt, 0),
		UpdatedAt: time.Unix(u.UpdatedAt, 0),
	}
}
