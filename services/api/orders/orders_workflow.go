package orders

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go-delivery/pb"
	"go-delivery/services/api/middlewares"
	"go-delivery/services/api/rest"
	"go-delivery/services/api/rest/form"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
)

type ordersWorkflowHandler struct {
	ordersWfClient pb.OrdersWorkflowServiceClient
	validate       *validator.Validate
}

func RegisterOrdersWorkflowHandlers(ordersWfClient pb.OrdersWorkflowServiceClient, m middlewares.Middlewares, router *mux.Router) {
	handler := &ordersWorkflowHandler{ordersWfClient: ordersWfClient, validate: validator.New()}

	router.
		Path("/workflows/customers/{id}/orders").
		HandlerFunc(m.Apply(handler.PostStartOrder, middlewares.Options{
			AuthRequired: true,
			UserRequired: true,
			RoleRequired: pb.Role_Customer,
		})).Methods(http.MethodPost)

	router.
		Path("/workflows/orders/{order_id}").
		HandlerFunc(m.Apply(handler.GetOrder, middlewares.Options{
			AuthRequired: true,
		})).Methods(http.MethodGet)

	router.
		Path("/workflows/sellers/{id}/orders/{order_id}").
		HandlerFunc(m.Apply(handler.PutAcceptOrder, middlewares.Options{
			AuthRequired: true,
			UserRequired: true,
			RoleRequired: pb.Role_Seller,
		})).Methods(http.MethodPut)

	router.
		Path("/workflows/deliverers/{id}/orders/{order_id}").
		HandlerFunc(m.Apply(handler.PutStartDeliveryOrder, middlewares.Options{
			AuthRequired: true,
			UserRequired: true,
			RoleRequired: pb.Role_Delivery,
		})).Methods(http.MethodPut)

	router.
		Path("/workflows/customers/{id}/orders/{order_id}").
		HandlerFunc(m.Apply(handler.PutConfirmDeliveredOrder, middlewares.Options{
			AuthRequired: true,
			UserRequired: true,
			RoleRequired: pb.Role_Customer,
		})).Methods(http.MethodPut)
}

func (h *ordersWorkflowHandler) PostStartOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerId, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	input := new(form.OrderInput)
	err = json.Unmarshal(body, input)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.validate.Struct(input)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	start := &pb.StartOrderRequest{
		CustomerId: customerId.Hex(),
		SellerId:   input.SellerId,
		ProductId:  input.ProductId,
		Quantity:   input.Quantity,
	}

	order, err := h.ordersWfClient.StartOrder(r.Context(), start)
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusCreated, order)
}

func (h *ordersWorkflowHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderId, err := primitive.ObjectIDFromHex(vars["order_id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	query := &pb.QueryOrderRequest{OrderId: orderId.Hex()}

	order, err := h.ordersWfClient.QueryOrder(r.Context(), query)
	if err != nil {
		rest.WriteError(w, http.StatusNotFound, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, form.FromOrder(order))
}

func (h *ordersWorkflowHandler) PutAcceptOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	orderId, err := primitive.ObjectIDFromHex(vars["order_id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	sellerId, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	accept := &pb.AcceptOrderRequest{
		Id:       orderId.Hex(),
		SellerId: sellerId.Hex(),
	}

	order, err := h.ordersWfClient.AcceptOrder(r.Context(), accept)
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, order)
}

func (h *ordersWorkflowHandler) PutStartDeliveryOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	orderId, err := primitive.ObjectIDFromHex(vars["order_id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	delivererId, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	start := &pb.StartDeliveryRequest{
		OrderId:     orderId.Hex(),
		DelivererId: delivererId.Hex(),
	}

	order, err := h.ordersWfClient.StartDelivery(r.Context(), start)
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, order)
}

func (h *ordersWorkflowHandler) PutConfirmDeliveredOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	orderId, err := primitive.ObjectIDFromHex(vars["order_id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	_, err = primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	confirm := &pb.ConfirmDeliveredOrderRequest{
		OrderId: orderId.Hex(),
	}

	order, err := h.ordersWfClient.ConfirmDeliveredOrder(r.Context(), confirm)
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, order)
}
