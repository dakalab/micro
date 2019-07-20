package main

import (
	"context"
	"log"
	"net/http"

	"github.com/dakalab/micro"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

func main() {

	reverseProxyFunc := func(
		ctx context.Context,
		mux *runtime.ServeMux,
		grpcHostAndPort string,
		opts []grpc.DialOption,
	) error {
		return nil
	}

	// add the /test endpoint
	route := micro.Route{
		Method:  "GET",
		Pattern: micro.PathPattern("test"),
		Handler: func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			w.Write([]byte("Hello!"))
		},
	}

	sf := func() {
		log.Println("Server shutting down")
	}

	s := micro.NewService(
		micro.Debug(),
		micro.RouteOpt(route),
		micro.ShutdownFunc(sf),
	)

	var httpPort, grpcPort uint16
	httpPort = 8888
	grpcPort = 9999
	if err := s.Start(httpPort, grpcPort, reverseProxyFunc); err != nil {
		log.Fatal(err)
	}
}
