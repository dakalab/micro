package micro

import (
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

// InitSpan - initiate the tracing span and set the http response header with X-Request-Id
func InitSpan(mux *runtime.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var serverSpan opentracing.Span
		// Extracting B3 tracing context from the request.
		// This step is important to extract the actual request context from outside of the applications.
		// By default, Jaeger use "uber-trace-id" to propagate tracing context,
		// and use prefix "uberctx-" to propagate baggage in http headers.
		// See https://github.com/jaegertracing/jaeger-client-go/blob/master/constants.go
		var wireContext, err = opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header),
		)
		// We will be using method name as the span name (method above)
		var methodName = r.Method + " " + r.URL.Path
		if err != nil {
			// Found no span in headers, start a new span as a parent span
			serverSpan = opentracing.StartSpan(methodName)
		} else {
			// Create span as a child of parent context
			serverSpan = opentracing.StartSpan(
				methodName,
				opentracing.ChildOf(wireContext),
			)
		}
		serverSpan.SetTag("http.url.host", r.URL.Hostname())
		serverSpan.SetTag("peer.address", r.RemoteAddr)
		serverSpan.SetTag("http.url", r.URL.RequestURI())
		serverSpan.SetTag("http.url.query", r.URL.RawQuery)

		var footprint string
		if footprint = serverSpan.BaggageItem("footprint"); footprint != "" {
			serverSpan.SetTag("footprint", footprint)
		} else {
			footprint = RequestID(r)
			serverSpan.SetBaggageItem("footprint", footprint)
			serverSpan.SetTag("footprint", footprint)
		}

		// Set the http response header with X-Request-Id
		w.Header().Set("X-Request-Id", footprint)

		// We are passing the span as an item in Go context
		var ctx = opentracing.ContextWithSpan(r.Context(), serverSpan)

		mux.ServeHTTP(w, r.WithContext(ctx))

		// Span needs to be finished in order to report it to Jaeger collector
		serverSpan.Finish()
	})
}

// InitJaeger - initiate an instance of Jaeger Tracer as global tracer
func InitJaeger(service, samplingServerURL, localAgentHost string) (io.Closer, error) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:              jaeger.SamplerTypeConst,
			Param:             1,
			SamplingServerURL: samplingServerURL,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: localAgentHost,
		},
	}

	return cfg.InitGlobalTracer(service, config.Logger(jaeger.StdLogger), config.ZipkinSharedRPCSpan(true))
}
