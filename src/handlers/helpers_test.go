package handlers

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInjectHostPortInBody(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		tcpServerHost  string
		tcpServerPort  string
		expectedBody   map[string]interface{}
		expectedErrMsg string
	}{
		{
			name:          "Valid JSON body",
			body:          `{"key": "value"}`,
			tcpServerHost: "localhost",
			tcpServerPort: "8080",
			expectedBody: map[string]interface{}{
				"key":         "value",
				"STREAM_HOST": "localhost",
				"STREAM_PORT": "8080",
			},
			expectedErrMsg: "",
		},
		{
			name:           "Empty JSON body",
			body:           `{}`,
			tcpServerHost:  "localhost",
			tcpServerPort:  "8080",
			expectedBody:   map[string]interface{}{"STREAM_HOST": "localhost", "STREAM_PORT": "8080"},
			expectedErrMsg: "",
		},
		{
			name:           "Invalid JSON body",
			body:           `{"key": "value"`,
			tcpServerHost:  "localhost",
			tcpServerPort:  "8080",
			expectedBody:   nil,
			expectedErrMsg: "unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			require.NoError(t, err)

			actualBody, err := injectHostPortInBody(req, tt.tcpServerHost, tt.tcpServerPort)
			if tt.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, actualBody)

			require.Equal(t, actualBody, tt.expectedBody)
		})
	}
}

func TestGetNamespaceAndAction(t *testing.T) {
	tests := []struct {
		name           string
		action         string
		pkg            string
		expectedNs     string
		expectedAction string
	}{
		{
			name:           "Namespace and action without package",
			action:         "action1",
			pkg:            "",
			expectedNs:     "ns1",
			expectedAction: "action1",
		},
		{
			name:           "Namespace, package and action",
			action:         "action1",
			pkg:            "pkg1",
			expectedNs:     "ns1",
			expectedAction: "pkg1/action1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "", nil)
			require.NoError(t, err)

			if tt.pkg != "" {
				req.SetPathValue("pkg", tt.pkg)
			}
			req.SetPathValue("ns", tt.expectedNs)
			req.SetPathValue("action", tt.action)

			ns, action := getNamespaceAndAction(req)
			require.Equal(t, tt.expectedNs, ns)
			require.Equal(t, tt.expectedAction, action)
		})
	}
}
