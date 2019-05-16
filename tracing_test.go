package micro

import (
	"context"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func TestInitJaeger(t *testing.T) {
	closer, err := InitJaeger("", "localhost:6831", "localhost:6831")
	assert.Nil(t, closer)
	assert.Error(t, err)
}

func TestNewChildSpanFromContext(t *testing.T) {
	ctx := context.Background()
	span, ok := NewChildSpanFromContext(ctx, "child")
	assert.Nil(t, span)
	assert.False(t, ok)

	rootSpan := opentracing.StartSpan("root")
	ctx = context.WithValue(ctx, SpanKey, rootSpan)
	span, ok = NewChildSpanFromContext(ctx, "child")
	assert.NotNil(t, span)
	assert.True(t, ok)
}
