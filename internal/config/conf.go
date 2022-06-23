package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

const (
	signingKeyEnv  string = "SIGNING_KEY"
	tokenExpiresAt string = "TOKEN_EXPIRES_AT"
	dbAdress       string = "DB_ADDRESS"
)

type Conf struct {
	SigningKey     []byte
	TokenExpiresAt time.Duration
	DBAddress      string
}

func GetConf() *Conf {
	values := make(map[string]string)
	values[signingKeyEnv] = ""
	values[tokenExpiresAt] = ""
	values[dbAdress] = ""
	for k := range values {
		var ok bool
		if values[k], ok = getEnvValue(k); !ok {
			return nil
		}
	}

	d, err := strconv.ParseInt(values[tokenExpiresAt], 0, 64)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return &Conf{
		SigningKey:     []byte(values[signingKeyEnv]),
		TokenExpiresAt: time.Minute * time.Duration(d),
		DBAddress:      values[dbAdress],
	}
}

func getEnvValue(envName string) (string, bool) {
	envVal, exist := os.LookupEnv(envName)
	if !exist {
		log.Printf("%s is not setted!\n", envName)
	}
	return envVal, exist
}
