package main

import (
	"context"
	"log"
	"net/http"

	"github.com/dakalab/micro"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"github.com/dakalab/micro/example/proto"
)

// Greeter - class to implement all gRPC endpoints of Greeter
type Greeter struct {
}

// SayHello implements gRPC endpoint "SayHello"
func (s *Greeter) SayHello(
	ctx context.Context,
	req *proto.HelloRequest,
) (*proto.HelloResponse, error) {
	return &proto.HelloResponse{
		Message: "Hello " + req.Name,
	}, nil
}

var _ proto.GreeterServer = (*Greeter)(nil) // make sure it implements the interface

func main() {

	reverseProxyFunc := func(
		ctx context.Context,
		mux *runtime.ServeMux,
		grpcHostAndPort string,
		opts []grpc.DialOption,
	) error {
		return proto.RegisterGreeterHandlerFromEndpoint(ctx, mux, grpcHostAndPort, opts)
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
		micro.Debug(true),
		micro.RouteOpt(route),
		micro.ShutdownFunc(sf),
	)

	proto.RegisterGreeterServer(s.GRPCServer, &Greeter{})

	var httpPort, grpcPort uint16
	httpPort = 8888
	grpcPort = 9999
	if err := s.Start(httpPort, grpcPort, reverseProxyFunc); err != nil {
		log.Fatal(err)
	}
}
