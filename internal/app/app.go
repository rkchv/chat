package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	auth_interceptors "github.com/rkchv/auth/pkg/user_v1/auth/grpc-interceptors"
	"github.com/rkchv/chat/lib/closer"
	"github.com/rkchv/chat/lib/logger"
	"github.com/rkchv/chat/lib/tracer"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	grpc_server "github.com/rkchv/chat/internal/grpc-server"
	"github.com/rkchv/chat/internal/grpc-server/interceptors"
	"github.com/rkchv/chat/pkg/chat_v1"
)

type App struct {
	grpc             *grpc.Server
	srvProvider      *serviceProvider
	traceExporter    *otlptrace.Exporter
	prometheusServer *http.Server
}

func NewApp(ctx context.Context) *App {
	app := &App{srvProvider: newServiceProvider()}
	app.init(ctx)
	app.initTracing(ctx, "chat-service")
	return app
}

func (a *App) init(ctx context.Context) {
	lg := logger.SetupLogger(logger.Env(a.srvProvider.Config().Env))
	a.grpc = grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.StreamInterceptor(interceptors.NewStreamAccessInterceptor([]string{
			chat_v1.ChatV1_Connect_FullMethodName,
		}, a.srvProvider.Config().SecretKey)),
		grpc.ChainUnaryInterceptor(
			interceptors.NewLoggerInterceptor(lg),
			auth_interceptors.NewAccessInterceptor([]string{
				chat_v1.ChatV1_Create_FullMethodName,
				chat_v1.ChatV1_SendMessage_FullMethodName,
				chat_v1.ChatV1_Delete_FullMethodName,
			}, a.srvProvider.Config().SecretKey),
		),
	)

	reflection.Register(a.grpc)
	chat_v1.RegisterChatV1Server(a.grpc, grpc_server.NewServer(
		a.srvProvider.ChatService(ctx),
		a.srvProvider.Config().ChatGarbageCycle,
		a.srvProvider.Config().ChatExpired))
}

func (a *App) initTracing(ctx context.Context, serviceName string) {
	//экспортер в jaeger
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(a.srvProvider.Config().Trace.ExporterGRPCAddress))
	if err != nil {
		log.Fatalf("failed to create trace exporter: %v", err)
	}
	a.traceExporter = exporter

	//собиратель трейсов
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(serviceName)),
	)
	if err != nil {
		log.Fatalf("failed to create trace provider: %v", err)
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithExportTimeout(time.Second*time.Duration(a.srvProvider.Config().Trace.BatchTimeout))),
		sdktrace.WithResource(r))

	//пробрасываем провайдер для исп. в других местах приложения
	tracer.Init(traceProvider.Tracer(serviceName))
	//регистрируем глобально
	prop := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(prop)
}

func (a *App) Start() error {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	conn, err := net.Listen("tcp", fmt.Sprintf("%s", a.srvProvider.Config().GRPC.Address()))
	if err != nil {
		return err
	}

	log.Printf("ChatAPI service started on %s\n", a.srvProvider.Config().GRPC.Address())

	closer.Add(func() error {
		a.grpc.GracefulStop()
		return nil
	})

	if err = a.grpc.Serve(conn); err != nil {
		return err
	}

	return nil
}

// StartPrometheusServer запускает сервер prometheus
func (a *App) StartPrometheusServer() error {
	log.Printf("Prometheus server started on %s\n", a.srvProvider.Config().Prometheus.Address())

	if a.prometheusServer == nil {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		a.prometheusServer = &http.Server{
			Addr:    a.srvProvider.Config().Prometheus.Address(),
			Handler: mux,
		}
	}

	if err := a.prometheusServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
