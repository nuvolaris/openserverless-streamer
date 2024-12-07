package main

import "os"

func main() {
	owApihost := os.Getenv("OW_APIHOST")
	if owApihost == "" {
		panic("OW_APIHOST is not set")
	}

	streamerAddr := os.Getenv("STREAMER_ADDR")
	if streamerAddr == "" {
		panic("STREAMER_ADDR is not set")
	}

	startHTTPServer(streamerAddr, owApihost)
}
