package service

import (
	"context"
	"go-delivery/pb"
	"go-delivery/services/orders/store"
	"go-delivery/services/orders/workflows"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.temporal.io/sdk/client"
	"log"
	"time"
)

type workflowService struct {
	ordersStore    store.OrdersStore
	walletsClient  pb.WalletsServiceClient
	accountsClient pb.AccountsServiceClient
	productsClient pb.ProductsServiceClient
	temporalClient client.Client
	pb.UnimplementedOrdersWorkflowServiceServer
}

func NewOrdersWorkflowService(
	ordersStore store.OrdersStore,
	walletsClient pb.WalletsServiceClient,
	accountsClient pb.AccountsServiceClient,
	productsClient pb.ProductsServiceClient,
	temporalClient client.Client,
) pb.OrdersWorkflowServiceServer {

	return &workflowService{
		ordersStore:    ordersStore,
		walletsClient:  walletsClient,
		accountsClient: accountsClient,
		productsClient: productsClient,
		temporalClient: temporalClient,
	}
}

func (s *workflowService) StartOrder(ctx context.Context, req *pb.StartOrderRequest) (*pb.DefaultOrderResponse, error) {

	product, err := s.productsClient.GetProduct(ctx, &pb.GetProductRequest{
		Id: req.ProductId,
	})
	if err != nil {
		return nil, err
	}

	sellerWallet, err := s.walletsClient.GetUserWallet(ctx, &pb.GetUserWalletRequest{
		UserId: req.CustomerId,
	})

	customerWallet, err := s.walletsClient.GetUserWallet(ctx, &pb.GetUserWalletRequest{
		UserId: req.CustomerId,
	})
	if err != nil {
		return nil, err
	}

	orderId := primitive.NewObjectID()

	workflowInput := workflows.OrderWorkflowState{
		Id:       orderId,
		Order:    req,
		Customer: customerWallet,
		Seller:   sellerWallet,
		Product:  product,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:                 orderId.Hex(),
		TaskQueue:          workflows.TaskQueueName,
		WorkflowRunTimeout: time.Minute * 10,
	}

	we, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.OrderWorkflow, workflowInput)
	if err != nil {
		return nil, err
	}

	log.Printf("OrderId=%v, WorkflowId=%v, RunId=%v\n", workflowInput.Id, we.GetID(), we.GetRunID())

	return &pb.DefaultOrderResponse{OrderId: orderId.Hex()}, nil
}

func (s *workflowService) QueryOrder(ctx context.Context, req *pb.QueryOrderRequest) (*pb.Order, error) {
	res, err := s.temporalClient.QueryWorkflow(ctx, req.OrderId, "", workflows.QueryOrderByIdName)
	if err != nil {
		return nil, err
	}

	var state workflows.OrderWorkflowState

	err = res.Get(&state)
	if err != nil {
		return nil, err
	}

	order := &pb.Order{
		Id:           state.Id.Hex(),
		CustomerId:   state.Order.CustomerId,
		SellerId:     state.Order.SellerId,
		ProductId:    state.Order.ProductId,
		Status:       state.Status,
		UnitPrice:    state.UnitPrice,
		Quantity:     state.Order.Quantity,
		DeliveryCost: state.Product.DeliveryCost,
		Amount:       state.Amount,
		CreatedAt:    state.CreatedAt.Unix(),
		UpdatedAt:    state.UpdatedAt.Unix(),
	}

	if state.Deliverer != nil {
		order.DelivererId = state.Deliverer.UserId
	}

	return order, nil
}

func (s *workflowService) AcceptOrder(ctx context.Context, req *pb.AcceptOrderRequest) (*pb.DefaultOrderResponse, error) {

	log.Println("AcceptOrder: ", req.Id)

	err := s.temporalClient.SignalWorkflow(ctx, req.Id, "", workflows.AcceptOrderSignalName, int32(pb.OrderStatus_Accepted))
	if err != nil {
		return nil, err
	}
	return &pb.DefaultOrderResponse{OrderId: req.Id}, nil
}

func (s *workflowService) StartDelivery(ctx context.Context, req *pb.StartDeliveryRequest) (*pb.DefaultOrderResponse, error) {

	deliverer, err := s.walletsClient.GetUserWallet(ctx, &pb.GetUserWalletRequest{UserId: req.DelivererId})
	if err != nil {
		return nil, err
	}

	log.Println("StartDelivery: ", req.DelivererId)

	err = s.temporalClient.SignalWorkflow(ctx, req.OrderId, "", workflows.StartDeliverySignalName, deliverer)
	if err != nil {
		return nil, err
	}

	return &pb.DefaultOrderResponse{OrderId: req.OrderId}, nil
}
func (s *workflowService) ConfirmDeliveredOrder(ctx context.Context, req *pb.ConfirmDeliveredOrderRequest) (*pb.DefaultOrderResponse, error) {

	log.Println("ConfirmDeliveredOrder: ", req.OrderId)

	err := s.temporalClient.SignalWorkflow(ctx, req.OrderId, "", workflows.DeliveredOrderSignalName, int32(pb.OrderStatus_Delivered))
	if err != nil {
		return nil, err
	}

	return &pb.DefaultOrderResponse{OrderId: req.OrderId}, nil

}
