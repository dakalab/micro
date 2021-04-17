package micro

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// Route represents the route for mux
type Route struct {
	Method  string
	Path    string
	Handler runtime.HandlerFunc
}
