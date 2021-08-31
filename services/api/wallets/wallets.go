package wallets

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

type walletsHandler struct {
	walletsClient pb.WalletsServiceClient
	validate      *validator.Validate
}

func RegisterWalletsHandlers(walletsClient pb.WalletsServiceClient, m middlewares.Middlewares, router *mux.Router) {
	handlers := walletsHandler{walletsClient: walletsClient, validate: validator.New()}

	router.Path("/users/{id}/wallets").HandlerFunc(
		m.Apply(handlers.CreateWallet, middlewares.Options{
			AuthRequired: true,
			UserRequired: true,
		}),
	).Methods(http.MethodPost)

	router.Path("/users/{id}/wallets/{wallet_id}").HandlerFunc(
		m.Apply(handlers.PostCredit, middlewares.Options{
			AuthRequired: true,
			UserRequired: true,
		}),
	).Methods(http.MethodPut)

	router.Path("/users/{id}/wallets").HandlerFunc(
		m.Apply(handlers.GetUserWallet, middlewares.Options{
			AuthRequired: true,
			UserRequired: true,
		}),
	).Methods(http.MethodGet)

	router.Path("/wallets/{id}").HandlerFunc(
		m.Apply(handlers.GetWallet, middlewares.Options{
			AuthRequired: true,
			RoleRequired: pb.Role_Admin,
		}),
	).Methods(http.MethodGet)

	router.Path("/wallets").HandlerFunc(
		m.Apply(handlers.GetWallets, middlewares.Options{
			AuthRequired: true,
			RoleRequired: pb.Role_Admin,
		}),
	).Methods(http.MethodGet)
}

func (h *walletsHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	wallet := &pb.Wallet{
		Id:        primitive.NewObjectID().Hex(),
		UserId:    userId.Hex(),
		Cash:      0,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	wallet, err = h.walletsClient.CreateWallet(r.Context(), wallet)
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusCreated, form.FromWallet(wallet))
}

func (h *walletsHandler) PostCredit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletId, err := primitive.ObjectIDFromHex(vars["wallet_id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	input := new(form.CreditInput)
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

	credit := &pb.CreditRequest{
		WalletId: walletId.Hex(),
		Amount:   input.Amount,
	}

	wallet, err := h.walletsClient.Credit(r.Context(), credit)
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, form.FromWallet(wallet))
}

func (h *walletsHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletId, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	get := &pb.GetWalletRequest{Id: walletId.Hex()}
	wallet, err := h.walletsClient.GetWallet(r.Context(), get)
	if err != nil {
		rest.WriteError(w, http.StatusNotFound, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, form.FromWallet(wallet))
}

func (h *walletsHandler) GetUserWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	get := &pb.GetUserWalletRequest{UserId: userId.Hex()}
	wallet, err := h.walletsClient.GetUserWallet(r.Context(), get)
	if err != nil {
		rest.WriteError(w, http.StatusNotFound, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, form.FromWallet(wallet))
}

func (h *walletsHandler) GetWallets(w http.ResponseWriter, r *http.Request) {
	stream, err := h.walletsClient.ListWallets(r.Context(), &pb.ListWalletsRequest{})
	if err != nil {
		rest.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var wallets []*form.Wallet

	for {
		wallet, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			rest.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		wallets = append(wallets, form.FromWallet(wallet))
	}

	rest.WriteAsJson(w, http.StatusOK, wallets)
}
