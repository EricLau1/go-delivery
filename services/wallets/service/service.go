package service

import (
	"context"
	"fmt"
	"go-delivery/pb"
	"go-delivery/services/wallets/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type serviceImpl struct {
	walletsStore store.WalletsStore
	pb.UnimplementedWalletsServiceServer
}

func NewService(walletsStore store.WalletsStore) pb.WalletsServiceServer {
	return &serviceImpl{walletsStore: walletsStore}
}

func (s *serviceImpl) CreateWallet(ctx context.Context, req *pb.Wallet) (*pb.Wallet, error) {
	id, err := primitive.ObjectIDFromHex(req.UserId)
	if err != nil {
		return nil, err
	}

	wallet, err := s.walletsStore.GetByUser(ctx, id)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	if wallet != nil {
		return wallet.ToProto(), nil
	}

	dbWallet, err := store.FromProto(req)
	if err != nil {
		return nil, err
	}

	err = s.walletsStore.Create(ctx, dbWallet)
	if err != nil {
		return nil, err
	}

	return dbWallet.ToProto(), nil
}

func (s *serviceImpl) GetUserWallet(ctx context.Context, req *pb.GetUserWalletRequest) (*pb.Wallet, error) {
	id, err := primitive.ObjectIDFromHex(req.UserId)
	if err != nil {
		return nil, err
	}

	wallet, err := s.walletsStore.GetByUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return wallet.ToProto(), nil
}

func (s *serviceImpl) GetWallet(ctx context.Context, req *pb.GetWalletRequest) (*pb.Wallet, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	wallet, err := s.walletsStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return wallet.ToProto(), nil
}

func (s *serviceImpl) Credit(ctx context.Context, req *pb.CreditRequest) (*pb.Wallet, error) {
	if req.Amount < 0 {
		return nil, fmt.Errorf("invalid amount for credit: %f", req.Amount)
	}

	id, err := primitive.ObjectIDFromHex(req.WalletId)
	if err != nil {
		return nil, err
	}

	wallet, err := s.walletsStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	wallet.Cash += req.Amount
	wallet.UpdatedAt = time.Now()

	err = s.walletsStore.Update(ctx, wallet)
	if err != nil {
		return nil, err
	}

	return wallet.ToProto(), nil
}
func (s *serviceImpl) Debit(ctx context.Context, req *pb.DebitRequest) (*pb.Wallet, error) {
	if req.Amount < 0 {
		return nil, fmt.Errorf("invalid amount for debit: %f", req.Amount)
	}

	id, err := primitive.ObjectIDFromHex(req.WalletId)
	if err != nil {
		return nil, err
	}

	wallet, err := s.walletsStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	wallet.Cash -= req.Amount
	wallet.UpdatedAt = time.Now()

	err = s.walletsStore.Update(ctx, wallet)
	if err != nil {
		return nil, err
	}

	return wallet.ToProto(), nil
}
func (s *serviceImpl) ListWallets(_ *pb.ListWalletsRequest, stream pb.WalletsService_ListWalletsServer) error {
	wallets, err := s.walletsStore.GetAll(context.Background())
	if err != nil {
		return err
	}

	for index := range wallets {
		err = stream.Send(wallets[index].ToProto())
		if err != nil {
			return err
		}
	}

	return nil
}
