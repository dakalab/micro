package micro

import jaeger "github.com/uber/jaeger-client-go"

var logger jaeger.Logger

func init() {
	logger = jaeger.NullLogger
}
