package service

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"go-delivery/pb"
	"go-delivery/services/sellers/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type service struct {
	productsStore store.ProductsStore
	pb.UnimplementedProductsServiceServer
}

func NewService(productsStore store.ProductsStore) pb.ProductsServiceServer {
	return &service{productsStore: productsStore}
}

func (s *service) CreateProduct(ctx context.Context, req *pb.Product) (*pb.Product, error) {
	sellerId, err := primitive.ObjectIDFromHex(req.SellerId)
	if err != nil {
		return nil, err
	}

	items, err := s.productsStore.GetBySeller(ctx, sellerId)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.Name == req.Name {
			return nil, fmt.Errorf("duplicated product: name=%s, seller_id=%s", req.Name, req.SellerId)
		}
	}

	product, err := store.FromProto(req)

	err = s.productsStore.Create(ctx, product)
	if err != nil {
		return nil, err
	}

	return product.ToProto(), nil
}

func (s *service) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	product, err := s.productsStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	product.Name = req.Name
	product.Price = req.Price
	product.Quantity = req.Quantity
	product.UpdatedAt = time.Now()

	err = s.productsStore.Update(ctx, product)
	if err != nil {
		return nil, err
	}

	return product.ToProto(), nil
}

func (s *service) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	product, err := s.productsStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return product.ToProto(), nil
}

func (s *service) ListSellerProducts(req *pb.ListSellerProductsRequest, stream pb.ProductsService_ListSellerProductsServer) error {
	sellerId, err := primitive.ObjectIDFromHex(req.SellerId)
	if err != nil {
		return err
	}

	items, err := s.productsStore.GetBySeller(context.Background(), sellerId)
	if err != nil {
		return err
	}

	for index := range items {
		err = stream.Send(items[index].ToProto())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *service) ListProducts(_ *pb.ListProductsRequest, stream pb.ProductsService_ListProductsServer) error {
	items, err := s.productsStore.GetAll(context.Background())
	if err != nil {
		return err
	}

	for index := range items {
		err = stream.Send(items[index].ToProto())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *service) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*empty.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	err = s.productsStore.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}
