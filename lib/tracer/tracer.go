package tracer

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"github.com/rkchv/chat/lib/db"
)

var globalTracer trace.Tracer

type spanFlusher struct {
	span trace.Span
}

func (f *spanFlusher) Flush() {
	f.span.End()
}

func Init(t trace.Tracer) {
	globalTracer = t
}

func Tracer() trace.Tracer {
	return globalTracer
}
func Span(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return globalTracer.Start(ctx, spanName, opts...)
}

func SpanFlush(span trace.Span) db.LogFlush {
	return &spanFlusher{span: span}
}
