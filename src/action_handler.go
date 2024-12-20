package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/apache/openserverless-streaming-proxy/tcp"
)

func handleActionStream(streamingProxyAddr string, apihost string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, done := context.WithCancel(context.Background())

		namespace, actionToInvoke := getNamespaceAndAction(r)

		log.Println(fmt.Sprintf("Received request to invoke action: %s (%s)", actionToInvoke, namespace))

		apiKey, err := extractAuthToken(r)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			done()
			return
		}

		// Create OpenWhisk client
		client := NewOpenWhiskClient(apihost, apiKey, namespace)

		// opens a socket for listening in a random port
		sock, err := tcp.SetupTcpServer(ctx, streamingProxyAddr)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			done()
			return
		}

		enrichedBody, err := injectHostPortInBody(r, sock.Host, sock.Port)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			done()
			return
		}

		// invoke the action
		res, httpResp, err := client.Actions.Invoke(actionToInvoke, enrichedBody, false, false)
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

		for {
			select {
			case data := <-sock.StreamDataChan:
				if string(data) == "EOF" {
					log.Println("EOF received, closing connection")
					done()
					return
				}
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
