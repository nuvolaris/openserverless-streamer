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

package tcp

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type SocketsServer struct {
	ctx            context.Context
	listener       net.Listener
	wg             sync.WaitGroup
	Host           string
	Port           string
	StreamDataChan chan []byte
}

func SetupTcpServer(ctx context.Context, streamingProxyAddr string) (*SocketsServer, error) {
	socketServer, err := startTCPServer(ctx, streamingProxyAddr)
	if err != nil {
		return nil, err
	}
	go socketServer.WaitToCleanUp()

	tcpServerHost, tcpServerPort, err := net.SplitHostPort(socketServer.listener.Addr().String())
	if err != nil {
		return nil, err
	}

	socketServer.Host = tcpServerHost
	socketServer.Port = tcpServerPort

	return socketServer, nil
}

func startTCPServer(ctx context.Context, streamingProxyAddr string) (*SocketsServer, error) {
	listener, err := net.Listen("tcp", streamingProxyAddr+":0")
	if err != nil {
		return nil, errors.New("Error starting TCP server")
	}

	s := &SocketsServer{
		ctx:            ctx,
		listener:       listener,
		StreamDataChan: make(chan []byte),
	}

	s.wg.Add(1)
	go s.acceptConnections()

	log.Println("New TCP server listening on:", s.listener.Addr().String())
	return s, nil
}

func (s *SocketsServer) acceptConnections() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				log.Println("accept error", err.Error())
			}
		} else {
			s.wg.Add(1)
			go func() {
				s.handleConnection(conn)
				s.wg.Done()
			}()
		}
	}

}

func (s *SocketsServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Println("New TCP connection accepted!")
	buf := make([]byte, 2048)

ReadLoop:
	for {
		select {

		case <-s.ctx.Done():
			return

		default:
			conn.SetDeadline(time.Now().Add(100 * time.Millisecond))

			for {
				n, err := conn.Read(buf)

				if err != nil {
					if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
						continue ReadLoop
					} else if err != io.EOF {
						log.Println("Error reading from TCP connection", err)
						return
					}
				}

				if n == 0 {
					continue ReadLoop
				}

				s.StreamDataChan <- buf[:n]
			}

		}
	}
}

func (s *SocketsServer) WaitToCleanUp() {
	<-s.ctx.Done()
	log.Println("Stopping listening on", s.listener.Addr().String())
	s.listener.Close()
	s.wg.Wait()
	log.Print("TCP server closed\n\n")
}
