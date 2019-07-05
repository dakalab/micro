package micro

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	jaeger "github.com/uber/jaeger-client-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

// Service - to represent the microservice
type Service struct {
	GRPCServer         *grpc.Server
	HTTPServer         *http.Server
	httpHandler        HTTPHandlerFunc
	errorHandler       runtime.ProtoErrorHandlerFunc
	annotators         []AnnotatorFunc
	redoc              *RedocOpts
	staticDir          string
	mux                *runtime.ServeMux
	routes             []Route
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor
	debug              bool
}

// ReverseProxyFunc - a callback that the caller should implement to steps to reverse-proxy the HTTP/1 requests to gRPC
type ReverseProxyFunc func(ctx context.Context, mux *runtime.ServeMux, grpcHostAndPort string, opts []grpc.DialOption) error

// HTTPHandlerFunc - http middleware handler function
type HTTPHandlerFunc func(*runtime.ServeMux) http.Handler

// AnnotatorFunc - annotator function is for injecting meta data from http request into gRPC context
type AnnotatorFunc func(context.Context, *http.Request) metadata.MD

// DefaultHTTPHandler - default http handler which will initiate the tracing span and set the http response header with X-Request-Id
func DefaultHTTPHandler(mux *runtime.ServeMux) http.Handler {
	return InitSpan(mux)
}

// DefaultAnnotator - pass span info into gRPC context
func DefaultAnnotator(ctx context.Context, req *http.Request) metadata.MD {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	var footprint string
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		footprint = span.BaggageItem("footprint")
		md.Set(jaeger.TraceBaggageHeaderPrefix+"footprint", footprint)
		// IMPORTANT, otherwise the gRPC service will create a root span instead of a child span
		childSpan := opentracing.StartSpan(
			"child", // this name will be replaced with actual rpc name in the open tracing interceptor
			opentracing.ChildOf(span.Context()),
		)
		md.Set(jaeger.TraceContextHeaderName, fmt.Sprintf("%+v", childSpan))
	}
	if footprint == "" {
		footprint = RequestID(req)
	}

	md.Set("x-request-id", footprint)

	return md
}

// RequestID - get X-Request-Id from http request header, if it does not exist then generate one
func RequestID(req *http.Request) string {
	id := req.Header.Get("X-Request-Id")

	if id == "" {
		id = uuid.New().String()
	}

	// set it back into request header
	req.Header.Set("X-Request-Id", id)

	return id
}

func defaultService() *Service {
	s := Service{}
	s.annotators = append(s.annotators, DefaultAnnotator)
	s.errorHandler = runtime.DefaultHTTPError
	s.httpHandler = DefaultHTTPHandler

	s.streamInterceptors = []grpc.StreamServerInterceptor{}
	s.unaryInterceptors = []grpc.UnaryServerInterceptor{}

	// install prometheus interceptor
	s.streamInterceptors = append(s.streamInterceptors, grpc_prometheus.StreamServerInterceptor)
	s.unaryInterceptors = append(s.unaryInterceptors, grpc_prometheus.UnaryServerInterceptor)

	// install validator interceptor
	s.streamInterceptors = append(s.streamInterceptors, grpc_validator.StreamServerInterceptor())
	s.unaryInterceptors = append(s.unaryInterceptors, grpc_validator.UnaryServerInterceptor())

	// install panic handler
	s.streamInterceptors = append(s.streamInterceptors, StreamPanicHandler)
	s.unaryInterceptors = append(s.unaryInterceptors, UnaryPanicHandler)

	// add /metrics HTTP/1 endpoint
	routeMetrics := Route{
		Method:  "GET",
		Pattern: PathPattern("metrics"),
		Handler: func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			promhttp.Handler().ServeHTTP(w, r)
		},
	}
	s.routes = append(s.routes, routeMetrics)

	return &s
}

// NewService - create a new microservice
func NewService(opts ...Option) *Service {
	s := defaultService()

	s.apply(opts...)

	// default tracer is NoopTracer, you need to use an acutal tracer for tracing
	tracer := opentracing.GlobalTracer()

	// install open tracing interceptor
	if s.debug {
		SetLogger(jaeger.StdLogger)
		s.streamInterceptors = append(s.streamInterceptors, otgrpc.OpenTracingStreamServerInterceptor(tracer, otgrpc.LogPayloads()))
		s.unaryInterceptors = append(s.unaryInterceptors, otgrpc.OpenTracingServerInterceptor(tracer, otgrpc.LogPayloads()))
	} else {
		s.streamInterceptors = append(s.streamInterceptors, otgrpc.OpenTracingStreamServerInterceptor(tracer))
		s.unaryInterceptors = append(s.unaryInterceptors, otgrpc.OpenTracingServerInterceptor(tracer))
	}

	s.GRPCServer = grpc.NewServer(
		grpc_middleware.WithStreamServerChain(s.streamInterceptors...),
		grpc_middleware.WithUnaryServerChain(s.unaryInterceptors...),
	)

	return s
}

// Start - start the microservice with listening on the ports
func (s *Service) Start(httpPort uint16, grpcPort uint16, reverseProxyFunc ReverseProxyFunc) error {

	errChan := make(chan error, 1)

	// start HTTP/1.0 gateway server
	go func() {
		errChan <- s.startGRPCGateway(httpPort, grpcPort, reverseProxyFunc)
	}()

	// start gRPC server
	go func() {
		errChan <- s.startGRPCServer(grpcPort)
	}()

	return <-errChan
}

func (s *Service) startGRPCServer(grpcPort uint16) error {
	// setup /metrics for prometheus
	grpc_prometheus.Register(s.GRPCServer)

	// register reflection service on gRPC server.
	reflection.Register(s.GRPCServer)

	grpcHost := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", grpcHost)
	if err != nil {
		return err
	}

	return s.GRPCServer.Serve(lis)
}

func (s *Service) startGRPCGateway(httpPort uint16, grpcPort uint16, reverseProxyFunc ReverseProxyFunc) error {
	var muxOptions []runtime.ServeMuxOption
	muxOptions = append(muxOptions, runtime.WithMarshalerOption(
		runtime.MIMEWildcard,
		&runtime.JSONPb{OrigName: true, EmitDefaults: true},
	))
	muxOptions = append(muxOptions, runtime.WithProtoErrorHandler(s.errorHandler))

	for _, annotator := range s.annotators {
		muxOptions = append(muxOptions, runtime.WithMetadata(annotator))
	}

	s.mux = runtime.NewServeMux(muxOptions...)

	if s.redoc.Up {
		// add /docs HTTP/1 endpoint
		routeDocs := Route{
			Method:  "GET",
			Pattern: PathPattern(s.redoc.Route),
			Handler: func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
				s.redoc.Serve(w, r, pathParams)
			},
		}
		s.routes = append(s.routes, routeDocs)
	}

	// apply routes
	for _, route := range s.routes {
		s.mux.Handle(route.Method, route.Pattern, route.Handler)
	}

	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := reverseProxyFunc(context.Background(), s.mux, fmt.Sprintf("localhost:%d", grpcPort), opts)
	if err != nil {
		return err
	}

	// this is the fallback handler that will serve static files,
	// if file does not exist, then a 404 error will be returned.
	s.mux.Handle("GET", AllPattern(), func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		dir := s.staticDir
		if s.staticDir == "" {
			dir, _ = os.Getwd()
		}

		// check if the file exists and fobid showing directory
		path := filepath.Join(dir, r.URL.Path)
		if fileInfo, err := os.Stat(path); os.IsNotExist(err) || fileInfo.IsDir() {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, path)
	})

	s.HTTPServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: handlers.RecoveryHandler()(s.httpHandler(s.mux)),
	}

	return s.HTTPServer.ListenAndServe()
}

// Stop - stop the microservice
func (s *Service) Stop() {
	s.GRPCServer.Stop()
	s.HTTPServer.Shutdown(context.Background())
}
