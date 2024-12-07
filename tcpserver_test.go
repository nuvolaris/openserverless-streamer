package main

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestHandleConnection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	streamDataChan := make(chan []byte, 10)
	server := &SocketsServer{
		ctx:            ctx,
		streamDataChan: streamDataChan,
	}

	// Create a pipe to simulate a network connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.handleConnection(serverConn)
	}()

	// Write data to the client connection
	testData := []byte("test data")
	clientConn.Write(testData)

	// Read data from the streamDataChan
	select {
	case receivedData := <-streamDataChan:
		if string(receivedData) != string(testData) {
			t.Errorf("Expected %s, but got %s", string(testData), string(receivedData))
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for data")
	}

	// Cancel the context to stop the server
	cancel()
	wg.Wait()
}

func TestHandleConnectionContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	streamDataChan := make(chan []byte, 10)
	server := &SocketsServer{
		ctx:            ctx,
		streamDataChan: streamDataChan,
	}

	// Create a pipe to simulate a network connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.handleConnection(serverConn)
	}()

	// Cancel the context to stop the server
	cancel()
	wg.Wait()

	// Write data to the client connection after context is cancelled
	testData := []byte("test data")
	clientConn.Write(testData)

	// Ensure no data is read from the streamDataChan
	select {
	case <-streamDataChan:
		t.Error("Expected no data, but got some")
	case <-time.After(100 * time.Millisecond):
		// Expected case
	}
}
