package micro

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

// Option - service functional option
//
// See this post about the "functional options" pattern:
// http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(s *Service)

// Debug - return an Option to set the service to debug mode
func Debug() Option {
	return func(s *Service) {
		s.debug = true
	}
}

// StaticDir - return an Option to set the staticDir
func StaticDir(staticDir string) Option {
	return func(s *Service) {
		s.staticDir = staticDir
	}
}

// Redoc - return an Option to set the Redoc
func Redoc(redoc *RedocOpts) Option {
	return func(s *Service) {
		s.redoc = redoc
	}
}

// Annotator - return an Option to append an annotator
func Annotator(annotator AnnotatorFunc) Option {
	return func(s *Service) {
		s.annotators = append(s.annotators, annotator)
	}
}

// ErrorHandler - return an Option to set the errorHandler
func ErrorHandler(errorHandler runtime.ProtoErrorHandlerFunc) Option {
	return func(s *Service) {
		s.errorHandler = errorHandler
	}
}

// HTTPHandler - return an Option to set the httpHandler
func HTTPHandler(httpHandler HTTPHandlerFunc) Option {
	return func(s *Service) {
		s.httpHandler = httpHandler
	}
}

// UnaryInterceptor - return an Option to append an unaryInterceptor
func UnaryInterceptor(unaryInterceptor grpc.UnaryServerInterceptor) Option {
	return func(s *Service) {
		s.unaryInterceptors = append(s.unaryInterceptors, unaryInterceptor)
	}
}

// StreamInterceptor - return an Option to append an streamInterceptor
func StreamInterceptor(streamInterceptor grpc.StreamServerInterceptor) Option {
	return func(s *Service) {
		s.streamInterceptors = append(s.streamInterceptors, streamInterceptor)
	}
}

// RouteOpt - return an Option to append a route
func RouteOpt(route Route) Option {
	return func(s *Service) {
		s.routes = append(s.routes, route)
	}
}

// ShutdownFunc - return an Option to register a function which will be called when server shutdown
func ShutdownFunc(f func()) Option {
	return func(s *Service) {
		s.shutdownFunc = f
	}
}

func (s *Service) apply(opts ...Option) {
	for _, opt := range opts {
		opt(s)
	}
}
