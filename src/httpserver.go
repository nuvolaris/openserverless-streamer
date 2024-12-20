// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
)

func startHTTPServer(streamingProxyAddr string, apihost string) {
	httpPort := os.Getenv("HTTP_SERVER_PORT")
	if httpPort == "" {
		httpPort = "80"
	}

	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Streamer proxy running"))
	})
	// router.HandleFunc("POST /web/{ns}/{action}", handleWebActionStream(streamingProxyAddr, apihost))
	// router.HandleFunc("POST /web/{ns}/{pkg}/{action}", handleWebActionStream(streamingProxyAddr, apihost))
	router.HandleFunc("POST /action/{ns}/{action}", handleActionStream(streamingProxyAddr, apihost))
	router.HandleFunc("POST /action/{ns}/{pkg}/{action}", handleActionStream(streamingProxyAddr, apihost))

	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	log.Println("HTTP server listening on port", httpPort)
	if err := server.ListenAndServe(); err != nil {
		log.Println("Error starting HTTP server:", err)
	}
}

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
