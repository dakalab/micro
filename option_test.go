package micro

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestStaticDir(t *testing.T) {
	s := NewService(StaticDir("/a/b/c"))
	assert.Equal(t, "/a/b/c", s.staticDir)
}

func TestAnnotator(t *testing.T) {
	s := NewService(
		Annotator(func(ctx context.Context, req *http.Request) metadata.MD {
			md := metadata.New(nil)
			md.Set("key", "value")
			return md
		}),
	)

	assert.Len(t, s.annotators, 2)
}

func TestErrorHandler(t *testing.T) {
	s := NewService(ErrorHandler(nil))
	assert.Nil(t, s.errorHandler)
}

func TestHTTPHandler(t *testing.T) {
	s := NewService(HTTPHandler(nil))
	assert.Nil(t, s.httpHandler)
}

func TestUnaryInterceptor(t *testing.T) {
	s := NewService(
		UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			return nil, nil
		}),
	)

	assert.Len(t, s.unaryInterceptors, 5)
}

func TestStreamInterceptor(t *testing.T) {
	s := NewService(
		StreamInterceptor(func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return nil
		}),
	)

	assert.Len(t, s.streamInterceptors, 5)
}
