package app

import (
	"context"
	"log"

	"github.com/rkchv/chat/lib/closer"
	"github.com/rkchv/chat/lib/db"
	"github.com/rkchv/chat/lib/db/pg"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/rkchv/chat/internal/config"
	grpc_client "github.com/rkchv/chat/internal/grpc-client"
	"github.com/rkchv/chat/internal/repository"
	"github.com/rkchv/chat/internal/repository/postgres"
	"github.com/rkchv/chat/internal/services"
)

type serviceProvider struct {
	conf           *config.Config
	chatService    *services.Service
	chatRepository repository.Repository
	dbc            db.Client
	authService    *grpc.ClientConn
}

func newServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

func (sp *serviceProvider) Config() config.Config {
	if sp.conf == nil {
		cfg := config.MustLoad()
		sp.conf = &cfg
	}

	return *sp.conf
}

func (sp *serviceProvider) DbClient(ctx context.Context) db.Client {
	if sp.dbc == nil {
		client, err := pg.NewClient(ctx, sp.Config().Postgres.DSN())
		if err != nil {
			log.Fatalf("failed to connect to pg: %v", err)
		}

		err = client.DB().Ping(ctx)
		if err != nil {
			log.Fatalf("failed ping to pg: %v", err)
		}

		sp.dbc = client
		closer.Add(sp.dbc.Close)
	}

	return sp.dbc
}

func (sp *serviceProvider) ChatRepository(ctx context.Context) repository.Repository {
	if sp.chatRepository == nil {
		sp.chatRepository = postgres.New(sp.DbClient(ctx))
	}

	return sp.chatRepository
}

func (sp *serviceProvider) ChatService(ctx context.Context) *services.Service {
	if sp.chatService == nil {
		sp.chatService = services.NewService(
			sp.ChatRepository(ctx),
			grpc_client.NewAuth(sp.AuthService(ctx)),
		)
	}

	return sp.chatService
}

func (sp *serviceProvider) AuthService(_ context.Context) *grpc.ClientConn {
	if sp.authService == nil {
		cl, err := grpc.NewClient(sp.Config().AuthServiceAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
		if err != nil {
			log.Fatalf("failed to connect to auth service: %v", err)
		}

		sp.authService = cl
	}

	return sp.authService
}
