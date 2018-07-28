package opentracing

import (
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/sirupsen/logrus"
)

type (
	LoggerAdapter struct {
		logger *logrus.Logger
	}
)

func NewTracer(serviceName, backendHostPort string, metricsFactory metrics.Factory, logrus *logrus.Logger) opentracing.Tracer {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
	}

	jaegerLogger := LoggerAdapter{logger: logrus}

	var sender jaeger.Transport

	if strings.HasPrefix(backendHostPort, "http://") {
		sender = transport.NewHTTPTransport(
			backendHostPort,
			transport.HTTPBatchSize(1),
		)
	} else {
		if s, err := jaeger.NewUDPTransport(backendHostPort, 0); err != nil {
			logrus.Fatal(err, "cannot initialize UDP sender")
		} else {
			sender = s
		}
	}

	tracer, _, err := cfg.New(
		serviceName,
		config.Reporter(jaeger.NewRemoteReporter(
			sender,
			jaeger.ReporterOptions.BufferFlushInterval(1*time.Second),
			jaeger.ReporterOptions.Logger(jaegerLogger),
		)),

		config.Metrics(metricsFactory),
		config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)

	if err != nil {
		logrus.Fatal(err, "cannot initialize Jaeger Tracer")
	}

	return tracer
}

func (l LoggerAdapter) Error(msg string) {
	l.logger.Error(msg)
}

func (l LoggerAdapter) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}
