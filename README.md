# gRPC HTTP gateway with non-tls
According to gRPC proto we are able to generate a HTTP RESTful gateway：  
https://coreos.com/blog/grpc-protobufs-swagger.html

[grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) project mentioned at above article however has a big limitation that is only support HTTPS.
The reason is official net/http package only support tls+http2 but no support for non-tls http2. More detail can reference this [article](http://nullget.sourceforge.net/?q=node/885&lang=zh-hant)。

Actually, HTTPS is overhead in internal network. The mainstream idea is having nginx as the https door routing backend http service. So I deep 
into github to get whether someone has implement http gateway. However after 1 more days search, no one has done that...

So I implement by myself according to a [article](http://nullget.sourceforge.net/?q=node/885&lang=zh-hant).

First of all, write a .proto file:
```
syntax = "proto3";
package userapi;

import "google/api/annotations.proto";

service UserApi {
    rpc GetUser(GetUserRequest) returns (GetUserResponse) {
        option (google.api.http).get = "/v1/users/{id}";
    }
}

message User{
    int64 id = 1;
    string name = 2;
}

message GetUserRequest {
    int64 id = 1; 
}

message GetUserResponse {
    User user = 1;
}
```

Next, generate grpc server as well as gateway server:
```
protoc -I./ -I/usr/local/include -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:.  \
		./userapi/api.proto
protoc -I./ -I/usr/local/include -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
        --grpc-gateway_out=logtostderr=true:. \
		./userapi/api.proto
```

Third, most important to launch grpc server with http2 without tls:
```
func launchGrpcServer(addr string) {
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
			opts := &http2.ServeConnOpts{Handler: grpcServer}
			srv.ServeConn(conn, opts)
		}()
	}
}
```
Another simple way is launch grpc server with raw listener which has same effect with above one:
```
func launchGrpcServer(addr string) {
	grpcServer := grpc.NewServer()
	userapi.RegisterUserApiServer(grpcServer, &server.UserHandler{})

	lsner, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	grpcServer.Serve(lsner)
}
```
After grpc server launched, then launch grpc http gateway with http1.1 server:
```
func LaunchGrpcGateway(grpcAddr string, gatewayAddr string){
    gwmux := runtime.NewServeMux()
    err := userapi.RegisterUserApiHandlerFromEndpoint(context.Background(), gwmux, grpcAddr, []grpc.DialOption{grpc.WithInsecure()})
    if err != nil {
        panic(err)
    }
    fmt.Println("serving")
    http.ListenAndServe(gatewayAddr, gwmux) 
}
```

Now, both grpc and gateway server are ready, serving two different addresses respectively:
```
curl http://127.0.0.1:10001/v1/users/1
{"user":{"id":"1","name":"hangchen"}}
```

