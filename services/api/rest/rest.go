package rest

import (
	"encoding/json"
	"go-delivery/security/tokens"
	"net/http"
	"strings"
)

type Err struct {
	String string `json:"error"`
}

func NewErr(e error) *Err {
	err := Err{String: "error"}
	if e != nil {
		err.String = e.Error()
	}
	return &err
}

func WriteAsJson(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, statusCode int, err error) {
	WriteAsJson(w, statusCode, NewErr(err))
}

func GetToken(r *http.Request) (*tokens.TokenPayload, error) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	return tokens.Parse(token)
}