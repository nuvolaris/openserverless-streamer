package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAsyncPostWebAction(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		body           []byte
		expectedErrMsg string
		handler        http.HandlerFunc
	}{
		{
			name: "Successful request",
			url:  "/success",
			body: []byte(`{"key": "value"}`),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name:           "Error in request creation",
			url:            "://invalid-url",
			body:           []byte(`{"key": "value"}`),
			expectedErrMsg: "parse \"://invalid-url\": missing protocol scheme",
		},
		{
			name: "Non-200 status code",
			url:  "/error",
			body: []byte(`{"key": "value"}`),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedErrMsg: "Error invoking action: 500 Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errChan := make(chan error, 1)

			if tt.handler != nil {
				server := httptest.NewServer(tt.handler)
				defer server.Close()
				tt.url = server.URL + tt.url
			}

			go asyncPostWebAction(errChan, tt.url, tt.body)

			err := <-errChan
			if tt.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
