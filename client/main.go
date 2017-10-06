package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

func main() {
	fmt.Println("client")
	tracerCloser := createTracer("client")
	if tracerCloser != nil {
		defer tracerCloser.Close()
	}

	span := opentracing.StartSpan("get data")
	defer span.Finish()

	// ctx := context.Background()
	// span := opentracing.SpanFromContext(ctx)
	if span != nil {
		time.Sleep(100 * time.Millisecond)

		httpReq, _ := http.NewRequest("GET", "http://127.0.0.1:9090/", nil)

		// Transmit the span's TraceContext as HTTP headers on our
		// outbound request.
		opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(httpReq.Header))

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			log.Fatal("GET error: ", err)
		}
		time.Sleep(100 * time.Millisecond)

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("RESPONSE: %d %s\n", resp.StatusCode, body)
	}
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
