package tracing

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerprom "github.com/uber/jaeger-lib/metrics/prometheus"
)

func init() {
	viper.SetDefault("jaegertransport", "localhost:5775")
}

// Init returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func Init(service string) (opentracing.Tracer, io.Closer, error) {
	factory := jaegerprom.New()
	metrics := jaeger.NewMetrics(factory, map[string]string{"lib": "jaeger"})

	transport, err := jaeger.NewUDPTransport(viper.GetString("jaegertransport"), 0)
	if err != nil {
		log.Errorln(err.Error())
		return nil, nil, err
	}
	logAdapt := LogrusAdapter{}
	reporter := jaeger.NewCompositeReporter(
		jaeger.NewLoggingReporter(logAdapt),
		jaeger.NewRemoteReporter(transport,
			jaeger.ReporterOptions.Metrics(metrics),
			jaeger.ReporterOptions.Logger(logAdapt),
		),
	)

	sampler := jaeger.NewConstSampler(true)
	tracer, closer := jaeger.NewTracer(service,
		sampler,
		reporter,
		jaeger.TracerOptions.Metrics(metrics),
	)

	opentracing.SetGlobalTracer(tracer)

	return tracer, closer, nil
}

//HTTPMiddleware for check context from request header or span a new one
func HTTPMiddleware(h http.Handler, s string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		wireCtx, err := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))

		var serverSpan opentracing.Span
		if err == nil {
			serverSpan = tracer.StartSpan(fmt.Sprintf("%s-%s", s, r.URL.Path),
				ext.RPCServerOption(wireCtx))
		} else {
			serverSpan = tracer.StartSpan(fmt.Sprintf("%s-%s", s, r.URL.Path))
		}
		defer serverSpan.Finish()
		serverSpan.SetTag("http.method", r.Method)

		tracer.Inject(serverSpan.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))

		req, _ := httputil.DumpRequest(r, true)
		log.Infof("middleware - ot >> %s", string(req))

		h.ServeHTTP(w, r)
	})
}

//LogrusAdapter is an adapter for Jaeger
type LogrusAdapter struct{}

func (l LogrusAdapter) Error(msg string) {
	log.Errorf(msg)
}

//Infof for Jaeger
func (l LogrusAdapter) Infof(msg string, args ...interface{}) {
	log.Infof(msg, args)
}
