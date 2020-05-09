// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"crypto/tls"
	"github.com/open-falcon/falcon-plus/modules/transfer/g"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

func StartRpc() {
	if !g.Config().Rpc.Enabled {
		return
	}

	addr := g.Config().Rpc.Listen
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalf("net.ResolveTCPAddr fail: %s", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("listen %s fail: %s", addr, err)
	} else {
		log.Println("rpc listening", addr)
	}

	server := rpc.NewServer()
	server.Register(new(Transfer))

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("listener.Accept occur error:", err)
			continue
		}

		conn.SetKeepAlive(true)
		go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}

func StartRpcWithTLS() {
	if !g.Config().RpcWithTLS.Enabled {
		return
	}

	addr := g.Config().RpcWithTLS.Listen
	_, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalf("net.ResolveTCPAddr fail: %s", err)
	}

	certFilePath := g.Config().RpcWithTLS.CrtFile
	keyFilePath := g.Config().RpcWithTLS.KeyFile

	cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
	if err != nil {
		log.Fatalf("load key pair fail: %s", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
	}

	listener, err := tls.Listen("tcp", addr,config)
	if err != nil {
		log.Fatalf("listen %s fail: %s", addr, err)
	} else {
		log.Println("rpc listening", addr)
	}

	server := rpc.NewServer()
	server.Register(new(Transfer))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("listener.Accept occur error:", err)
			continue
		}

		go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}