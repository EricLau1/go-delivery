package service

import (
	"context"
	"errors"
	"go-delivery/pb"
	"go-delivery/security/passwords"
	"go-delivery/security/tokens"
	"go-delivery/services/accounts/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type service struct {
	usersStore store.UsersStore
	pb.UnimplementedAccountsServiceServer
}

func NewService(usersStore store.UsersStore) pb.AccountsServiceServer {
	return &service{usersStore: usersStore}
}

func (s *service) SignUp(ctx context.Context, req *pb.User) (*pb.User, error) {

	user, err := s.usersStore.GetByEmail(ctx, req.Email)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	if user != nil {
		return nil, errors.New("email already registered")
	}

	user, err = store.FromProto(req)
	if err != nil {
		return nil, err
	}

	err = s.usersStore.Create(ctx, user)

	return user.ToProto(), nil
}

func (s *service) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {

	user, err := s.usersStore.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	err = passwords.OK(user.Password, req.Password)
	if err != nil {
		return nil, err
	}

	token, err := tokens.New(user.ToProto())
	if err != nil {
		return nil, err
	}

	return &pb.SignInResponse{Token: token}, nil
}

func (s *service) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {

	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	user, err := s.usersStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return user.ToProto(), nil
}

func (s *service) ListUsers(_ *pb.ListUsersRequest, stream pb.AccountsService_ListUsersServer) error {

	users, err := s.usersStore.GetAll(context.Background())
	if err != nil {
		return err
	}

	for index := range users {
		err := stream.Send(users[index].ToProto())
		if err != nil {
			return err
		}
	}

	return nil
}
