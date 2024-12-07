package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
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
	router.HandleFunc("POST /stream/{ns}/{action}", handleHTTPStream(streamingProxyAddr, apihost))
	router.HandleFunc("POST /stream/{ns}/{pkg}/{action}", handleHTTPStream(streamingProxyAddr, apihost))

	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	log.Println("HTTP server listening on port", httpPort)
	if err := server.ListenAndServe(); err != nil {
		log.Println("Error starting HTTP server:", err)
	}
}

func handleHTTPStream(streamingProxyAddr string, apihost string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, done := context.WithCancel(context.Background())

		streamDataChan := make(chan []byte)

		namespace := r.PathValue("ns")
		pkg := r.PathValue("pkg")
		action := r.PathValue("action")

		actionToInvoke := action
		if pkg != "" {
			actionToInvoke = pkg + "/" + action
		}

		log.Println("Received request for", namespace, pkg, action)

		apiKey := r.Header.Get("Authorization")
		if apiKey == "" {
			http.Error(w, "Missing Authorization header", http.StatusBadRequest)
			done()
			return
		}

		// get the apikey without the Bearer prefix
		apiKey = strings.TrimPrefix(apiKey, "Bearer ")

		// Create OpenWhisk client
		client := NewOpenWhiskClient(apihost, apiKey, namespace)

		// opens a socket for listening in a random port
		socketServer, err := startTCPServer(ctx, streamingProxyAddr, streamDataChan)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			done()
			return
		}
		go socketServer.WaitToCleanUp()

		tcpServerHost, tcpServerPort, err := net.SplitHostPort(socketServer.listener.Addr().String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			done()
			return
		}

		// parse the json body and add STREAM_HOST and STREAM_PORT
		body := r.Body
		defer body.Close()

		jsonBody := make(map[string]interface{})
		err = json.NewDecoder(body).Decode(&jsonBody)
		if err != nil {
			http.Error(w, "Error decoding JSON body: "+err.Error(), http.StatusInternalServerError)
			done()
			return
		}

		jsonBody["STREAM_HOST"] = tcpServerHost
		jsonBody["STREAM_PORT"] = tcpServerPort

		log.Println("Enriched JSON body with STREAM_HOST and STREAM_PORT")

		// invoke the action
		res, httpResp, err := client.Actions.Invoke(actionToInvoke, jsonBody, false, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			done()
			return
		}

		if httpResp.StatusCode != http.StatusAccepted {
			http.Error(w, "Error invoking action: "+httpResp.Status, http.StatusInternalServerError)
			done()
			return
		}

		if m, ok := res.(map[string]interface{}); ok {
			log.Println("Action invoked:", m["activationId"])
		} else {
			http.Error(w, "Unexpected reply from action invocation", http.StatusInternalServerError)
			done()
			return
		}

		// Flush the headers
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			done()
			return
		}

		log.Println("Setup streaming data to client")

		for {
			select {
			case data := <-streamDataChan:
				log.Println("Sending data to client:", string(data))
				_, err := w.Write([]byte("data: " + string(data) + "\n\n"))
				if err != nil {
					log.Println("Error writing to HTTP response:", err)
					done()
					return
				}
				flusher.Flush()
			case <-r.Context().Done():
				log.Println("HTTP Client closed connection")
				done()
				return
			}
		}
	}

}
