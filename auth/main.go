package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func main() {
	fmt.Println("auth service")
	tracerCloser := createTracer("auth")
	if tracerCloser != nil {
		defer tracerCloser.Close()
	}

	http.HandleFunc("/", doAuth)

	log.Println("serving on port 9190")
	err := http.ListenAndServe(":9190", nil)
	if err != nil {
		log.Fatal("error while serving: ", err)
	}
}

func doAuth(w http.ResponseWriter, r *http.Request) {
	var serverSpan opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		fmt.Printf("ERROR Accessing Span: %v\n", err)
	}

	// Create the span referring to the RPC client if available.
	// If wireContext == nil, a root span will be created.
	serverSpan = opentracing.StartSpan(
		"doAuth",
		ext.RPCServerOption(wireContext))

	defer serverSpan.Finish()

	time.Sleep(150 * time.Millisecond)
	w.Write([]byte("hello"))
}

func createTracer(serviceName string) io.Closer {
	// Sample configuration for testing. Use constant sampling to sample every trace
	// and enable LogSpan to log every span via configured Logger.
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	closer, err := cfg.InitGlobalTracer(
		serviceName,
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return nil
	}
	return closer
}
