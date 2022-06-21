package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/whiterthanwhite/fizzsanger/internal/config"
)

func TestCreateCustomClaims(t *testing.T) {
	type claimsData struct {
		login string
		conf  *config.Conf
	}

	tests := []struct {
		name       string
		claimsData claimsData
	}{
		{
			name: "empty login",
			claimsData: claimsData{
				login: "",
				conf:  nil,
			},
		},
		{
			name: "not empty login",
			claimsData: claimsData{
				login: "testLogin",
				conf:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := CreateCustomClaims(tt.claimsData.login, tt.claimsData.conf)
			require.Equal(t, tt.claimsData.login, claims.Login)
		})
	}
}

func TestCreateToken(t *testing.T) {
	type want struct {
		tokenErr error
	}

	tests := []struct {
		name  string
		login string
		conf  *config.Conf
		want  want
	}{
		{
			name:  "test 1",
			login: "testUser",
			conf:  nil,
			want: want{
				tokenErr: errors.New("empty configuration file"),
			},
		},
		{
			name:  "test 2",
			login: "testUser",
			conf: &config.Conf{
				SigningKey:     []byte("testkey"),
				TokenExpiresAt: time.Minute,
			},
			want: want{
				tokenErr: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := CreateCustomClaims(tt.login, tt.conf)
			_, err := CreateToken(claims, tt.conf)
			require.Equal(t, tt.want.tokenErr, err)
		})
	}
}
