package tokens

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"go-delivery/pb"
	"os"
	"time"
)

const tokenExpiration = time.Hour * 24 * 30

var jwtSecretKey = os.Getenv("JWT_SECRET_KEY")

type TokenPayload struct {
	Id        string
	Role      string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Issuer    string
}

func New(user *pb.User) (string, error) {
	issuedAt := time.Now()

	claims := jwt.StandardClaims{
		Audience:  user.Role.String(),
		ExpiresAt: issuedAt.Add(tokenExpiration).Unix(),
		IssuedAt:  issuedAt.Unix(),
		Issuer:    os.Getenv("APP_NAME"),
		Subject:   user.Id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)

	return token.SignedString([]byte(jwtSecretKey))
}

func getKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(jwtSecretKey), nil
}

func Parse(raw string) (*TokenPayload, error) {
	claims := new(jwt.StandardClaims)

	token, err := jwt.ParseWithClaims(raw, claims, getKey)
	if err != nil {
		return nil, err
	}

	var ok bool
	claims, ok = token.Claims.(*jwt.StandardClaims)
	if !token.Valid || !ok {
		return nil, errors.New("invalid token")
	}

	payload := TokenPayload{
		Id:        claims.Subject,
		Role:      claims.Audience,
		IssuedAt:  time.Unix(claims.IssuedAt, 0),
		ExpiresAt: time.Unix(claims.ExpiresAt, 0),
		Issuer:    claims.Issuer,
	}

	return &payload, nil
}
