.PHONY: all
all: help

GOLANG_IMG=hyperjiang/golang:1.12.7

.PHONY: help
help:
	#########################################################################
	#
	# protoc      - compile *.proto to golang code
	# server      - start the grpc server
	# server-down - stop the grpc server
	# client      - run the grpc client
	#
	#########################################################################
	@echo "Enjoy!"

.PHONY: protoc
protoc:
	docker run --rm -t -v "${PWD}:/app" -w "/app" ${GOLANG_IMG} \
		bash -c 'make _protoc; gofmt -w proto'

.PHONY: _protoc
_protoc:
	find ./proto -name "*.proto" | xargs \
	/usr/bin/protoc \
		-I proto \
		-I $$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I $$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
		-I $$GOPATH/src \
		--gogo_out=plugins=grpc:./proto/ \
		--grpc-gateway_out=logtostderr=true:./proto/ \
		--swagger_out=./proto/

.PHONY: server
server: server-down
	docker run -d --name=micro-demo-server --rm \
	-p 8888:8888 -p 9999:9999 -p 18888:18888 -p 19999:19999 -p 28888:28888 -p 29999:29999 \
	-v "${PWD}:/app" -w "/app" ${GOLANG_IMG} go run server/main.go

.PHONY: server-down
server-down:
	docker rm -f micro-demo-server

.PHONY: client
client:
	docker exec -w "/app" micro-demo-server go run client/main.go
