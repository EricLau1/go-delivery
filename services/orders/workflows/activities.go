package workflows

import (
	"context"
	"fmt"
	"go-delivery/pb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.temporal.io/sdk/activity"
	"time"
)

type Activities struct {
}

type OrderWorkflowState struct {
	Id           primitive.ObjectID
	Order        *pb.StartOrderRequest
	Customer     *pb.Wallet
	Seller       *pb.Wallet
	Deliverer    *pb.Wallet
	Product      *pb.Product
	Amount       float32
	UnitPrice    float32
	Status       pb.OrderStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (c *Activities) NewOrder(ctx context.Context, state OrderWorkflowState) (OrderWorkflowState, error) {

	logger := activity.GetLogger(ctx)
	logger.Info("starting activity: processing new order")

	product := state.Product
	customer := state.Customer

	if product == nil {
		return OrderWorkflowState{}, fmt.Errorf("invalid order, product not found: productId=%v, sellerId=%s", product.Id, state.Order.SellerId)
	}
	if product.Quantity < state.Order.Quantity {
		return OrderWorkflowState{}, fmt.Errorf("invalid order, products insufficient: productId=%v", product.Id)
	}

	amount := (product.Price * float32(state.Order.Quantity)) + product.DeliveryCost

	if customer.Cash < amount {
		return OrderWorkflowState{}, fmt.Errorf("invalid order, amount insufficient: customerId=%v", state.Order.CustomerId)
	}

	state.Status = pb.OrderStatus_Placed
	state.UnitPrice = product.Price
	state.Amount = amount
	state.Product.Quantity -= state.Order.Quantity
	state.Customer.Cash -= state.Amount
	state.CreatedAt = time.Now()
	state.UpdatedAt = time.Now()

	logger.Info("new order activity is finished!")

	return state, nil
}
