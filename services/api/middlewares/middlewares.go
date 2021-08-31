package middlewares

import (
	"github.com/gorilla/mux"
	"go-delivery/pb"
	"go-delivery/services/api/rest"
	"log"
	"net/http"
)

type Middlewares interface {
	EnsureAuthentication(next http.HandlerFunc) http.HandlerFunc
	EnsureUser(next http.HandlerFunc) http.HandlerFunc
	Apply(next http.HandlerFunc, opt Options) http.HandlerFunc
}

type impl struct {
	accountsService pb.AccountsServiceClient
}

func New(accountsService pb.AccountsServiceClient) Middlewares {
	return &impl{accountsService: accountsService}
}

type Options struct {
	AuthRequired bool
	UserRequired bool
	RoleRequired pb.Role
}

func (i *impl) Apply(next http.HandlerFunc, opt Options) http.HandlerFunc {
	if opt.AuthRequired {
		next = i.EnsureAuthentication(next)
	}

	if opt.UserRequired {
		next = i.EnsureUser(next)
	}

	switch opt.RoleRequired {
	case pb.Role_Customer:
		next = i.OnlyCustomer(next)
	case pb.Role_Seller:
		next = i.OnlySeller(next)
	case pb.Role_Delivery:
		next = i.OnlyDelivery(next)
	case pb.Role_Admin:
		next = i.OnlyAdmin(next)
	}

	return next
}

func (i *impl) EnsureAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		_, err := rest.GetToken(r)
		if err != nil {
			log.Println("invalid token: ", err.Error())

			WriteUnauthorized(w)
			return
		}

		next(w, r)
	}
}

func (i *impl) EnsureUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token, err := rest.GetToken(r)
		if err != nil {
			log.Println("invalid token: ", err.Error())
			WriteUnauthorized(w)
			return
		}

		vars := mux.Vars(r)
		if vars["id"] != token.Id {
			WriteUnauthorized(w)
			return
		}

		user, err := i.accountsService.GetUser(r.Context(), &pb.GetUserRequest{Id: token.Id})
		if err != nil {
			WriteUnauthorized(w)
			return
		}

		if user.Role.String() != token.Role {
			WriteUnauthorized(w)
			return
		}

		next(w, r)
	}
}

func (i *impl) OnlyAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token, err := rest.GetToken(r)
		if err != nil {
			log.Println("invalid token: ", err.Error())
			WriteUnauthorized(w)
			return
		}

		role := pb.Role_value[token.Role]

		if role != int32(pb.Role_Admin) {
			WriteUnauthorized(w)
			return
		}

		next(w, r)
	}
}

func (i *impl) OnlySeller(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token, err := rest.GetToken(r)
		if err != nil {
			log.Println("invalid token: ", err.Error())
			WriteUnauthorized(w)
			return
		}

		role := pb.Role_value[token.Role]

		if role != int32(pb.Role_Seller) {
			WriteUnauthorized(w)
			return
		}

		next(w, r)
	}
}

func (i *impl) OnlyDelivery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token, err := rest.GetToken(r)
		if err != nil {
			log.Println("invalid token: ", err.Error())
			WriteUnauthorized(w)
			return
		}

		role := pb.Role_value[token.Role]

		if role != int32(pb.Role_Delivery) {
			WriteUnauthorized(w)
			return
		}

		next(w, r)
	}
}

func (i *impl) OnlyCustomer(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		token, err := rest.GetToken(r)
		if err != nil {
			log.Println("invalid token: ", err.Error())
			WriteUnauthorized(w)
			return
		}

		role := pb.Role_value[token.Role]

		if role != int32(pb.Role_Customer) {
			WriteUnauthorized(w)
			return
		}

		next(w, r)
	}
}

func WriteUnauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte("Unauthorized"))
}
