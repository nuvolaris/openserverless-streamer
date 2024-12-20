package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func injectHostPortInBody(r *http.Request, tcpServerHost string, tcpServerPort string) (map[string]interface{}, error) {
	body := r.Body
	defer body.Close()

	jsonBody := make(map[string]interface{})
	if err := json.NewDecoder(body).Decode(&jsonBody); err != nil {
		return nil, err
	}

	jsonBody["STREAM_HOST"] = tcpServerHost
	jsonBody["STREAM_PORT"] = tcpServerPort
	return jsonBody, nil
}

func getNamespaceAndAction(r *http.Request) (string, string) {
	namespace := r.PathValue("ns")
	pkg := r.PathValue("pkg")
	action := r.PathValue("action")

	actionToInvoke := action
	if pkg != "" {
		actionToInvoke = pkg + "/" + action
	}

	return namespace, actionToInvoke
}

func extractAuthToken(r *http.Request) (string, error) {
	apiKey := r.Header.Get("Authorization")
	if apiKey == "" {
		return "", errors.New("Missing Authorization header")
	}

	// get the apikey without the Bearer prefix
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")
	return apiKey, nil
}
