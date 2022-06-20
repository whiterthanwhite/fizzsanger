package auth

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type (
	UserCredentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	MyCustomClaims struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		jwt.StandardClaims
	}
)

func CreateToken(claims MyCustomClaims) (string, error) {
	signingKey := []byte("qwerty")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	convertedString, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return convertedString, nil
}

func ParseToken(token string, claims *MyCustomClaims) (*jwt.Token, error) {
	newToken, err := jwt.ParseWithClaims(token, claims, myKeyFunc)
	if err != nil {
		return nil, err
	}

	return newToken, nil
}

func CreateCustomClaims(login, password string) MyCustomClaims {
	return MyCustomClaims{
		login,
		password,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute).Unix(),
		},
	}
}

func myKeyFunc(token *jwt.Token) (interface{}, error) {
	return []byte("qwerty"), nil
}
