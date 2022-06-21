package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/whiterthanwhite/fizzsanger/internal/config"
)

type (
	UserCredentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	MyCustomClaims struct {
		Login string `json:"login"`
		jwt.StandardClaims
	}
)

func CreateToken(claims MyCustomClaims, conf *config.Conf) (string, error) {
	if conf == nil {
		return "", errors.New("empty configuration file")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	convertedString, err := token.SignedString(conf.SigningKey)
	if err != nil {
		return "", err
	}

	return convertedString, nil
}

func ParseToken(token string, claims *MyCustomClaims, conf *config.Conf) (*jwt.Token, error) {
	newToken, err := jwt.ParseWithClaims(token, claims, myKeyFunc(conf))
	if err != nil {
		return nil, err
	}

	return newToken, nil
}

func CreateCustomClaims(login string, conf *config.Conf) MyCustomClaims {
	addTime := time.Minute
	if conf != nil {
		addTime = conf.TokenExpiresAt
	}

	return MyCustomClaims{
		login,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(addTime).Unix(),
		},
	}
}

func myKeyFunc(conf *config.Conf) jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		return conf.SigningKey, nil
	}
}
