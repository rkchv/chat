package interceptors

import (
	"context"

	"github.com/rkchv/chat/lib/logger"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
)

var lg *slog.Logger

// NewLoggerInterceptor навешивает логгер на контекст запросов
func NewLoggerInterceptor(l *slog.Logger) grpc.UnaryServerInterceptor {
	lg = l
	return loggerInterceptor
}

func loggerInterceptor(ctx context.Context, req interface{}, i *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log := lg

	traceId := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	log = lg.With(slog.String("trace_id", traceId))

	ctx = logger.AssignLogger(ctx, log)
	log.Debug("called", slog.String("method", i.FullMethod))

	return handler(ctx, req)
}
