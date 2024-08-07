package testutils

import (
	"context"
	"github.com/basemind-ai/monorepo/internal/utils/grpcutils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"testing"
)

func CreateTestGRPCServer[T any](
	t *testing.T,
	grpcRegistrar func(s grpc.ServiceRegistrar, srv T),
	service T,
	serverOpts ...grpc.ServerOption,
) *bufconn.Listener {
	t.Helper()
	listen := bufconn.Listen(101024 * 1024)
	server := grpcutils.CreateGRPCServer(
		grpcutils.Options{
			Environment: "test",
			ServiceName: "test-service",
			ServiceRegistrars: []grpcutils.ServiceRegistrar{
				func(s grpc.ServiceRegistrar) {
					grpcRegistrar(s, service)
				},
			},
		},
		serverOpts...,
	)

	go func() {
		if err := server.Serve(listen); err != nil {
			log.Fatalf("error serving test gRPC server: %v", err) //	skipcq: RVV-A0003
		}
	}()

	t.Cleanup(func() {
		if closeErr := listen.Close(); closeErr != nil {
			t.Fatalf("error closing test gRPC server: %v", closeErr)
		}
		server.Stop()
	})

	return listen
}

func CreateTestGRPCClient[T any](
	t *testing.T,
	listen *bufconn.Listener,
	clientFactory func(cc grpc.ClientConnInterface) T,
) T {
	t.Helper()
	conn, dialErr := grpc.DialContext(context.TODO(), "",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(
			func(context.Context, string) (net.Conn, error) {
				return listen.Dial()
			},
		),
	)

	if dialErr != nil {
		t.Fatalf("error dialing test gRPC server: %v", dialErr)
	}

	return clientFactory(conn)
}
