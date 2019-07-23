.PHONY: all
all: protoc

GOLANG_IMG=hyperjiang/golang:1.12.6

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
