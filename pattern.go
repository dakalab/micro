package micro

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
)

// PathPattern - return a pattern which matches exactly with the path
func PathPattern(path string) runtime.Pattern {
	return runtime.MustPattern(runtime.NewPattern(1, []int{int(utilities.OpLitPush), 0}, []string{path}, ""))
}

// AllPattern - return a pattern which matches any url
func AllPattern() runtime.Pattern {
	return runtime.MustPattern(runtime.NewPattern(1, []int{int(utilities.OpPush), 0}, []string{""}, ""))
}
