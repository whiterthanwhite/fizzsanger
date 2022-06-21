package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserRegister(t *testing.T) {
	type (
		reqParams struct {
			method string
			target string
			body   []byte
		}
		resParams struct {
			statusCode  int
			contentType string
		}
	)

	tests := []struct {
		name      string
		reqParams reqParams
		resParams resParams
	}{
		{
			name: "test wrong method",
			reqParams: reqParams{
				method: http.MethodGet,
				target: "/register",
				body:   nil,
			},
			resParams: resParams{
				statusCode:  http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "test successfull",
			reqParams: reqParams{
				method: http.MethodPost,
				target: "/register",
				body:   []byte("{\"login\":\"login\",\"password\":\"password\"}"),
			},
			resParams: resParams{
				statusCode:  http.StatusOK,
				contentType: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.reqParams.method, tt.reqParams.target,
				bytes.NewBuffer(tt.reqParams.body))
			request.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(UserRegister)
			handler.ServeHTTP(recorder, request)
			result := recorder.Result()
			defer result.Body.Close()

			var err error
			var rBody []byte

			if rBody, err = io.ReadAll(recorder.Body); err != nil {
				t.Error(err.Error())
			}
			t.Log(string(rBody))

			assert.Equal(t, tt.resParams.statusCode, result.StatusCode)
			assert.Equal(t, tt.resParams.contentType, result.Header.Get("Content-Type"))
		})
	}
}
