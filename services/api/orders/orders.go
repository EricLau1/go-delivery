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
	"strconv"
	"time"
)

type ordersHandler struct {
	ordersClient pb.OrdersServiceClient
	validate     *validator.Validate
}

func RegisterOrdersHandlers(ordersClient pb.OrdersServiceClient, m middlewares.Middlewares, router *mux.Router) {
	handler := ordersHandler{ordersClient: ordersClient, validate: validator.New()}

	router.Path("/orders/customers/{id}").
		HandlerFunc(
			m.Apply(handler.PostOrder, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Customer,
			}),
		).Methods(http.MethodPost)

	router.Path("/orders/sellers/{id}").
		HandlerFunc(
			m.Apply(handler.GetSellerOrders, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Seller,
			}),
		).Methods(http.MethodGet)

	router.Path("/orders/deliverers/{id}").
		HandlerFunc(
			m.Apply(handler.GetOrdersAccepted, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Delivery,
			}),
		).Methods(http.MethodGet)

	router.Path("/orders/{order_id}/sellers/{id}").
		HandlerFunc(
			m.Apply(handler.PutApproveOrder, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Seller,
			}),
		).Methods(http.MethodPut)

	router.Path("/orders/{order_id}/deliverers/{id}").
		HandlerFunc(
			m.Apply(handler.PutDeliverOrder, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Delivery,
			}),
		).Methods(http.MethodPut)

	router.Path("/orders/{order_id}/customers/{id}").
		HandlerFunc(
			m.Apply(handler.PutOrderDelivered, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Customer,
			}),
		).Methods(http.MethodPut)

	router.Path("/orders/{order_id}/admins/{id}").
		HandlerFunc(
			m.Apply(handler.GetOrder, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Admin,
			}),
		).Methods(http.MethodGet)

	router.Path("/orders/{order_id}/admins/{id}").
		HandlerFunc(
			m.Apply(handler.DeleteOrder, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Admin,
			}),
		).Methods(http.MethodDelete)

	router.Path("/orders/admins/{id}").
		HandlerFunc(
			m.Apply(handler.GetOrders, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Admin,
			}),
		).Methods(http.MethodGet)

	router.Path("/orders/status/{status}/admins/{id}").
		HandlerFunc(
			m.Apply(handler.GetByStatus, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Admin,
			}),
		).Methods(http.MethodGet)
}

func (h *ordersHandler) PostOrder(w http.ResponseWriter, r *http.Request) {
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

	order := &pb.Order{
		Id:         primitive.NewObjectID().Hex(),
		CustomerId: customerId.Hex(),
		SellerId:   input.SellerId,
		ProductId:  input.ProductId,
		Quantity:   input.Quantity,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}

	order, err = h.ordersClient.CreateOrder(r.Context(), order)
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusCreated, form.FromOrder(order))
}

func (h *ordersHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderId, err := primitive.ObjectIDFromHex(vars["order_id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	order, err := h.ordersClient.GetOrder(r.Context(), &pb.GetOrderRequest{Id: orderId.Hex()})
	if err != nil {
		rest.WriteError(w, http.StatusNotFound, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, form.FromOrder(order))
}

func (h *ordersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	stream, err := h.ordersClient.ListOrders(r.Context(), &pb.ListOrdersRequest{})
	if err != nil {
		rest.WriteError(w, http.StatusNotFound, err)
		return
	}

	var orders []*form.Order

	for {
		order, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			rest.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		orders = append(orders, form.FromOrder(order))
	}

	rest.WriteAsJson(w, http.StatusOK, orders)
}

func (h *ordersHandler) GetByStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	status, err := strconv.Atoi(vars["status"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	list := &pb.ListOrdersByStatusRequest{Status: pb.OrderStatus(status)}

	stream, err := h.ordersClient.ListOrdersByStatus(r.Context(), list)
	if err != nil {
		rest.WriteError(w, http.StatusNotFound, err)
		return
	}

	var orders []*form.Order

	for {
		order, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			rest.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		orders = append(orders, form.FromOrder(order))
	}

	rest.WriteAsJson(w, http.StatusOK, orders)
}

func (h *ordersHandler) GetOrdersAccepted(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	list := &pb.ListOrdersByStatusRequest{Status: pb.OrderStatus_Accepted}

	stream, err := h.ordersClient.ListOrdersByStatus(r.Context(), list)
	if err != nil {
		rest.WriteError(w, http.StatusNotFound, err)
		return
	}

	var orders []*form.Order

	for {
		order, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			rest.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		orders = append(orders, form.FromOrder(order))
	}

	rest.WriteAsJson(w, http.StatusOK, orders)
}

func (h *ordersHandler) GetSellerOrders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sellerId, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	list := &pb.ListOrdersBySellerRequest{SellerId: sellerId.Hex()}

	stream, err := h.ordersClient.ListOrdersBySeller(r.Context(), list)
	if err != nil {
		rest.WriteError(w, http.StatusNotFound, err)
		return
	}

	var orders []*form.Order

	for {
		order, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			rest.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		orders = append(orders, form.FromOrder(order))
	}

	rest.WriteAsJson(w, http.StatusOK, orders)
}

func (h *ordersHandler) PutApproveOrder(w http.ResponseWriter, r *http.Request) {
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

	order, err := h.ordersClient.ApproveOrder(r.Context(), &pb.ApproveOrderRequest{
		Id:       orderId.Hex(),
		SellerId: sellerId.Hex(),
	})
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, form.FromOrder(order))
}

func (h *ordersHandler) PutDeliverOrder(w http.ResponseWriter, r *http.Request) {
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

	order, err := h.ordersClient.DeliverOrder(r.Context(), &pb.DeliverOrderRequest{
		Id:         orderId.Hex(),
		DeliveryId: delivererId.Hex(),
	})
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, form.FromOrder(order))
}

func (h *ordersHandler) PutOrderDelivered(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderId, err := primitive.ObjectIDFromHex(vars["order_id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	customerId, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	order, err := h.ordersClient.ConfirmOrderDelivered(r.Context(), &pb.ConfirmOrderDeliveredRequest{
		Id:         orderId.Hex(),
		CustomerId: customerId.Hex(),
	})
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, form.FromOrder(order))
}

func (h *ordersHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderId, err := primitive.ObjectIDFromHex(vars["order_id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	_, err = h.ordersClient.DeleteOrder(r.Context(), &pb.DeleteOrderRequest{Id: orderId.Hex()})
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusNoContent, nil)
}
