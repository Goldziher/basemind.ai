package main

import (
	"context"
	"fmt"
	"github.com/basemind-ai/gateway/gen/gateway/v2"
	"github.com/basemind-ai/gateway/internal/server"
	"github.com/basemind-ai/gateway/internal/utils/config"
	"github.com/basemind-ai/gateway/internal/utils/grpcutils"
	"github.com/basemind-ai/gateway/internal/utils/logging"
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

	srv := grpcutils.CreateGRPCServer(
		grpcutils.Options{
			AuthHandler: grpcutils.NewAuthHandler(cfg.JWTSecret).HandleAuth,
			Environment: cfg.Environment,
			ServiceName: "api-gateway",
			ServiceRegistrars: []grpcutils.ServiceRegistrar{
				func(s grpc.ServiceRegistrar) {
					gateway.RegisterAPIGatewayServiceServer(s, server.APIGatewayServer{})
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
		return srv.Serve(listen)
	})

	g.Go(func() error {
		<-gCtx.Done()
		srv.Stop()
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Info().Msg(err.Error())
	}
}
