package main

import (
	"context"
	"fmt"
	"github.com/basemind-ai/monorepo/gen/proto/v1"
	"github.com/basemind-ai/monorepo/internal/utils/config"
	grpcutils2 "github.com/basemind-ai/monorepo/internal/utils/grpcutils"
	"github.com/basemind-ai/monorepo/internal/utils/logging"
	"github.com/basemind-ai/monorepo/internal/utils/rediscache"
	"github.com/basemind-ai/monorepo/services/api-gateway/internal/connectors"
	"github.com/basemind-ai/monorepo/services/api-gateway/internal/services"
	"github.com/basemind-ai/monorepo/shared/go/db"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()

	cfg := config.Get(ctx)

	logging.Configure(cfg.Environment != "production")

	connectors.Init(ctx)

	rediscache.New(cfg.RedisURL)

	conn, connErr := db.CreateConnection(ctx, cfg.DatabaseURL)
	if connErr != nil {
		log.Fatal().Err(connErr).Msg("failed to connect to DB")
	}

	defer conn.Close()

	server := grpcutils2.CreateGRPCServer(
		grpcutils2.Options{
			AuthHandler: grpcutils2.NewAuthHandler(cfg.JWTSecret).HandleAuth,
			Environment: cfg.Environment,
			ServiceName: "api-gateway",
			ServiceRegistrars: []grpcutils2.ServiceRegistrar{
				func(s grpc.ServiceRegistrar) {
					gateway.gateway.RegisterAPIGatewayServiceServer(s, services.APIGatewayServer{})
				},
				func(s grpc.ServiceRegistrar) {
					ptesting.RegisterPromptTestingServiceServer(s, services.PromptTestingServer{})
				},
			},
		},
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		address := fmt.Sprintf("0.0.0.0:%d", cfg.ServerPort)

		listen, listenErr := net.Listen("tcp", address)
		if listenErr != nil {
			return listenErr
		}

		log.Info().Str("service", "api-gateway").Str("address", address).Msg("server starting")
		return server.Serve(listen)
	})

	g.Go(func() error {
		<-gCtx.Done()
		server.Stop()
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Info().Msg(err.Error())
	}
}
