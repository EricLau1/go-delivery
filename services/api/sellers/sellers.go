package sellers

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
	"time"
)

type sellersHandler struct {
	productsClient pb.ProductsServiceClient
	validate       *validator.Validate
}

func RegisterSellersHandlers(productsClient pb.ProductsServiceClient, m middlewares.Middlewares, router *mux.Router) {

	handler := &sellersHandler{productsClient: productsClient, validate: validator.New()}

	router.Path("/sellers/{id}/products").
		HandlerFunc(
			m.Apply(handler.PostProduct, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Seller,
			}),
		).Methods(http.MethodPost)

	router.Path("/sellers/{id}/products/{product_id}").
		HandlerFunc(
			m.Apply(handler.PutProduct, middlewares.Options{
				AuthRequired: true,
				UserRequired: true,
				RoleRequired: pb.Role_Seller,
			}),
		).Methods(http.MethodPut)

	router.Path("/sellers/{id}/products").
		HandlerFunc(
			m.Apply(handler.GetProductsBySeller, middlewares.Options{}),
		).Methods(http.MethodGet)

	router.Path("/products/{product_id}").
		HandlerFunc(
			m.Apply(handler.GetProduct, middlewares.Options{}),
		).Methods(http.MethodGet)

	router.Path("/products").
		HandlerFunc(
			m.Apply(handler.GetProducts, middlewares.Options{}),
		).Methods(http.MethodGet)

}

func (h *sellersHandler) PostProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sellerId, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	input := new(form.ProductInput)
	err = json.Unmarshal(body, input)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	input.Clear()

	err = h.validate.Struct(input)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	product := &pb.Product{
		Id:           primitive.NewObjectID().Hex(),
		SellerId:     sellerId.Hex(),
		Name:         input.Name,
		Price:        input.Price,
		DeliveryCost: input.DeliveryCost,
		Quantity:     input.Quantity,
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}

	product, err = h.productsClient.CreateProduct(r.Context(), product)
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusCreated, form.FromProduct(product))
}

func (h *sellersHandler) PutProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productId, err := primitive.ObjectIDFromHex(vars["product_id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	input := new(form.ProductInput)
	err = json.Unmarshal(body, input)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	input.Clear()

	err = h.validate.Struct(input)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	update := &pb.UpdateProductRequest{
		Id:           productId.Hex(),
		Name:         input.Name,
		Price:        input.Price,
		DeliveryCost: input.DeliveryCost,
		Quantity:     input.Quantity,
	}

	product, err := h.productsClient.UpdateProduct(r.Context(), update)
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, form.FromProduct(product))
}

func (h *sellersHandler) GetProductsBySeller(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sellerId, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	get := &pb.ListSellerProductsRequest{SellerId: sellerId.Hex()}

	stream, err := h.productsClient.ListSellerProducts(r.Context(), get)
	if err != nil {
		rest.WriteError(w, http.StatusNotFound, err)
		return
	}

	var products []*form.Product

	for {
		product, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			rest.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		products = append(products, form.FromProduct(product))
	}

	rest.WriteAsJson(w, http.StatusOK, products)
}

func (h *sellersHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productId, err := primitive.ObjectIDFromHex(vars["product_id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	get := &pb.GetProductRequest{
		Id: productId.Hex(),
	}

	product, err := h.productsClient.GetProduct(r.Context(), get)
	if err != nil {
		rest.WriteError(w, http.StatusNotFound, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, product)
}

func (h *sellersHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	get := &pb.ListProductsRequest{}

	stream, err := h.productsClient.ListProducts(r.Context(), get)
	if err != nil {
		rest.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var products []*form.Product

	for {
		product, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			rest.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		products = append(products, form.FromProduct(product))
	}

	rest.WriteAsJson(w, http.StatusOK, products)
}
