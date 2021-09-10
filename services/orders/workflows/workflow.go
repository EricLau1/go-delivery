package workflows

import (
	"fmt"
	"go-delivery/pb"
	"go.temporal.io/sdk/workflow"
	"time"
)

const (
	TaskQueueName            = "OrdersTaskQueue"
	QueryOrderByIdName       = "QueryOrderById"
	AcceptOrderSignalName    = "AcceptOrder"
	StartDeliverySignalName  = "StartDelivery"
	DeliveredOrderSignalName = "DeliveredOrder"
)

func OrderWorkflow(ctx workflow.Context, in OrderWorkflowState) (OrderWorkflowState, error) {
	logger := workflow.GetLogger(ctx)

	logger.Info("Starting order workflow", "OrderId", in.Id)

	var state OrderWorkflowState

	err := workflow.SetQueryHandler(ctx, QueryOrderByIdName, func() (OrderWorkflowState, error) {
		return state, nil
	})
	if err != nil {
		return state, err
	}

	acceptOrderSelector := workflow.NewSelector(ctx)
	acceptOrderSignal := workflow.GetSignalChannel(ctx, AcceptOrderSignalName)
	acceptOrderSelector.AddReceive(acceptOrderSignal, func(ch workflow.ReceiveChannel, _ bool) {
		var status int32
		ch.Receive(ctx, &status)
		state.Status = pb.OrderStatus(status)
		logger.Info("Received signal", "status", state.Status.String())
	})

	startDeliverySelector := workflow.NewSelector(ctx)
	startDeliverySignal := workflow.GetSignalChannel(ctx, StartDeliverySignalName)
	startDeliverySelector.AddReceive(startDeliverySignal, func(ch workflow.ReceiveChannel, _ bool) {
		var deliverer pb.Wallet
		ch.Receive(ctx, &deliverer)
		state.Deliverer = &deliverer
		state.Status = pb.OrderStatus_Delivering
		logger.Info("Received signal", "status", state.Status.String())
	})

	deliveredOrderSelector := workflow.NewSelector(ctx)
	deliveredOrderSignal := workflow.GetSignalChannel(ctx, DeliveredOrderSignalName)
	deliveredOrderSelector.AddReceive(deliveredOrderSignal, func(ch workflow.ReceiveChannel, _ bool) {
		var status int32
		ch.Receive(ctx, &status)
		state.Status = pb.OrderStatus(status)
		logger.Info("Received signal", "status", state.Status.String())
	})

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 30,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	var activities *Activities

	err = workflow.ExecuteActivity(ctx, activities.NewOrder, in).Get(ctx, &state)
	if err != nil {
		return state, err
	}

	acceptOrderSelector.Select(ctx)

	if state.Status != pb.OrderStatus_Accepted {
		return state, fmt.Errorf("order was not accepted: OrderId=%v, Status=%v", state.Id.Hex(), state.Status.String())
	}

	// activity: notificar cliente que o pedido foi aceito

	startDeliverySelector.Select(ctx)

	if state.Status != pb.OrderStatus_Delivering {
		return state, fmt.Errorf("order is not ready to delivery: OrderId=%v, Status=%v", state.Id.Hex(), state.Status.String())
	}

	// activity: notificar cliente que o pedido saiu para entrega

	deliveredOrderSelector.Select(ctx)

	if state.Status != pb.OrderStatus_Delivered {
		return state, fmt.Errorf("order was not delivered: OrderId=%v, Status=%v", state.Id.Hex(), state.Status.String())
	}

	// activity: notificar vendedor que o cliente recebeu o produto

	logger.Info("Order workflow finished!")

	return state, nil
}
