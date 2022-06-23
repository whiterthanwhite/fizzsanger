package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetConf(t *testing.T) {
	type envData struct {
		envName string
		envVal  string
	}

	tests := []struct {
		name    string
		envData []envData
		isNil   bool
	}{
		{
			name:    "nil test",
			envData: nil,
			isNil:   true,
		},
		{
			name: "set env",
			envData: []envData{
				{
					envName: signingKeyEnv,
					envVal:  "TestSigningKey",
				},
				{
					envName: tokenExpiresAt,
					envVal:  "1",
				},
			},
			isNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, ed := range tt.envData {
				err := os.Setenv(ed.envName, ed.envVal)
				require.Nil(t, err)
			}
			conf := GetConf()
			if tt.isNil {
				require.Nil(t, conf)
			} else {
				require.NotNil(t, conf)
			}
			fmt.Println(conf)
		})
	}
}
