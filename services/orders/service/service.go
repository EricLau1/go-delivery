package service

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"go-delivery/pb"
	"go-delivery/services/orders/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"time"
)

type service struct {
	ordersStore    store.OrdersStore
	walletsClient  pb.WalletsServiceClient
	accountsClient pb.AccountsServiceClient
	productsClient pb.ProductsServiceClient
	pb.UnimplementedOrdersServiceServer
}

func NewService(
	ordersStore store.OrdersStore,
	walletsClient pb.WalletsServiceClient,
	accountsClient pb.AccountsServiceClient,
	productsClient pb.ProductsServiceClient,
) pb.OrdersServiceServer {

	return &service{
		ordersStore:    ordersStore,
		walletsClient:  walletsClient,
		accountsClient: accountsClient,
		productsClient: productsClient,
	}
}

func (s *service) CreateOrder(ctx context.Context, req *pb.Order) (*pb.Order, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	stream, err := s.productsClient.ListSellerProducts(ctx, &pb.ListSellerProductsRequest{SellerId: req.SellerId})
	if err != nil {
		return nil, err
	}

	var product *pb.Product

	for {
		received, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if received.Id == req.ProductId {
			product = received
			break
		}
	}

	if product == nil {
		return nil, fmt.Errorf("invalid order, product not found: productId=%v, sellerId=%s", req.ProductId, req.SellerId)
	}

	if product.Quantity < req.Quantity {
		return nil, fmt.Errorf("invalid order, products insufficient: productId=%v", req.ProductId)
	}

	amount := (product.Price * float32(req.Quantity)) + product.DeliveryCost

	wallet, err := s.walletsClient.GetUserWallet(ctx, &pb.GetUserWalletRequest{UserId: req.CustomerId})
	if err != nil {
		return nil, err
	}

	if wallet.Cash < amount {
		return nil, fmt.Errorf("invalid order, amount insufficient: customerId=%v", req.CustomerId)
	}

	order := &store.Order{
		Id:           id,
		CustomerId:   req.CustomerId,
		SellerId:     req.SellerId,
		ProductId:    req.ProductId,
		Status:       int32(pb.OrderStatus_Placed),
		Quantity:     req.Quantity,
		UnitPrice:    product.Price,
		DeliveryCost: product.DeliveryCost,
		Amount:       amount,
		CreatedAt:    time.Unix(req.CreatedAt, 0),
		UpdatedAt:    time.Unix(req.UpdatedAt, 0),
	}

	err = s.ordersStore.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	product.Quantity -= req.Quantity

	_, err = s.productsClient.UpdateProduct(ctx, &pb.UpdateProductRequest{
		Id:           product.Id,
		Name:         product.Name,
		Price:        product.Price,
		DeliveryCost: product.DeliveryCost,
		Quantity:     product.Quantity,
	})
	if err != nil {
		return nil, err
	}

	debit := &pb.DebitRequest{
		WalletId: wallet.Id,
		Amount:   amount,
	}

	_, err = s.walletsClient.Debit(ctx, debit)
	if err != nil {
		return nil, err
	}

	return order.ToProto(), nil
}

func (s *service) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.Order, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	order, err := s.ordersStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return order.ToProto(), nil
}

func (s *service) ListOrders(_ *pb.ListOrdersRequest, stream pb.OrdersService_ListOrdersServer) error {
	orders, err := s.ordersStore.GetAll(context.Background())
	if err != nil {
		return err
	}

	for index := range orders {
		err = stream.Send(orders[index].ToProto())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *service) ListOrdersBySeller(req *pb.ListOrdersBySellerRequest, stream pb.OrdersService_ListOrdersBySellerServer) error {
	id, err := primitive.ObjectIDFromHex(req.SellerId)
	if err != nil {
		return err
	}

	orders, err := s.ordersStore.GetBySeller(context.Background(), id)
	if err != nil {
		return err
	}

	for index := range orders {
		err = stream.Send(orders[index].ToProto())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *service) ListOrdersByStatus(req *pb.ListOrdersByStatusRequest, stream pb.OrdersService_ListOrdersByStatusServer) error {
	orders, err := s.ordersStore.GetByStatus(context.Background(), int32(req.Status))
	if err != nil {
		return err
	}

	for index := range orders {
		err = stream.Send(orders[index].ToProto())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *service) ApproveOrder(ctx context.Context, req *pb.ApproveOrderRequest) (*pb.Order, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	order, err := s.ordersStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if order.Status != int32(pb.OrderStatus_Placed) {
		return nil, fmt.Errorf("can't change order status to accepted: orderId=%v", order.Id.Hex())
	}

	seller, err := s.accountsClient.GetUser(ctx, &pb.GetUserRequest{Id: req.SellerId})
	if err != nil {
		return nil, err
	}

	if seller.Role != pb.Role_Seller {
		return nil, fmt.Errorf("can't change order status to accepted, role is not allowed: sellerId=%v", seller.Id)
	}

	order.Status = int32(pb.OrderStatus_Accepted)
	order.UpdatedAt = time.Now()

	err = s.ordersStore.Update(ctx, order)
	if err != nil {
		return nil, err
	}

	return order.ToProto(), nil
}

func (s *service) DeliverOrder(ctx context.Context, req *pb.DeliverOrderRequest) (*pb.Order, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	order, err := s.ordersStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if order.Status != int32(pb.OrderStatus_Accepted) {
		return nil, fmt.Errorf("can't change order status to delivering: orderId=%v", order.Id.Hex())
	}

	deliverer, err := s.accountsClient.GetUser(ctx, &pb.GetUserRequest{Id: req.DeliveryId})
	if err != nil {
		return nil, err
	}

	if deliverer.Role != pb.Role_Delivery {
		return nil, fmt.Errorf("can't change order status to delivering, role is not allowed: delivererId=%v", deliverer.Id)
	}

	order.DeliveryId = req.DeliveryId
	order.Status = int32(pb.OrderStatus_Delivering)
	order.UpdatedAt = time.Now()

	err = s.ordersStore.Update(ctx, order)
	if err != nil {
		return nil, err
	}

	return order.ToProto(), nil
}

func (s *service) ConfirmOrderDelivered(ctx context.Context, req *pb.ConfirmOrderDeliveredRequest) (*pb.Order, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	order, err := s.ordersStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if order.Status != int32(pb.OrderStatus_Delivering) {
		return nil, fmt.Errorf("can't change order status to delivered: orderId=%v", order.Id.Hex())
	}

	customer, err := s.accountsClient.GetUser(ctx, &pb.GetUserRequest{Id: req.CustomerId})
	if err != nil {
		return nil, err
	}

	if customer.Role != pb.Role_Customer {
		return nil, fmt.Errorf("can't change order status to delivered, role is not allowed: customerId=%v", customer.Id)
	}

	walletSeller, err := s.walletsClient.GetUserWallet(ctx, &pb.GetUserWalletRequest{UserId: order.SellerId})
	if err != nil {
		return nil, err
	}

	sellerAmount := order.UnitPrice * float32(order.Quantity)

	_, err = s.walletsClient.Credit(ctx, &pb.CreditRequest{
		WalletId: walletSeller.Id,
		Amount:   sellerAmount,
	})
	if err != nil {
		return nil, err
	}

	walletDeliverer, err := s.walletsClient.GetUserWallet(ctx, &pb.GetUserWalletRequest{UserId: order.DeliveryId})
	if err != nil {
		return nil, err
	}

	delivererAmount := order.Amount - sellerAmount

	_, err = s.walletsClient.Credit(ctx, &pb.CreditRequest{
		WalletId: walletDeliverer.Id,
		Amount:   delivererAmount,
	})
	if err != nil {
		return nil, err
	}

	order.Status = int32(pb.OrderStatus_Delivered)
	order.UpdatedAt = time.Now()

	err = s.ordersStore.Update(ctx, order)
	if err != nil {
		return nil, err
	}

	return order.ToProto(), nil
}

func (s *service) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*empty.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	order, err := s.ordersStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if !order.CanCancel() {
		return nil, fmt.Errorf("order can't be canceled: orderId=%v", order.Id)
	}

	customerWallet, err := s.walletsClient.GetUserWallet(ctx, &pb.GetUserWalletRequest{UserId: order.CustomerId})
	if err != nil {
		return nil, err
	}

	product, err := s.productsClient.GetProduct(ctx, &pb.GetProductRequest{Id: order.ProductId})
	if err != nil {
		return nil, err
	}

	credit := &pb.CreditRequest{
		WalletId: customerWallet.Id,
		Amount:   order.Amount,
	}

	_, err = s.walletsClient.Credit(ctx, credit)
	if err != nil {
		return nil, err
	}

	_, err = s.productsClient.UpdateProduct(ctx, &pb.UpdateProductRequest{
		Id:           product.Id,
		Name:         product.Name,
		Price:        product.Price,
		DeliveryCost: product.DeliveryCost,
		Quantity:     product.Quantity + order.Quantity,
	})
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}
func (s *service) DeleteOrder(ctx context.Context, req *pb.DeleteOrderRequest) (*empty.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		return nil, err
	}

	order, err := s.ordersStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if order.CanCancel() {
		_, err = s.CancelOrder(ctx, &pb.CancelOrderRequest{Id: id.Hex()})
		if err != nil {
			return nil, err
		}
	}

	err = s.ordersStore.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}
