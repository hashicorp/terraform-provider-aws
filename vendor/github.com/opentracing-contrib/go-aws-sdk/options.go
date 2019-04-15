package otaws

import "github.com/opentracing/opentracing-go"

type Option func(*config)

func WithTracer(tracer opentracing.Tracer) Option {
	return func(c *config) {
		c.tracer = tracer
	}
}

type config struct {
	tracer opentracing.Tracer
}

func defaultConfig() *config {
	return &config{
		tracer: opentracing.GlobalTracer(),
	}
}
