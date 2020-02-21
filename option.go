package micro

import (
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

// Option is service functional option
//
// See this post about the "functional options" pattern:
// http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(s *Service)

// StaticDir returns an Option to set the staticDir
func StaticDir(staticDir string) Option {
	return func(s *Service) {
		s.staticDir = staticDir
	}
}

// Redoc returns an Option to set the Redoc
func Redoc(redoc *RedocOpts) Option {
	return func(s *Service) {
		s.redoc = redoc
	}
}

// Annotator returns an Option to append an annotator
func Annotator(annotator AnnotatorFunc) Option {
	return func(s *Service) {
		s.annotators = append(s.annotators, annotator)
	}
}

// ErrorHandler returns an Option to set the errorHandler
func ErrorHandler(errorHandler runtime.ProtoErrorHandlerFunc) Option {
	return func(s *Service) {
		s.errorHandler = errorHandler
	}
}

// HTTPHandler returns an Option to set the httpHandler
func HTTPHandler(httpHandler HTTPHandlerFunc) Option {
	return func(s *Service) {
		s.httpHandler = httpHandler
	}
}

// UnaryInterceptor returns an Option to append an unaryInterceptor
func UnaryInterceptor(unaryInterceptor grpc.UnaryServerInterceptor) Option {
	return func(s *Service) {
		s.unaryInterceptors = append(s.unaryInterceptors, unaryInterceptor)
	}
}

// StreamInterceptor returns an Option to append an streamInterceptor
func StreamInterceptor(streamInterceptor grpc.StreamServerInterceptor) Option {
	return func(s *Service) {
		s.streamInterceptors = append(s.streamInterceptors, streamInterceptor)
	}
}

// RouteOpt returns an Option to append a route
func RouteOpt(route Route) Option {
	return func(s *Service) {
		s.routes = append(s.routes, route)
	}
}

// ShutdownFunc returns an Option to register a function which will be called when server shutdown
func ShutdownFunc(f func()) Option {
	return func(s *Service) {
		s.shutdownFunc = f
	}
}

// ShutdownTimeout returns an Option to set the timeout before the server shutdown abruptly
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Service) {
		s.shutdownTimeout = timeout
	}
}

// PreShutdownDelay returns an Option to set the time waiting for running goroutines
// to finish their jobs before the shutdown starts
func PreShutdownDelay(timeout time.Duration) Option {
	return func(s *Service) {
		s.preShutdownDelay = timeout
	}
}

// InterruptSignal returns an Option to append a interrupt signal
func InterruptSignal(signal os.Signal) Option {
	return func(s *Service) {
		s.interruptSignals = append(s.interruptSignals, signal)
	}
}

// GRPCServerOption returns an Option to append a gRPC server option
func GRPCServerOption(serverOption grpc.ServerOption) Option {
	return func(s *Service) {
		s.grpcServerOptions = append(s.grpcServerOptions, serverOption)
	}
}

// GRPCDialOption returns an Option to append a gRPC dial option
func GRPCDialOption(dialOption grpc.DialOption) Option {
	return func(s *Service) {
		s.grpcDialOptions = append(s.grpcDialOptions, dialOption)
	}
}

// MuxOption returns an Option to append a mux option
func MuxOption(muxOption runtime.ServeMuxOption) Option {
	return func(s *Service) {
		s.muxOptions = append(s.muxOptions, muxOption)
	}
}

// WithHTTPServer returns an Option to set the http server, note that the Addr and Handler will be
// reset in startGRPCGateway(), so you are not able to specify them
func WithHTTPServer(server *http.Server) Option {
	return func(s *Service) {
		s.HTTPServer = server
	}
}

// WithLogger uses the provided logger
func WithLogger(logger Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

func (s *Service) apply(opts ...Option) {
	for _, opt := range opts {
		opt(s)
	}
}
