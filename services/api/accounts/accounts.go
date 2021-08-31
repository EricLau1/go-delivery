package accounts

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go-delivery/pb"
	"go-delivery/security/passwords"
	"go-delivery/services/api/middlewares"
	"go-delivery/services/api/rest"
	"go-delivery/services/api/rest/form"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"time"
)

type accountsHandler struct {
	authClient pb.AccountsServiceClient
	validate   *validator.Validate
}

func RegisterAccountsHandlers(authClient pb.AccountsServiceClient, m middlewares.Middlewares, router *mux.Router) {

	h := &accountsHandler{
		authClient: authClient,
		validate:   validator.New(),
	}

	router.Path("/signup").HandlerFunc(h.PostSignUp).Methods(http.MethodPost)
	router.Path("/signin").HandlerFunc(h.PostSignIn).Methods(http.MethodPost)
	router.Path("/token").HandlerFunc(h.ValidateToken).Methods(http.MethodGet)

	router.Path("/users/{id}").HandlerFunc(m.Apply(h.GetUser, middlewares.Options{
		AuthRequired: true,
		UserRequired: true,
	})).Methods(http.MethodGet)

	router.Path("/users").HandlerFunc(m.Apply(h.GetUsers, middlewares.Options{
		AuthRequired: true,
		RoleRequired: pb.Role_Admin,
	})).Methods(http.MethodGet)
}

func (h *accountsHandler) PostSignUp(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	input := new(form.SignUpInput)

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

	user := new(pb.User)

	user.Password, err = passwords.New(input.Password)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	user.Id = primitive.NewObjectID().Hex()
	user.Email = input.Email
	user.Role = pb.Role(input.Role)
	user.CreatedAt = time.Now().Unix()
	user.UpdatedAt = user.CreatedAt

	user, err = h.authClient.SignUp(r.Context(), user)
	if err != nil {
		rest.WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	rest.WriteAsJson(w, http.StatusCreated, form.FromUser(user))
}

func (h *accountsHandler) PostSignIn(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	input := new(form.SignInInput)

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

	res, err := h.authClient.SignIn(r.Context(), input.ToProto())
	if err != nil {
		rest.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, res)
}

func (h *accountsHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	payload, err := rest.GetToken(r)
	if err != nil {
		rest.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	req := &pb.GetUserRequest{Id: payload.Id}

	user, err := h.authClient.GetUser(r.Context(), req)
	if err != nil {
		rest.WriteError(w, http.StatusUnauthorized, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, form.FromUser(user))
}

func (h *accountsHandler) GetUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		rest.WriteError(w, http.StatusBadRequest, err)
		return
	}

	req := &pb.GetUserRequest{Id: id.Hex()}

	user, err := h.authClient.GetUser(r.Context(), req)
	if err != nil {
		rest.WriteError(w, http.StatusNotFound, err)
		return
	}

	rest.WriteAsJson(w, http.StatusOK, form.FromUser(user))
}

func (h *accountsHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	stream, err := h.authClient.ListUsers(r.Context(), &pb.ListUsersRequest{})
	if err != nil {
		rest.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var users []*form.User

	for {

		user, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			rest.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		users = append(users, form.FromUser(user))
	}

	rest.WriteAsJson(w, http.StatusOK, users)
}
