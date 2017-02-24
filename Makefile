all:
	go build -o ./bin/server ./main.go

pb:
	protoc -I./ -I/usr/local/include -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:.  \
		./userapi/api.proto
	protoc -I./ -I/usr/local/include -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		   --grpc-gateway_out=logtostderr=true:. \
		    ./userapi/api.proto

cert:
	openssl genrsa -out server.key 2048	
	openssl req -new -x509 -key server.key -out server.pem -days 3650

.PHONY:
	pb cert

