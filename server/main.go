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

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func main() {
	fmt.Println("server")
	tracerCloser := createTracer("server")
	if tracerCloser != nil {
		defer tracerCloser.Close()
	}

	http.HandleFunc("/", doSomething)

	log.Println("serving on port 9090")
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("error while serving: ", err)
	}
}

func doSomething(w http.ResponseWriter, r *http.Request) {
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
		"doSomething",
		ext.RPCServerOption(wireContext))
	serverSpan.LogKV("foo", 123, "bar", 567)
	serverSpan.SetTag("mytag", "some value for the tag")

	defer serverSpan.Finish()

	httpReq, _ := http.NewRequest("GET", "http://127.0.0.1:9190/", nil)

	// Transmit the span's TraceContext as HTTP headers on our
	// outbound request.
	opentracing.GlobalTracer().Inject(
		serverSpan.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(httpReq.Header))

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		log.Fatal("GET error: ", err)
	}
	log.Printf("response from auth service: %d\n", resp.StatusCode)
	// simulate work after getting auth
	time.Sleep(50 * time.Millisecond)

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
