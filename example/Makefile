.PHONY: all
all: help

GOLANG_IMG=hyperjiang/golang:1.16.3

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
		bash -c 'buf beta mod update; buf generate'

.PHONY: logs
logs:
	docker logs -f --tail=30 micro-demo-server

.PHONY: server
server: server-down
	docker run -d --name=micro-demo-server --rm \
	-e GOPROXY=https://goproxy.cn \
	-p 8888:8888 -p 9999:9999 -p 18888:18888 -p 19999:19999 -p 28888:28888 -p 29999:29999 \
	-v "${PWD}:/app" -w "/app" ${GOLANG_IMG} go run server/main.go
	make logs

.PHONY: server-down
server-down:
	docker rm -f micro-demo-server || true

.PHONY: client
client:
	docker exec -e GOPROXY=https://goproxy.cn -e GODEBUG=x509ignoreCN=0 -w "/app" micro-demo-server go run client/main.go
