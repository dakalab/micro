package micro

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var reverseProxyFunc ReverseProxyFunc
var httpPort, grpcPort uint
var shutdownFunc func()

func init() {
	reverseProxyFunc = func(
		ctx context.Context,
		mux *runtime.ServeMux,
		grpcHostAndPort string,
		opts []grpc.DialOption,
	) error {
		return nil
	}

	httpPort = 8888
	grpcPort = 9999

	shutdownFunc = func() {
		fmt.Println("Server shutting down")
	}
}

func TestNewService(t *testing.T) {
	var should = require.New(t)

	redoc := &RedocOpts{
		Route: "/docs",
		Up:    true,
	}
	redoc.AddSpec("PetStore", "https://rebilly.github.io/ReDoc/swagger.yaml")
	redoc.AddSpec("Service", "/demo.swagger.json")
	redoc.AddSpec("Service2", "/demo.swagger.json") // it will not add a duplicate route

	// add the /test endpoint
	route := Route{
		Method: "GET",
		Path:   "/test",
		Handler: func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			w.Write([]byte("Hello!"))
		},
	}

	s := NewService(
		Redoc(redoc),
		RouteOpt(route),
		ShutdownFunc(shutdownFunc),
		PreShutdownDelay(0),
		WithLogger(LoggerFunc(log.Printf)),
	)

	// add the /health endpoint
	healthRoute := Route{
		Method: "GET",
		Path:   "/health",
		Handler: func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		},
	}
	s.AddRoutes(healthRoute)

	noRoute := Route{
		Method:  "GET",
		Path:    "/404",
		Handler: s.ServeFile,
	}
	s.AddRoutes(noRoute)

	go func() {
		err := s.Start(httpPort, grpcPort, reverseProxyFunc)
		should.NoError(err)
	}()

	// wait 1 second for the server start
	time.Sleep(1 * time.Second)

	// check if the http server is up
	httpHost := fmt.Sprintf(":%d", httpPort)
	_, err := net.Listen("tcp", httpHost)
	should.Error(err)

	// check if the grpc server is up
	grpcHost := fmt.Sprintf(":%d", grpcPort)
	_, err = net.Listen("tcp", grpcHost)
	should.Error(err)

	// check if the http endpoint works
	client := &http.Client{}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/", httpPort))
	should.NoError(err)
	should.Equal(http.StatusNotFound, resp.StatusCode)

	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", httpPort))
	should.NoError(err)
	should.Equal(http.StatusOK, resp.StatusCode)
	b, err := ioutil.ReadAll(resp.Body)
	should.NoError(err)
	should.Equal("OK", string(b))

	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/demo.swagger.json", httpPort))
	should.NoError(err)
	should.Equal(http.StatusOK, resp.StatusCode)

	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/404", httpPort))
	should.NoError(err)
	should.Equal(http.StatusNotFound, resp.StatusCode)

	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/docs", httpPort))
	if err != nil {
		t.Error(err)
	}
	should.Equal(http.StatusOK, resp.StatusCode)

	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", httpPort))
	should.NoError(err)
	should.Equal(http.StatusOK, resp.StatusCode)

	// create service s2 to trigger errChan1
	s2 := NewService(
		Redoc(&RedocOpts{
			Up: false,
		}),
	)

	// grpc port 9999 alreday in use
	err = s2.Start(httpPort, grpcPort, reverseProxyFunc)
	should.Error(err)

	// create service s3 to trigger errChan2
	s3 := NewService(
		Redoc(&RedocOpts{
			Up: false,
		}),
	)

	// http port 8888 already in use
	s.GRPCServer.Stop()
	err = s3.Start(httpPort, grpcPort, reverseProxyFunc)
	should.Error(err)

	// wait 1 second for s3 gRPC server start
	time.Sleep(1 * time.Second)

	// close all previous services
	s.HTTPServer.Close()
	s3.GRPCServer.Stop()

	// run a new service, we use different ports to make sure ci not complain
	httpPort = 18888
	grpcPort = 19999
	s4 := NewService(
		Redoc(&RedocOpts{
			Up: false,
		}),
		ShutdownTimeout(10*time.Second),
	)
	go func() {
		err := s4.Start(httpPort, grpcPort, reverseProxyFunc)
		should.NoError(err)
	}()

	// wait 1 second for the server start
	time.Sleep(1 * time.Second)

	// the redoc is not up for the second server
	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/docs", httpPort))
	should.NoError(err)
	should.Equal(http.StatusNotFound, resp.StatusCode)

	// send an interrupt signal to stop s4
	syscall.Kill(s4.Getpid(), syscall.SIGINT)

	// wait 3 second for the server shutdown
	time.Sleep(3 * time.Second)
}

func TestErrorReverseProxyFunc(t *testing.T) {
	var should = require.New(t)

	s := NewService(
		Redoc(&RedocOpts{
			Up: true,
		}),
	)

	// mock error from reverseProxyFunc
	errText := "reverse proxy func error"
	reverseProxyFunc = func(
		ctx context.Context,
		mux *runtime.ServeMux,
		grpcHostAndPort string,
		opts []grpc.DialOption,
	) error {
		return errors.New(errText)
	}

	err := s.startGRPCGateway(httpPort, grpcPort, reverseProxyFunc)
	should.EqualError(err, errText)
}
