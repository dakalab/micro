package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dakalab/micro"
	"github.com/dakalab/micro/example/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
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

	// add swagger definition endpoint
	route := micro.Route{
		Method:  "GET",
		Pattern: micro.PathPattern("hello.swagger.json"),
		Handler: func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			data, _ := ioutil.ReadFile("../proto/hello.swagger.json")
			w.Write(data)
		},
	}

	sf := func() {
		log.Println("Server shutting down")
	}

	// init redoc, enable api docs on http://localhost:8888
	redoc := &micro.RedocOpts{
		Up: true,
	}
	redoc.AddSpec("Greeter", "/hello.swagger.json")

	s := micro.NewService(
		micro.Debug(true),
		micro.RouteOpt(route),
		micro.ShutdownFunc(sf),
		micro.Redoc(redoc),
	)

	proto.RegisterGreeterServer(s.GRPCServer, &Greeter{})

	var httpPort, grpcPort uint16
	httpPort = 8888
	grpcPort = 9999
	if err := s.Start(httpPort, grpcPort, reverseProxyFunc); err != nil {
		log.Fatal(err)
	}
}
