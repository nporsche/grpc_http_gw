package main

import (
	"context"
	"net"
	"net/http"
	"sync"

	"golang.org/x/net/http2"

	"flag"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/nporsche/gapidemo/server"
	"github.com/nporsche/gapidemo/userapi"
	"google.golang.org/grpc"
)

var (
	grpcAddr        = flag.String("grpc-addr", "127.0.0.1:10000", "grpc serving address")
	grpcGatewayAddr = flag.String("grpc-gw-addr", "127.0.0.1:10001", "grpc gw serving address")
	wg              sync.WaitGroup
)

func main() {
	flag.Parse()
	wg.Add(2)
	go launchGrpcServer(*grpcAddr)
	go launchGrpcGatewayServer(*grpcAddr, *grpcGatewayAddr)
	wg.Wait()
}

func launchGrpcServer(addr string) {
	defer wg.Done()
	grpcServer := grpc.NewServer()
	userapi.RegisterUserApiServer(grpcServer, &server.UserHandler{})

	lsner, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	srv := &http2.Server{}
	for {
		conn, err := lsner.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			opts := &http2.ServeConnOpts{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					grpcServer.ServeHTTP(w, r)
				})}
			srv.ServeConn(conn, opts)
		}()
	}
}

func launchGrpcGatewayServer(grpcAddr string, gateway string) {
	defer wg.Done()
	gwmux := runtime.NewServeMux()
	err := userapi.RegisterUserApiHandlerFromEndpoint(context.Background(), gwmux, grpcAddr, []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		panic(err)
	}
	http.ListenAndServe(gateway, gwmux)
}
